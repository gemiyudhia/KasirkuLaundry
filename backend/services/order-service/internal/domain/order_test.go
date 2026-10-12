package domain

import "testing"

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from OrderStatus
		to   OrderStatus
		want bool
	}{
		{"received to processing", StatusReceived, StatusProcessing, true},
		{"processing to ready", StatusProcessing, StatusReady, true},
		{"ready to completed", StatusReady, StatusCompleted, true},

		{"skip: received to ready", StatusReceived, StatusReady, false},
		{"skip: received to completed", StatusReceived, StatusCompleted, false},
		{"skip: processing to completed", StatusProcessing, StatusCompleted, false},

		{"backward: ready to processing", StatusReady, StatusProcessing, false},
		{"backward: processing to received", StatusProcessing, StatusReceived, false},

		{"completed to received", StatusCompleted, StatusReceived, false},
		{"completed to processing", StatusCompleted, StatusProcessing, false},
		{"completed to completed", StatusCompleted, StatusCompleted, false},

		{"same status", StatusReceived, StatusReceived, false},

		{"completed to empty", StatusCompleted, "", false},
		{"unknown from status", OrderStatus("foo"), StatusProcessing, false},
		{"unknown to status", StatusReceived, OrderStatus("foo"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.want {
				t.Errorf("%q.CanTransitionTo(%q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		status OrderStatus
		want   bool
	}{
		{StatusReceived, true},
		{StatusProcessing, true},
		{StatusReady, true},
		{StatusCompleted, true},
		{"", false},
		{"foo", false},
		{"RECEIVED", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("%q.IsValid() = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
