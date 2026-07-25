package version

import "testing"

type stubProvider struct{}

func (stubProvider) Info() string    { return "info" }
func (stubProvider) Verbose() string { return "verbose" }

func TestNewService(t *testing.T) {
	t.Parallel()

	svc := NewService(stubProvider{})
	if svc == nil {
		t.Fatal("NewService() = nil")
	}
	if svc.provider == nil {
		t.Fatal("Service.provider = nil")
	}
}

func TestService_ForwardsProviderOutput(t *testing.T) {
	t.Parallel()

	svc := NewService(stubProvider{})

	if got := svc.Info(); got != "info" {
		t.Fatalf("Info() = %q, want %q", got, "info")
	}

	if got := svc.Verbose(); got != "verbose" {
		t.Fatalf("Verbose() = %q, want %q", got, "verbose")
	}
}
