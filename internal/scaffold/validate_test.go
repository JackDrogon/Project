package scaffold

import "testing"

func TestValidateProjectName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid simple", input: "myproject"},
		{name: "valid mixed", input: "My-Project_v2.0"},
		{name: "empty", input: "", wantErr: true},
		{name: "starts with number", input: "1project", wantErr: true},
		{name: "contains slash", input: "my/project", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateModulePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty is allowed", input: ""},
		{name: "domain path", input: "example.com/demo"},
		{name: "double slash", input: "example.com//demo", wantErr: true},
		{name: "scheme", input: "https://example.com/demo", wantErr: true},
		{name: "current segment", input: "example.com/./demo", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModulePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateModulePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
