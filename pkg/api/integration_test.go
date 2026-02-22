package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ghost-ops/pkg/protocol"
	"ghost-ops/pkg/registry"
	"ghost-ops/pkg/telemetry"
)

type MockErrorStore struct {
	listErr error
}

func (m *MockErrorStore) GetService(ctx context.Context, id string) (*protocol.ServiceRecord, error) {
	return nil, nil
}
func (m *MockErrorStore) UpdateService(ctx context.Context, r protocol.ServiceRecord) error {
	return nil
}
func (m *MockErrorStore) ListServices(ctx context.Context) ([]protocol.ServiceRecord, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return nil, nil
}
func (m *MockErrorStore) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}
func (m *MockErrorStore) Set(ctx context.Context, key string, val []byte) error {
	return nil
}

type MockErrorSource struct {
	err error
}

func (m *MockErrorSource) GetNextBlueprint(ctx context.Context) (*protocol.Blueprint, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, nil
}

func TestAPIErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		storeErr       error
		sourceErr      error
		expectedStatus int
	}{
		{
			name:           "ListServices_InternalError",
			method:         http.MethodGet,
			path:           "/services",
			storeErr:       protocol.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "ListServices_Unauthorized",
			method:         http.MethodGet,
			path:           "/services",
			storeErr:       protocol.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Reconcile_InternalError",
			method:         http.MethodPost,
			path:           "/reconcile",
			sourceErr:      protocol.ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Reconcile_InvalidInput",
			method:         http.MethodPost,
			path:           "/reconcile",
			sourceErr:      protocol.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &MockErrorStore{listErr: tc.storeErr}
			source := &MockErrorSource{err: tc.sourceErr}
			collector := telemetry.NewInMemoryCollector()
			// We need a partial registry, other deps can be nil if not used by the path
			// Note: We need to pass nil for engine and runtime as they are not used in these specific error paths
			// However, Reconcile might call them if source succeeds. But here source fails.
			reg := registry.NewRegistry(store, nil, source, nil, collector)
			server := NewServer(reg, collector)

			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			server.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
