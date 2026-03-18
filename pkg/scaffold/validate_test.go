package scaffold

import (
	"testing"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "myproject", false},
		{"valid with hyphen", "my-project", false},
		{"valid with underscore", "my_project", false},
		{"valid with dot", "my.project", false},
		{"valid with numbers", "project123", false},
		{"valid single letter", "a", false},
		{"valid mixed", "My-Project_v2.0", false},

		{"empty", "", true},
		{"starts with number", "1project", true},
		{"starts with hyphen", "-project", true},
		{"starts with dot", ".project", true},
		{"starts with underscore", "_project", true},
		{"contains space", "my project", true},
		{"contains slash", "my/project", true},
		{"path traversal", "../../../etc", true},
		{"contains at sign", "my@project", true},
		{"contains exclamation", "my!project", true},
		{"too long", string(make([]byte, 256)), true},
		{"max length", func() string {
			b := make([]byte, 255)
			b[0] = 'a'
			for i := 1; i < 255; i++ {
				b[i] = '0'
			}
			return string(b)
		}(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateModulePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty is allowed", "", false},
		{"simple path", "demo", false},
		{"domain path", "example.com/demo", false},
		{"nested path", "github.com/user/project", false},
		{"mixed casing", "github.com/Azure/project", false},
		{"tilde allowed", "example.com/lib~preview", false},

		{"leading slash", "/demo", true},
		{"trailing slash", "demo/", true},
		{"double slash", "example.com//demo", true},
		{"trailing dot in segment", "example.com/demo.", true},
		{"leading dot in segment", "example.com/.demo", true},
		{"space", "example.com/my project", true},
		{"backslash", `example.com\\demo`, true},
		{"scheme", "https://example.com/demo", true},
		{"current segment", "example.com/./demo", true},
		{"parent segment", "example.com/../demo", true},
		{"quote", "example.com/o'connor/demo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModulePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModulePath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
