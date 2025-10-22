package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/hoenirvili/axiogate/http/api"
	"github.com/hoenirvili/axiogate/http/response"
	"github.com/hoenirvili/axiogate/log"
	"github.com/hoenirvili/axiogate/shipment"
)

type Shipment struct {
	sender Sender
	log    *slog.Logger
}

type Option func(s *Shipment)

func WithLogger(log *slog.Logger) Option {
	return func(s *Shipment) {
		s.log = log.WithGroup("shipment")
	}
}

// Sender defines how we send the request down the chain.
type Sender interface {
	// Send sends the req to all providers in the providers list.
	// If providers slice is empty then we sent to all internal providers.
	Send(ctx context.Context, providers []string, req *api.ShippingRequest) ([]api.ShippingResponse, error)
}

// NewShipment create a new handler shipment instance to shipment http requests.
func NewShipment(sender Sender, options ...Option) *Shipment {
	ship := &Shipment{
		sender: sender,
		log:    log.Noop(),
	}
	for _, option := range options {
		option(ship)
	}
	return ship
}

func (s *Shipment) providerList(r *http.Request) []string {
	values := r.URL.Query()
	if values == nil {
		return nil
	}
	providers, ok := values["providers"]
	if !ok {
		return nil
	}
	out := []string{}
	for _, p := range providers {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// CreateShipping handles the create shipping http method.
func (s *Shipment) CreateShipping(w http.ResponseWriter, r *http.Request) {
	response := response.New(w)
	req := &api.ShippingRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		response.BadRequest("invalid body used, please consult the api")
		s.log.With(log.Error(err)).Error("Failed to decode body")
		return
	}
	defer r.Body.Close()

	providers := s.providerList(r)
	l := s.log.With(log.Strings("providers", providers))
	l.Info("Create shipping")

	resp, err := s.sender.Send(r.Context(), providers, req)
	if err != nil {
		var target *shipment.ErrProviderUnsupported
		if errors.As(err, &target) {
			response.BadRequest(target.Error())
			return
		}
		l.With(log.Error(err)).Error("Failed to ship to providers")
		response.InternalServer("shipment failed")
		return
	}

	response.Created(&api.ShippingResponses{Responses: resp})
}

// Append appends all shipment routes into the router.
func (s *Shipment) Append(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/createShipping", s.CreateShipping)
}
