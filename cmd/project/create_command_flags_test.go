package main

import (
	"strings"
	"testing"

	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func TestCreateCommandFlags_ParseSetRejectsMalformedDuplicateAndReservedKeys(t *testing.T) {
	tests := []struct {
		name      string
		setValues []string
		wantErr   string
	}{
		{
			name:      "missing equals separator",
			setValues: []string{"author"},
			wantErr:   "must be key=value",
		},
		{
			name:      "empty key",
			setValues: []string{"=alice"},
			wantErr:   "key must not be empty",
		},
		{
			name:      "duplicate key",
			setValues: []string{"author=alice", "author=bob"},
			wantErr:   "specified more than once",
		},
		{
			name:      "reserved lang key",
			setValues: []string{"lang=go"},
			wantErr:   "reserved for command options",
		},
		{
			name:      "reserved module path key",
			setValues: []string{"module_path=example.com/demo"},
			wantErr:   "reserved for command options",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := scaffoldCommandFlags{setValues: tt.setValues}

			_, err := appcreate.NewService().ParseSetValues(flags.toAppFlags())
			if err == nil {
				t.Fatal("ParseSetValues() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("ParseSetValues() error = %v, want contains %q", err, tt.wantErr)
			}
		})
	}
}
