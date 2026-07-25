package templatefs

import (
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"text/template"
	"text/template/parse"

	domain "github.com/JackDrogon/project/internal/scaffold"
)

type Manifest struct {
	SchemaVersion int
	Name          string
	Description   string
	Inputs        []domain.ManifestInput
}

type Summary struct {
	Name            string
	Description     string
	ManifestVersion int
	InputNames      []string
	FileCount       int
	TemplateCount   int
	Variables       []string
}

type FileDetail struct {
	Source     string
	Output     string
	IsTemplate bool
}

type Details struct {
	Summary
	Inputs []domain.ManifestInput
	Files  []FileDetail
}

// CollectDetails scans a template directory and collects comprehensive metadata including
// file listings, template variables, and manifest inputs. It walks the template tree,
// analyzes .tmpl files to extract variables, and returns structured details for inspection.
// Returns an error if the language is unsupported or if template parsing fails.
func CollectDetails(fsys fs.FS, lang string, manifest Manifest) (Details, error) {
	if _, err := fs.ReadDir(fsys, lang); err != nil {
		return Details{}, fmt.Errorf("unsupported language: %s", lang)
	}

	acc := newDetailsAccumulator()
	fileCount := 0

	if err := walkTemplateFiles(fsys, lang, func(srcPath string, isTemplate bool) error {
		fileCount++
		return acc.record(fsys, lang, srcPath, isTemplate)
	}); err != nil {
		return Details{}, err
	}

	files := acc.files
	templateCount := acc.templateCount

	slices.SortFunc(files, func(a, b FileDetail) int {
		return strings.Compare(a.Output, b.Output)
	})

	variableList := slices.Sorted(maps.Keys(acc.vars))

	inputNames := make([]string, 0, len(manifest.Inputs))
	for _, input := range manifest.Inputs {
		inputNames = append(inputNames, input.Name)
	}

	inputs := append([]domain.ManifestInput(nil), manifest.Inputs...)

	return Details{
		Summary: Summary{
			Name:            lang,
			Description:     manifest.Description,
			ManifestVersion: manifest.SchemaVersion,
			InputNames:      inputNames,
			FileCount:       fileCount,
			TemplateCount:   templateCount,
			Variables:       variableList,
		},
		Inputs: inputs,
		Files:  files,
	}, nil
}

func walkTemplateFiles(fsys fs.FS, dir string, visit func(srcPath string, isTemplate bool) error) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if isReservedTemplateFile(entry.Name()) {
			continue
		}

		srcPath := path.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := walkTemplateFiles(fsys, srcPath, visit); err != nil {
				return err
			}
			continue
		}

		if err := visit(srcPath, strings.HasSuffix(entry.Name(), TmplSuffix)); err != nil {
			return err
		}
	}

	return nil
}

// detailsAccumulator gathers what a template walk discovers. It exists so the
// walk callback can be a method with four parameters instead of a function with
// seven, three of which were out-parameters (*[]FileDetail, the vars map, *int).
type detailsAccumulator struct {
	files         []FileDetail
	vars          map[string]struct{}
	templateCount int
}

func newDetailsAccumulator() *detailsAccumulator {
	return &detailsAccumulator{vars: map[string]struct{}{}}
}

func (a *detailsAccumulator) record(fsys fs.FS, lang, srcPath string, isTemplate bool) error {
	relative := strings.TrimPrefix(srcPath, lang+"/")
	a.files = append(a.files, FileDetail{
		Source:     relative,
		Output:     strings.TrimSuffix(relative, TmplSuffix),
		IsTemplate: isTemplate,
	})

	if err := collectVarsFromTemplatePath(srcPath, relative, a.vars); err != nil {
		return err
	}

	if !isTemplate {
		return nil
	}

	a.templateCount++
	content, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return err
	}

	templateVars, err := extractTemplateVars(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", srcPath, err)
	}

	addTemplateVarNames(a.vars, templateVars)
	return nil
}

func collectVarsFromTemplatePath(srcPath, relative string, vars map[string]struct{}) error {
	pathVars, err := extractTemplateVars([]byte(relative))
	if err != nil {
		return fmt.Errorf("failed to parse template path %s: %w", srcPath, err)
	}

	addTemplateVarNames(vars, pathVars)
	return nil
}

func addTemplateVarNames(dest map[string]struct{}, names []string) {
	for _, name := range names {
		dest[name] = struct{}{}
	}
}

// extractTemplateVars parses template content and extracts all variable names referenced
// in the template syntax (e.g., {{.VariableName}}). It returns a sorted list of unique
// variable names. Returns an error if the template syntax is invalid.
func extractTemplateVars(content []byte) ([]string, error) {
	tmpl, err := template.New("template-vars").Parse(string(content))
	if err != nil {
		return nil, err
	}

	vars := map[string]struct{}{}
	if tmpl.Tree != nil {
		collectTemplateVars(tmpl.Root, vars)
	}

	return slices.Sorted(maps.Keys(vars)), nil
}

func collectTemplateVars(node parse.Node, vars map[string]struct{}) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ListNode:
		collectFromListNode(n, vars)
	case *parse.ActionNode:
		collectFromActionNode(n, vars)
	case *parse.IfNode:
		collectFromBranchNode(n.Pipe, n.List, n.ElseList, vars)
	case *parse.RangeNode:
		collectFromBranchNode(n.Pipe, n.List, n.ElseList, vars)
	case *parse.WithNode:
		collectFromBranchNode(n.Pipe, n.List, n.ElseList, vars)
	case *parse.TemplateNode:
		collectFromTemplateNode(n, vars)
	}
}

func collectFromListNode(n *parse.ListNode, vars map[string]struct{}) {
	if n == nil {
		return
	}
	for _, child := range n.Nodes {
		collectTemplateVars(child, vars)
	}
}

func collectFromActionNode(n *parse.ActionNode, vars map[string]struct{}) {
	if n == nil {
		return
	}
	collectPipeVars(n.Pipe, vars)
}

func collectFromBranchNode(pipe *parse.PipeNode, list, elseList *parse.ListNode, vars map[string]struct{}) {
	collectPipeVars(pipe, vars)
	collectTemplateVars(list, vars)
	collectTemplateVars(elseList, vars)
}

func collectFromTemplateNode(n *parse.TemplateNode, vars map[string]struct{}) {
	if n == nil {
		return
	}
	collectPipeVars(n.Pipe, vars)
}

func collectPipeVars(pipe *parse.PipeNode, vars map[string]struct{}) {
	if pipe == nil {
		return
	}

	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			collectArgVars(arg, vars)
		}
	}
}

func collectArgVars(node parse.Node, vars map[string]struct{}) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.FieldNode:
		if n == nil {
			return
		}
		if len(n.Ident) > 0 {
			vars[n.Ident[0]] = struct{}{}
		}
	case *parse.ChainNode:
		if n == nil {
			return
		}
		if len(n.Field) > 0 {
			if _, ok := n.Node.(*parse.DotNode); ok {
				vars[n.Field[0]] = struct{}{}
			}
		}
		collectArgVars(n.Node, vars)
	case *parse.PipeNode:
		if n == nil {
			return
		}
		collectPipeVars(n, vars)
	case *parse.CommandNode:
		if n == nil {
			return
		}
		for _, arg := range n.Args {
			collectArgVars(arg, vars)
		}
	}
}
