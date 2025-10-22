package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/hoenirvili/axiogate/http/api"
	"github.com/hoenirvili/axiogate/shipment"
)

type mockSender struct{ mock.Mock }

func (m *mockSender) Send(ctx context.Context, providers []string, req *api.ShippingRequest) ([]api.ShippingResponse, error) {
	args := m.Called(ctx, providers, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]api.ShippingResponse), args.Error(1)
}

func TestShipmentCreateShipping(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		queryParams        map[string][]string
		setupMock          func(*mockSender)
		expectedStatusCode int
		validateResponse   func(t *testing.T, body []byte)
	}{
		{
			name:               "invalid JSON body",
			requestBody:        "invalid json",
			setupMock:          func(ms *mockSender) {},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "invalid body used")
			},
		},
		{
			name:        "successful request without providers",
			requestBody: &api.ShippingRequest{},
			setupMock: func(ms *mockSender) {
				ms.On("Send", mock.Anything, []string(nil), mock.AnythingOfType("*api.ShippingRequest")).
					Return([]api.ShippingResponse{
						{
							Endpoint:    "https://provider1.example.com/ship",
							RawResponse: json.RawMessage(`{"tracking_id":"ABC123","status":"created"}`),
							Error:       "",
						},
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var resp api.ShippingResponses
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Len(t, resp.Responses, 1)
				assert.Equal(t, "https://provider1.example.com/ship", resp.Responses[0].Endpoint)
				assert.NotNil(t, resp.Responses[0].RawResponse)
				assert.Empty(t, resp.Responses[0].Error)
			},
		},
		{
			name:        "successful request with specific providers",
			requestBody: &api.ShippingRequest{},
			queryParams: map[string][]string{
				"providers": {"provider1", "provider2"},
			},
			setupMock: func(ms *mockSender) {
				ms.On("Send", mock.Anything, []string{"provider1", "provider2"}, mock.AnythingOfType("*api.ShippingRequest")).
					Return([]api.ShippingResponse{
						{
							Endpoint:    "https://provider1.example.com/ship",
							RawResponse: json.RawMessage(`{"tracking_id":"ABC123"}`),
							Error:       "",
						},
						{
							Endpoint:    "https://provider2.example.com/ship",
							RawResponse: json.RawMessage(`{"tracking_id":"DEF456"}`),
							Error:       "",
						},
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var resp api.ShippingResponses
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Len(t, resp.Responses, 2)
				assert.Equal(t, "https://provider1.example.com/ship", resp.Responses[0].Endpoint)
				assert.Equal(t, "https://provider2.example.com/ship", resp.Responses[1].Endpoint)
			},
		},
		{
			name:        "provider list filters with empty strings in between",
			requestBody: &api.ShippingRequest{},
			queryParams: map[string][]string{
				"providers": {"provider1", "", "provider2", ""},
			},
			setupMock: func(ms *mockSender) {
				ms.On("Send", mock.Anything, []string{"provider1", "provider2"}, mock.AnythingOfType("*api.ShippingRequest")).
					Return([]api.ShippingResponse{
						{
							Endpoint:    "https://provider1.example.com/ship",
							RawResponse: json.RawMessage(`{"tracking_id":"ABC123"}`),
							Error:       "",
						},
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var resp api.ShippingResponses
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Responses)
			},
		},
		{
			name:        "successful response with error field populated",
			requestBody: &api.ShippingRequest{},
			queryParams: map[string][]string{
				"providers": {"provider1"},
			},
			setupMock: func(ms *mockSender) {
				ms.On("Send", mock.Anything, []string{"provider1"}, mock.AnythingOfType("*api.ShippingRequest")).
					Return([]api.ShippingResponse{
						{
							Endpoint:    "https://provider1.example.com/ship",
							RawResponse: json.RawMessage(`{}`),
							Error:       "partial failure at provider level",
						},
					}, nil)
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var resp api.ShippingResponses
				err := json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Len(t, resp.Responses, 1)
				assert.Equal(t, "partial failure at provider level", resp.Responses[0].Error)
			},
		},
		{
			name:        "sender returns ErrProviderUnsupported",
			requestBody: &api.ShippingRequest{},
			queryParams: map[string][]string{
				"providers": {"whatever"},
			},
			setupMock: func(ms *mockSender) {
				ms.On("Send", mock.Anything, []string{"whatever"}, mock.AnythingOfType("*api.ShippingRequest")).
					Return(nil, &shipment.ErrProviderUnsupported{Provider: "whatever"})
			},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "unsupported")
			},
		},
		{
			name:        "sender returns generic error",
			requestBody: &api.ShippingRequest{},
			setupMock: func(ms *mockSender) {
				ms.On("Send", mock.Anything, []string(nil), mock.AnythingOfType("*api.ShippingRequest")).
					Return(nil, errors.New("internal service error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "shipment failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSender := new(mockSender)
			tt.setupMock(mockSender)
			shipment := NewShipment(mockSender)

			var bodyReader *bytes.Reader
			if strBody, ok := tt.requestBody.(string); ok {
				bodyReader = bytes.NewReader([]byte(strBody))
			} else {
				bodyBytes, err := json.Marshal(tt.requestBody)
				assert.NoError(t, err)
				bodyReader = bytes.NewReader(bodyBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/shipping", bodyReader)
			if tt.queryParams != nil {
				q := req.URL.Query()
				for key, values := range tt.queryParams {
					for _, value := range values {
						q.Add(key, value)
					}
				}
				req.URL.RawQuery = q.Encode()
			}
			rr := httptest.NewRecorder()

			shipment.CreateShipping(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if tt.validateResponse != nil {
				tt.validateResponse(t, rr.Body.Bytes())
			}

			mockSender.AssertExpectations(t)
		})
	}
}
