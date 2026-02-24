package protocol

import (
	"strings"
	"testing"
)

func TestBlueprint_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bp      Blueprint
		wantErr bool
	}{
		{
			name: "Valid Blueprint",
			bp: Blueprint{
				ServiceID:   "service-1",
				Intent:      "Do something",
				Constraints: map[string]interface{}{"key": "value"},
			},
			wantErr: false,
		},
		{
			name: "Missing ServiceID",
			bp: Blueprint{
				ServiceID:   "",
				Intent:      "Do something",
			},
			wantErr: true,
		},
		{
			name: "Invalid ServiceID",
			bp: Blueprint{
				ServiceID:   "service_1", // underscore not allowed
				Intent:      "Do something",
			},
			wantErr: true,
		},
		{
			name: "Missing Intent",
			bp: Blueprint{
				ServiceID:   "service-1",
				Intent:      "",
			},
			wantErr: true,
		},
		{
			name: "Constraints Too Large",
			bp: Blueprint{
				ServiceID:   "service-1",
				Intent:      "Do something",
				Constraints: map[string]interface{}{"large": strings.Repeat("x", 1024*1024+1)},
			},
			wantErr: true,
		},
		{
			name: "Invalid Dependency Format",
			bp: Blueprint{
				ServiceID:    "service-1",
				Intent:       "Do something",
				Dependencies: []string{"invalid_dependency"},
			},
			wantErr: true,
		},
		{
			name: "Self Dependency",
			bp: Blueprint{
				ServiceID:    "service-1",
				Intent:       "Do something",
				Dependencies: []string{"service-1"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.bp.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
