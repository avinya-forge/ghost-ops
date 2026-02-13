package protocol

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStateStore(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStateStore()

	tests := []struct {
		name    string
		service ServiceRecord
		wantErr bool
	}{
		{
			name: "CreateNewService",
			service: ServiceRecord{
				ServiceID:          "svc-1",
				WASMHash:           "hash-1",
				CurrentState:       StateActive,
				SynthesisTimestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "UpdateService",
			service: ServiceRecord{
				ServiceID:          "svc-1",
				WASMHash:           "hash-2",
				CurrentState:       StateShadow,
				SynthesisTimestamp: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.UpdateService(ctx, tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateService() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				got, err := store.GetService(ctx, tt.service.ServiceID)
				if err != nil {
					t.Fatalf("GetService() error = %v", err)
				}
				if got.WASMHash != tt.service.WASMHash {
					t.Errorf("GetService() hash = %v, want %v", got.WASMHash, tt.service.WASMHash)
				}
			}
		})
	}

	// Test GetService Not Found
	t.Run("GetNonExistentService", func(t *testing.T) {
		_, err := store.GetService(ctx, "non-existent")
		if err == nil {
			t.Error("GetService() expected error for non-existent service")
		}
	})

	// Test ListServices
	t.Run("ListServices", func(t *testing.T) {
		// We have "svc-1" from previous tests.
		// Add another one.
		store.UpdateService(ctx, ServiceRecord{ServiceID: "svc-2", CurrentState: StateActive})

		list, err := store.ListServices(ctx)
		if err != nil {
			t.Fatalf("ListServices() error = %v", err)
		}
		if len(list) != 2 {
			t.Errorf("ListServices() count = %d, want 2", len(list))
		}
	})
}
