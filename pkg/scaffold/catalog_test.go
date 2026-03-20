package scaffold

import (
	"bytes"
	"errors"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"
	"text/template"
	"text/template/parse"
)

type catalogErrorFS struct {
	fstest.MapFS
	readDirPath  string
	readFilePath string
	err          error
}

func (f catalogErrorFS) Open(name string) (fs.File, error) {
	return f.MapFS.Open(name)
}

func (f catalogErrorFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == f.readDirPath {
		return nil, f.err
	}
	return fs.ReadDir(f.MapFS, name)
}

func (f catalogErrorFS) ReadFile(name string) ([]byte, error) {
	if name == f.readFilePath {
		return nil, f.err
	}
	return fs.ReadFile(f.MapFS, name)
}

func TestListTemplateSummaries(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":       {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}")},
		"go/main.go.tmpl":      {Data: []byte("package main\n")},
		"go/.gitignore":        {Data: []byte("bin/")},
		"cpp/src/main.cc.tmpl": {Data: []byte("// {{.ProjectName}}")},
		"cpp/README.md.tmpl":   {Data: []byte("By {{.Author}}")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	summaries, err := c.ListTemplateSummaries()
	if err != nil {
		t.Fatalf("ListTemplateSummaries() error = %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(summaries))
	}

	want := map[string]TemplateSummary{
		"cpp": {
			Name:          "cpp",
			FileCount:     2,
			TemplateCount: 2,
			Variables:     []string{"Author", "ProjectName"},
		},
		"go": {
			Name:          "go",
			FileCount:     3,
			TemplateCount: 2,
			Variables:     []string{"GoVersion", "ModulePath"},
		},
	}

	for _, got := range summaries {
		expected, ok := want[got.Name]
		if !ok {
			t.Fatalf("unexpected summary: %q", got.Name)
		}
		if got.FileCount != expected.FileCount {
			t.Fatalf("%s FileCount = %d, want %d", got.Name, got.FileCount, expected.FileCount)
		}
		if got.TemplateCount != expected.TemplateCount {
			t.Fatalf("%s TemplateCount = %d, want %d", got.Name, got.TemplateCount, expected.TemplateCount)
		}
		if !reflect.DeepEqual(got.Variables, expected.Variables) {
			t.Fatalf("%s Variables = %v, want %v", got.Name, got.Variables, expected.Variables)
		}
	}
}

func TestInspectLang(t *testing.T) {
	fsys := fstest.MapFS{
		"go/go.mod.tmpl":  {Data: []byte("module {{.ModulePath}}\ngo {{.GoVersion}}")},
		"go/main.go.tmpl": {Data: []byte("// {{.ProjectName}}")},
		"go/.gitignore":   {Data: []byte("bin/")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	details, err := c.InspectLang("go")
	if err != nil {
		t.Fatalf("InspectLang() error = %v", err)
	}

	if details.Name != "go" {
		t.Fatalf("details.Name = %q, want %q", details.Name, "go")
	}
	if details.FileCount != 3 {
		t.Fatalf("details.FileCount = %d, want %d", details.FileCount, 3)
	}
	if details.TemplateCount != 2 {
		t.Fatalf("details.TemplateCount = %d, want %d", details.TemplateCount, 2)
	}
	if !reflect.DeepEqual(details.Variables, []string{"GoVersion", "ModulePath", "ProjectName"}) {
		t.Fatalf("details.Variables = %v, want %v", details.Variables, []string{"GoVersion", "ModulePath", "ProjectName"})
	}

	if len(details.Files) != 3 {
		t.Fatalf("len(details.Files) = %d, want %d", len(details.Files), 3)
	}

	wantFiles := []TemplateFile{
		{Source: ".gitignore", Output: ".gitignore", IsTemplate: false},
		{Source: "go.mod.tmpl", Output: "go.mod", IsTemplate: true},
		{Source: "main.go.tmpl", Output: "main.go", IsTemplate: true},
	}

	if !reflect.DeepEqual(details.Files, wantFiles) {
		t.Fatalf("details.Files = %#v, want %#v", details.Files, wantFiles)
	}
}

func TestInspectLang_InvalidTemplateReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	if _, err := c.InspectLang("go"); err == nil {
		t.Fatal("InspectLang() expected error for invalid template, got nil")
	}
}

func TestInspectLang_UnsupportedLanguage(t *testing.T) {
	c := NewCreator(fstest.MapFS{}, &bytes.Buffer{})
	if _, err := c.InspectLang("missing"); err == nil {
		t.Fatal("InspectLang() expected error for unsupported language, got nil")
	}
}

func TestExtractTemplateVars_ComplexTemplate(t *testing.T) {
	content := []byte(`{{define "inner"}}{{.ProjectName}}{{end}}
{{if .ProjectName}}{{template "inner" .}}{{end}}
{{with .Author}}{{.Name}}{{end}}
{{range .Items}}{{.Value}}{{else}}{{.Fallback}}{{end}}
{{printf "%s" .ModulePath}}
{{(print .ProjectName).Suffix}}`)

	got, err := extractTemplateVars(content)
	if err != nil {
		t.Fatalf("extractTemplateVars() error = %v", err)
	}

	want := []string{"Author", "Fallback", "Items", "ModulePath", "Name", "ProjectName", "Value"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractTemplateVars() = %v, want %v", got, want)
	}
}

func TestListTemplateSummaries_InvalidTemplateReturnsError(t *testing.T) {
	fsys := fstest.MapFS{
		"go/bad.txt.tmpl": {Data: []byte("{{.ProjectName")},
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	if _, err := c.ListTemplateSummaries(); err == nil {
		t.Fatal("ListTemplateSummaries() expected error for invalid template, got nil")
	}
}

func TestListTemplateSummaries_ReadDirError(t *testing.T) {
	c := NewCreator(failingReadDirFS{err: bytes.ErrTooLarge}, &bytes.Buffer{})
	if _, err := c.ListTemplateSummaries(); err == nil {
		t.Fatal("ListTemplateSummaries() expected read error, got nil")
	}
}

func TestCollectTemplateHelpers(t *testing.T) {
	vars := map[string]struct{}{}

	collectTemplateVars(nil, vars)
	collectTemplateVars(parse.Node((*parse.ListNode)(nil)), vars)
	collectTemplateVars(parse.Node((*parse.ActionNode)(nil)), vars)
	collectTemplateVars(parse.Node((*parse.IfNode)(nil)), vars)
	collectTemplateVars(parse.Node((*parse.RangeNode)(nil)), vars)
	collectTemplateVars(parse.Node((*parse.WithNode)(nil)), vars)
	collectTemplateVars(parse.Node((*parse.TemplateNode)(nil)), vars)
	collectArgVars(nil, vars)
	collectArgVars(parse.Node((*parse.FieldNode)(nil)), vars)
	collectArgVars(parse.Node((*parse.ChainNode)(nil)), vars)
	collectArgVars(parse.Node((*parse.PipeNode)(nil)), vars)
	collectArgVars(parse.Node((*parse.CommandNode)(nil)), vars)
	collectPipeVars(nil, vars)

	tmpl, err := template.New("helper").Parse(`{{define "inner"}}{{.Fallback}}{{end}}
{{.ProjectName}}
{{if .Author}}{{end}}
{{range .Items}}{{else}}{{end}}
{{with .ModulePath}}{{end}}
{{template "inner" .}}`)
	if err != nil {
		t.Fatalf("template.Parse() error = %v", err)
	}
	collectTemplateVars(tmpl.Root, vars)

	pipe := &parse.PipeNode{Cmds: []*parse.CommandNode{{Args: []parse.Node{
		&parse.FieldNode{Ident: []string{"ProjectName"}},
		&parse.ChainNode{Node: &parse.DotNode{}, Field: []string{"ChainVar"}},
	}}}}
	collectTemplateVars(&parse.ListNode{Nodes: []parse.Node{&parse.ActionNode{Pipe: pipe}}}, vars)

	collectArgVars(&parse.PipeNode{Cmds: []*parse.CommandNode{{Args: []parse.Node{&parse.FieldNode{Ident: []string{"PipeVar"}}}}}}, vars)
	collectArgVars(&parse.CommandNode{Args: []parse.Node{&parse.FieldNode{Ident: []string{"CommandVar"}}}}, vars)
	collectArgVars(&parse.ChainNode{Node: &parse.FieldNode{Ident: []string{"NestedVar"}}, Field: []string{"Ignored"}}, vars)

	want := []string{"Author", "ChainVar", "CommandVar", "Items", "ModulePath", "NestedVar", "PipeVar", "ProjectName"}
	var got []string
	for name := range vars {
		got = append(got, name)
	}
	if len(got) != len(want) {
		t.Fatalf("collected vars = %v, want %v", got, want)
	}
	for _, name := range want {
		if _, ok := vars[name]; !ok {
			t.Fatalf("expected variable %q to be collected, got %v", name, got)
		}
	}
}

func TestInspectLang_ReadFileError(t *testing.T) {
	fsys := catalogErrorFS{
		MapFS:        fstest.MapFS{"go/file.txt.tmpl": {Data: []byte("{{.ProjectName}}")}},
		readFilePath: "go/file.txt.tmpl",
		err:          errors.New("read file failed"),
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	if _, err := c.InspectLang("go"); err == nil {
		t.Fatal("InspectLang() expected read file error, got nil")
	}
}

func TestWalkTemplateFiles_NestedReadDirError(t *testing.T) {
	fsys := catalogErrorFS{
		MapFS:       fstest.MapFS{"go/sub/file.txt": {Data: []byte("content")}},
		readDirPath: "go/sub",
		err:         errors.New("nested read dir failed"),
	}

	c := NewCreator(fsys, &bytes.Buffer{})
	if err := c.walkTemplateFiles("go", func(srcPath string, isTemplate bool) error { return nil }); err == nil {
		t.Fatal("walkTemplateFiles() expected nested read error, got nil")
	}
}
