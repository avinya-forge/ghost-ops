package gateway

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockRuntimeHost implements a mock protocol.RuntimeHost for testing.
type MockRuntimeHost struct {
	InvokedService string
	InvokedMethod  string
	InvokedPayload []byte
}

func (m *MockRuntimeHost) LoadModule(ctx context.Context, serviceID, version string, wasmBytes []byte) error {
	return nil
}

func (m *MockRuntimeHost) Invoke(ctx context.Context, serviceID, method string, payload []byte) ([]byte, error) {
	m.InvokedService = serviceID
	m.InvokedMethod = method
	m.InvokedPayload = payload
	return []byte(fmt.Sprintf("mock response from %s", serviceID)), nil
}

func (m *MockRuntimeHost) SetActiveVersion(ctx context.Context, serviceID, version string) error {
	return nil
}

func (m *MockRuntimeHost) SetShadowVersion(ctx context.Context, serviceID, version string) error {
	return nil
}

func (m *MockRuntimeHost) UnsetShadowVersion(ctx context.Context, serviceID string) error {
	return nil
}

func (m *MockRuntimeHost) UnloadVersion(ctx context.Context, serviceID, version string) error {
	return nil
}

func (m *MockRuntimeHost) CheckHealth(ctx context.Context, serviceID string) error {
	return nil
}

func (m *MockRuntimeHost) GetActiveServiceCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *MockRuntimeHost) GetLogs(ctx context.Context, serviceID string) ([]byte, error) {
	return nil, nil
}

func (m *MockRuntimeHost) Close(ctx context.Context) error {
	return nil
}

func TestPathRouting(t *testing.T) {
	mockHost := &MockRuntimeHost{}
	gateway := NewGateway(mockHost)

	req, _ := http.NewRequest("GET", "/api/v1/auth", bytes.NewBufferString("test payload"))
	w := httptest.NewRecorder()

	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if mockHost.InvokedService != "auth-service" {
		t.Errorf("expected invoked service 'auth-service', got '%s'", mockHost.InvokedService)
	}

	if w.Body.String() != "mock response from auth-service" {
		t.Errorf("expected mock response, got '%s'", w.Body.String())
	}
}

func TestHeaderRouting(t *testing.T) {
	mockHost := &MockRuntimeHost{}
	gateway := NewGateway(mockHost)

	req, _ := http.NewRequest("POST", "/api/v1/user", bytes.NewBufferString("test payload"))
	req.Header.Set("X-Beta", "true")
	w := httptest.NewRecorder()

	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if mockHost.InvokedService != "beta-service" {
		t.Errorf("expected invoked service 'beta-service', got '%s'", mockHost.InvokedService)
	}

	if w.Body.String() != "mock response from beta-service" {
		t.Errorf("expected mock response, got '%s'", w.Body.String())
	}
}

func TestNotFound(t *testing.T) {
	mockHost := &MockRuntimeHost{}
	gateway := NewGateway(mockHost)

	req, _ := http.NewRequest("GET", "/unknown/path", bytes.NewBufferString("test payload"))
	w := httptest.NewRecorder()

	gateway.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
