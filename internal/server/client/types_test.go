package client

import "testing"

func TestResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "Success status",
			status: statusSuccess,
			want:   true,
		},
		{
			name:   "Error status",
			status: statusError,
			want:   false,
		},
		{
			name:   "Warning status",
			status: statusWarning,
			want:   false,
		},
		{
			name:   "Empty status",
			status: "",
			want:   false,
		},
		{
			name:   "Unknown status",
			status: "unknown",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &response{Status: tt.status}
			if got := r.IsSuccess(); got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}
