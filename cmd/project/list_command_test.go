package main

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	appcatalog "github.com/JackDrogon/project/internal/app/catalog"
	appcreate "github.com/JackDrogon/project/internal/app/create"
)

func useCatalogServiceFactory(t *testing.T, factory func() *appcatalog.Service) {
	t.Helper()
	oldFactory := newCatalogService
	newCatalogService = factory
	t.Cleanup(func() {
		newCatalogService = oldFactory
	})
}

func newCommandTestCatalogService() *appcatalog.Service {
	fsys := fstest.MapFS{
		"go/.project-template-manifest.toml":        {Data: []byte("version = 2\nname = \"go\"\ndescription = \"Go starter\"\n\n[[inputs]]\nkey = \"module_path\"\ntemplate_var = \"ModulePath\"\nrequired = true\n\n[[inputs]]\nkey = \"go_version\"\ntemplate_var = \"GoVersion\"\nrequired = true\n")},
		"cpp/.project-template-manifest.toml":       {Data: []byte("version = 2\nname = \"cpp\"\ndescription = \"C++ starter\"\n\n[[inputs]]\nkey = \"author\"\ntemplate_var = \"Author\"\nrequired = false\n")},
		"go/.gitignore":                             {Data: []byte("bin/\n")},
		"go/go.mod.tmpl":                            {Data: []byte("module {{.ModulePath}}\n")},
		"go/cmd/{{.ProjectNameLower}}":              {Mode: fs.ModeDir | 0o755},
		"go/cmd/{{.ProjectNameLower}}/main.go.tmpl": {Data: []byte("package main\n")},
		"cpp/src/main.cc.tmpl":                      {Data: []byte("// {{.ProjectName}}\n")},
		"cpp/README.md.tmpl":                        {Data: []byte("By {{.Author}}\n")},
	}

	return appcatalog.NewService(fsys, nil)
}

func TestSelectedOutputFormat(t *testing.T) {
	if got := selectedOutputFormat(false); got != outputFormatText {
		t.Fatalf("selectedOutputFormat(false) = %q, want %q", got, outputFormatText)
	}
	if got := selectedOutputFormat(true); got != outputFormatTOML {
		t.Fatalf("selectedOutputFormat(true) = %q, want %q", got, outputFormatTOML)
	}
}

func TestListCmdOutputs(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantEither  [][2]string
	}{
		{name: "text", wantContain: []string{"cpp", "go"}},
		{name: "toml", args: []string{"--toml"}, wantContain: []string{"languages = ", "cpp", "go"}},
		{name: "detail text", args: []string{"--detail"}, wantContain: []string{`desc="C++ starter"`, `manifest=v2`, `inputs=[author]`, `files=2`, `vars=[Author, ProjectName]`}},
		{name: "detail toml", args: []string{"--detail", "--toml"}, wantContain: []string{"[[templates]]", "manifest_version = 2", "file_count = 2", "template_count = 2"}, wantEither: [][2]string{{"name = 'cpp'", `name = "cpp"`}, {"input_names = ['author']", `input_names = ["author"]`}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := requireSubcommand(t, creator, commandKeyList)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			got := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want contains %q", got, want)
				}
			}
			for _, want := range tt.wantEither {
				if !strings.Contains(got, want[0]) && !strings.Contains(got, want[1]) {
					t.Fatalf("output = %q, want contains either %q or %q", got, want[0], want[1])
				}
			}
		})
	}
}

func TestListCmd_PropagatesCatalogErrors(t *testing.T) {
	useCatalogServiceFactory(t, func() *appcatalog.Service {
		return appcatalog.NewService(failingCatalogFS{err: errors.New("boom")}, nil)
	})
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	tests := []struct {
		name string
		args []string
	}{
		{name: "langs", args: nil},
		{name: "summaries", args: []string{"--detail"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := requireSubcommand(t, creator, commandKeyList)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if !strings.Contains(err.Error(), "failed to read templates") {
				t.Fatalf("Execute() error = %v, want template read error", err)
			}
		})
	}
}

func TestInspectCmdOutputs(t *testing.T) {
	useCatalogServiceFactory(t, newCommandTestCatalogService)
	creator := appcreate.NewCreator(fstest.MapFS{}, &bytes.Buffer{})

	tests := []struct {
		name        string
		args        []string
		wantContain []string
		wantEither  [][2]string
		wantErr     string
	}{
		{name: "text", args: []string{"go"}, wantContain: []string{"name: go", "manifest_version: 2", "mode: all", "shown: 3"}},
		{name: "toml render", args: []string{"go", "--toml", "--mode", "render"}, wantContain: []string{"manifest_version = 2", "mode = ", "shown_count = 2", "[[files]]", "source = "}, wantEither: [][2]string{{"name = 'go'", `name = "go"`}, {"mode = 'render'", `mode = "render"`}, {"source = 'go.mod.tmpl'", `source = "go.mod.tmpl"`}}},
		{name: "invalid mode", args: []string{"go", "--mode", "bogus"}, wantErr: "invalid --mode"},
		{name: "unsupported language", args: []string{"missing"}, wantErr: "unsupported language"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := requireSubcommand(t, creator, commandKeyInspect)
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("Execute() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Execute() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			got := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Fatalf("output = %q, want contains %q", got, want)
				}
			}
			for _, want := range tt.wantEither {
				if !strings.Contains(got, want[0]) && !strings.Contains(got, want[1]) {
					t.Fatalf("output = %q, want contains either %q or %q", got, want[0], want[1])
				}
			}
		})
	}
}
