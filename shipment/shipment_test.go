package shipment

import (
	"context"
	"errors"
	"testing"

	"github.com/hoenirvili/axiogate/http/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPayloader struct{ mock.Mock }

func (m *mockPayloader) Payload(req *api.ShippingRequest) []byte {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]byte)
}

func (m *mockPayloader) To() string {
	args := m.Called()
	return args.String(0)
}

type mockStorage struct{ mock.Mock }

func (m *mockStorage) Save(ctx context.Context, provider string, payload []byte) error {
	args := m.Called(ctx, provider, payload)
	return args.Error(0)
}

type mockClient struct{ mock.Mock }

func (m *mockClient) Do(ctx context.Context, to string, payload any) ([]byte, error) {
	args := m.Called(ctx, to, payload)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestShipment_Send(t *testing.T) {
	tests := []struct {
		name           string
		providers      []string
		setupMocks     func(*mockClient, *mockStorage, map[string]*mockPayloader)
		setupProviders func() map[string]Payloader
		request        *api.ShippingRequest
		wantErr        bool
		errType        error
		validateResp   func(t *testing.T, responses []api.ShippingResponse)
	}{
		{
			name:      "successful send to all providers when providers list is empty",
			providers: []string{},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				p1.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider1"}`))
				p1.On("To").Return("https://provider1.example.com")

				p2 := new(mockPayloader)
				p2.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider2"}`))
				p2.On("To").Return("https://provider2.example.com")

				return map[string]Payloader{
					"provider1": p1,
					"provider2": p2,
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				mc.On("Do", mock.Anything, "https://provider1.example.com", []byte(`{"provider":"provider1"}`)).
					Return([]byte(`{"tracking_id":"1"}`), nil)
				mc.On("Do", mock.Anything, "https://provider2.example.com", []byte(`{"provider":"provider2"}`)).
					Return([]byte(`{"tracking_id":"2"}`), nil)

				ms.On("Save", mock.Anything, "provider1", []byte(`{"tracking_id":"1"}`)).Return(nil)
				ms.On("Save", mock.Anything, "provider2", []byte(`{"tracking_id":"2"}`)).Return(nil)
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 10.5, Unit: "KG"},
			},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Len(t, responses, 2)
				// Verify all responses have no errors
				for _, resp := range responses {
					assert.Empty(t, resp.Error)
					assert.NotNil(t, resp.RawResponse)
				}
			},
		},
		{
			name:      "successful send to specific providers",
			providers: []string{"provider1"},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				p1.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider1"}`))
				p1.On("To").Return("https://provider1.example.com")

				p2 := new(mockPayloader)
				// p2 should not be called

				return map[string]Payloader{
					"provider1": p1,
					"provider2": p2,
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				mc.On("Do", mock.Anything, "https://provider1.example.com", []byte(`{"provider":"provider1"}`)).
					Return([]byte(`{"tracking_id":"1"}`), nil)

				ms.On("Save", mock.Anything, "provider1", []byte(`{"tracking_id":"1"}`)).Return(nil)
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 5.0, Unit: "KG"},
			},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Len(t, responses, 1)
				assert.Equal(t, "https://provider1.example.com", responses[0].Endpoint)
				assert.Empty(t, responses[0].Error)
			},
		},
		{
			name:      "error when provider is not supported",
			providers: []string{"unsupported_provider"},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				return map[string]Payloader{
					"provider1": p1,
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				// No mocks needed as error should occur before calling client
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 5.0, Unit: "KG"},
			},
			wantErr: true,
			errType: &ErrProviderUnsupported{},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Nil(t, responses)
			},
		},
		{
			name:      "client returns error during Do call",
			providers: []string{"provider1"},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				p1.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider1"}`))
				p1.On("To").Return("https://provider1.example.com")

				return map[string]Payloader{
					"provider1": p1,
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				mc.On("Do", mock.Anything, "https://provider1.example.com", []byte(`{"provider":"provider1"}`)).
					Return([]byte(`{"error":"network timeout"}`), errors.New("network timeout"))
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 5.0, Unit: "KG"},
			},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Len(t, responses, 1)
				assert.Equal(t, "network timeout", responses[0].Error)
				assert.Equal(t, "https://provider1.example.com", responses[0].Endpoint)
				assert.NotNil(t, responses[0].RawResponse)
			},
		},
		{
			name:      "storage returns error during Save call",
			providers: []string{"provider1"},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				p1.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider1"}`))
				p1.On("To").Return("https://provider1.example.com")

				return map[string]Payloader{
					"provider1": p1,
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				mc.On("Do", mock.Anything, "https://provider1.example.com", []byte(`{"provider":"provider1"}`)).
					Return([]byte(`{"tracking_id":"1"}`), nil)

				ms.On("Save", mock.Anything, "provider1", []byte(`{"tracking_id":"1"}`)).
					Return(errors.New("database connection failed"))
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 5.0, Unit: "KG"},
			},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Len(t, responses, 1)
				assert.Equal(t, "database connection failed", responses[0].Error)
				assert.Equal(t, "https://provider1.example.com", responses[0].Endpoint)
				assert.NotNil(t, responses[0].RawResponse)
			},
		},
		{
			name:      "multiple providers with mixed success and failure",
			providers: []string{"provider1", "provider2", "provider3"},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				p1.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider1"}`))
				p1.On("To").Return("https://provider1.example.com")

				p2 := new(mockPayloader)
				p2.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider2"}`))
				p2.On("To").Return("https://provider2.example.com")

				p3 := new(mockPayloader)
				p3.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider3"}`))
				p3.On("To").Return("https://provider3.example.com")

				return map[string]Payloader{
					"provider1": p1,
					"provider2": p2,
					"provider3": p3,
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				// provider1 succeeds
				mc.On("Do", mock.Anything, "https://provider1.example.com", []byte(`{"provider":"provider1"}`)).
					Return([]byte(`{"tracking_id":"1"}`), nil)
				ms.On("Save", mock.Anything, "provider1", []byte(`{"tracking_id":"1"}`)).Return(nil)

				// provider2 fails at client Do
				mc.On("Do", mock.Anything, "https://provider2.example.com", []byte(`{"provider":"provider2"}`)).
					Return(nil, errors.New("client error"))

				// provider3 fails at storage Save
				mc.On("Do", mock.Anything, "https://provider3.example.com", []byte(`{"provider":"provider3"}`)).
					Return([]byte(`{"tracking_id":"3"}`), nil)
				ms.On("Save", mock.Anything, "provider3", []byte(`{"tracking_id":"3"}`)).
					Return(errors.New("storage error"))
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 5.0, Unit: "KG"},
			},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Len(t, responses, 3)

				// Check that we have one success and two failures
				successCount := 0
				errorCount := 0
				for _, resp := range responses {
					if resp.Error == "" {
						successCount++
					} else {
						errorCount++
					}
				}
				assert.Equal(t, 1, successCount)
				assert.Equal(t, 2, errorCount)
			},
		},
		{
			name:      "nil payloader is skipped",
			providers: []string{},
			setupProviders: func() map[string]Payloader {
				p1 := new(mockPayloader)
				p1.On("Payload", mock.AnythingOfType("*api.ShippingRequest")).Return([]byte(`{"provider":"provider1"}`))
				p1.On("To").Return("https://provider1.example.com")

				return map[string]Payloader{
					"provider1": p1,
					"provider2": nil, // nil payloader should be skipped
				}
			},
			setupMocks: func(mc *mockClient, ms *mockStorage, mp map[string]*mockPayloader) {
				mc.On("Do", mock.Anything, "https://provider1.example.com", []byte(`{"provider":"provider1"}`)).
					Return([]byte(`{"tracking_id":"1"}`), nil)
				ms.On("Save", mock.Anything, "provider1", []byte(`{"tracking_id":"1"}`)).Return(nil)
			},
			request: &api.ShippingRequest{
				Weight: api.Weight{Value: 5.0, Unit: "KG"},
			},
			validateResp: func(t *testing.T, responses []api.ShippingResponse) {
				assert.Len(t, responses, 1)
				assert.Equal(t, "https://provider1.example.com", responses[0].Endpoint)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockClient := new(mockClient)
			mockStorage := new(mockStorage)
			providers := tt.setupProviders()

			mockPayloaders := make(map[string]*mockPayloader)
			for name, p := range providers {
				if mp, ok := p.(*mockPayloader); ok {
					mockPayloaders[name] = mp
				}
			}

			tt.setupMocks(mockClient, mockStorage, mockPayloaders)

			shipment := New(mockClient, providers, mockStorage)
			responses, err := shipment.Send(
				context.Background(),
				tt.providers,
				tt.request,
			)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			// Validate responses
			if tt.validateResp != nil {
				tt.validateResp(t, responses)
			}

			// Assert expectations on mocks
			mockClient.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
			for _, mp := range mockPayloaders {
				mp.AssertExpectations(t)
			}
		})
	}
}
