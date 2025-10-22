package shipment

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/hoenirvili/axiogate/http/api"
	"github.com/hoenirvili/axiogate/log"
)

// Payloader defines custom payload for a provider.
type Payloader interface {
	Payload(req *api.ShippingRequest) []byte
	To() string
}

type Storage interface {
	Save(ctx context.Context, provider string, payload []byte) error
}

type Client interface {
	Do(ctx context.Context, to string, payload any) ([]byte, error)
}

type Shipment struct {
	providers map[string]Payloader
	log       *slog.Logger
	storage   Storage
	client    Client
}

type Option func(p *Shipment)

func WithLogger(log *slog.Logger) Option {
	return func(s *Shipment) {
		s.log = log
	}
}

// New return a new shipment service that handlers the
// multi provider fan out shipment.
func New(cli Client, providers map[string]Payloader, st Storage, options ...Option) *Shipment {
	s := &Shipment{
		client:    cli,
		providers: providers,
		log:       log.Noop(),
		storage:   st,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// ErrProviderUnsupported error returned when the caller makes a shipment request to an unknown provider.
type ErrProviderUnsupported struct {
	Provider string
}

var _ error = (*ErrProviderUnsupported)(nil)

func (e *ErrProviderUnsupported) Error() string {
	return fmt.Sprintf("unsupported provider %s", e.Provider)
}

func (s *Shipment) allJobs() []job {
	jobs := make([]job, 0, len(s.providers))
	for provider, payloader := range s.providers {
		jobs = append(jobs, job{provider: provider, payloader: payloader})
	}
	return jobs
}

type job struct {
	payloader Payloader
	provider  string
}

func (s *Shipment) jobs(providers []string) ([]job, error) {
	if len(providers) == 0 {
		return s.allJobs(), nil
	}
	jobs := make([]job, 0, len(s.providers))
	for _, provider := range providers {
		val, ok := s.providers[provider]
		if !ok {
			return nil, &ErrProviderUnsupported{Provider: provider}
		}
		jobs = append(jobs, job{provider: provider, payloader: val})
	}
	return jobs, nil
}

func (s *Shipment) Send(ctx context.Context, providers []string, req *api.ShippingRequest) ([]api.ShippingResponse, error) {
	jobs, err := s.jobs(providers)
	if err != nil {
		return nil, err
	}
	return s.send(ctx, jobs, req)
}

func (s *Shipment) send(ctx context.Context, jobs []job, req *api.ShippingRequest) ([]api.ShippingResponse, error) {
	res := make(chan api.ShippingResponse, len(jobs))
	wg := new(sync.WaitGroup)
	wg.Add(len(jobs))
	for _, j := range jobs {
		go func(job job) {
			defer wg.Done()
			if job.payloader == nil {
				return
			}
			payload := job.payloader.Payload(req)
			s.log.With(slog.String("payload", string(payload))).
				Info("Sending resulting payload")
			payload, err := s.client.Do(ctx, job.payloader.To(), payload)
			if err != nil {
				res <- api.ShippingResponse{
					Endpoint:    job.payloader.To(),
					RawResponse: payload,
					Error:       err.Error(),
				}
				return
			}
			if err := s.storage.Save(ctx, job.provider, payload); err != nil {
				res <- api.ShippingResponse{
					Endpoint:    job.payloader.To(),
					RawResponse: payload,
					Error:       err.Error(),
				}
				return
			}
			res <- api.ShippingResponse{
				Endpoint:    job.payloader.To(),
				RawResponse: payload,
			}
		}(j)
	}
	wg.Wait()
	close(res)
	responses := []api.ShippingResponse{}
	for resp := range res {
		responses = append(responses, resp)
	}
	return responses, nil
}
