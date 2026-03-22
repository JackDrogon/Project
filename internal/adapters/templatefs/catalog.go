package templatefs

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"

	domain "github.com/JackDrogon/project/internal/domain/scaffold"
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

func CollectDetails(fsys fs.FS, lang string, manifest Manifest) (Details, error) {
	if _, err := fs.ReadDir(fsys, lang); err != nil {
		return Details{}, fmt.Errorf("unsupported language: %s", lang)
	}

	var files []FileDetail
	vars := map[string]struct{}{}
	fileCount := 0
	templateCount := 0

	if err := walkTemplateFiles(fsys, lang, func(srcPath string, isTemplate bool) error {
		fileCount++
		return recordTemplateDetails(fsys, lang, srcPath, isTemplate, &files, vars, &templateCount)
	}); err != nil {
		return Details{}, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Output < files[j].Output
	})

	variableList := make([]string, 0, len(vars))
	for name := range vars {
		variableList = append(variableList, name)
	}
	sort.Strings(variableList)

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

func recordTemplateDetails(fsys fs.FS, lang, srcPath string, isTemplate bool, files *[]FileDetail, vars map[string]struct{}, templateCount *int) error {
	relative := strings.TrimPrefix(srcPath, lang+"/")
	*files = append(*files, FileDetail{
		Source:     relative,
		Output:     strings.TrimSuffix(relative, TmplSuffix),
		IsTemplate: isTemplate,
	})

	if err := collectVarsFromTemplatePath(srcPath, relative, vars); err != nil {
		return err
	}

	if !isTemplate {
		return nil
	}

	*templateCount++
	content, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return err
	}

	templateVars, err := ExtractTemplateVars(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", srcPath, err)
	}

	addTemplateVarNames(vars, templateVars)
	return nil
}

func collectVarsFromTemplatePath(srcPath, relative string, vars map[string]struct{}) error {
	pathVars, err := ExtractTemplateVars([]byte(relative))
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

func ExtractTemplateVars(content []byte) ([]string, error) {
	tmpl, err := template.New("template-vars").Parse(string(content))
	if err != nil {
		return nil, err
	}

	vars := map[string]struct{}{}
	if tmpl.Tree != nil {
		collectTemplateVars(tmpl.Root, vars)
	}

	result := make([]string, 0, len(vars))
	for name := range vars {
		result = append(result, name)
	}
	sort.Strings(result)

	return result, nil
}

func collectTemplateVars(node parse.Node, vars map[string]struct{}) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}
		for _, child := range n.Nodes {
			collectTemplateVars(child, vars)
		}
	case *parse.ActionNode:
		if n == nil {
			return
		}
		collectPipeVars(n.Pipe, vars)
	case *parse.IfNode:
		if n == nil {
			return
		}
		collectPipeVars(n.Pipe, vars)
		collectTemplateVars(n.List, vars)
		collectTemplateVars(n.ElseList, vars)
	case *parse.RangeNode:
		if n == nil {
			return
		}
		collectPipeVars(n.Pipe, vars)
		collectTemplateVars(n.List, vars)
		collectTemplateVars(n.ElseList, vars)
	case *parse.WithNode:
		if n == nil {
			return
		}
		collectPipeVars(n.Pipe, vars)
		collectTemplateVars(n.List, vars)
		collectTemplateVars(n.ElseList, vars)
	case *parse.TemplateNode:
		if n == nil {
			return
		}
		collectPipeVars(n.Pipe, vars)
	}
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
