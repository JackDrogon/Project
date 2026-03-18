package scaffold

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"
)

// TemplateSummary describes a template language at a glance.
type TemplateSummary struct {
	Name          string   `json:"name"`
	FileCount     int      `json:"file_count"`
	TemplateCount int      `json:"template_count"`
	Variables     []string `json:"variables"`
}

// TemplateFile describes one file in a language template.
type TemplateFile struct {
	Source     string `json:"source"`
	Output     string `json:"output"`
	IsTemplate bool   `json:"is_template"`
}

// TemplateDetails is a full inspection result for a language template.
type TemplateDetails struct {
	TemplateSummary
	Files []TemplateFile `json:"files"`
}

// ListTemplateSummaries returns metadata for all available template languages.
func (c *Creator) ListTemplateSummaries() ([]TemplateSummary, error) {
	langs, err := c.ListLangs()
	if err != nil {
		return nil, err
	}

	summaries := make([]TemplateSummary, 0, len(langs))
	for _, lang := range langs {
		details, err := c.inspectLangDetails(lang)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, details.TemplateSummary)
	}

	return summaries, nil
}

// InspectLang returns detailed file-level metadata for one language template.
func (c *Creator) InspectLang(lang string) (TemplateDetails, error) {
	if _, err := fs.ReadDir(c.fsys, lang); err != nil {
		return TemplateDetails{}, fmt.Errorf("unsupported language: %s", lang)
	}

	return c.inspectLangDetails(lang)
}

func (c *Creator) inspectLangDetails(lang string) (TemplateDetails, error) {
	var files []TemplateFile
	vars := map[string]struct{}{}
	fileCount := 0
	templateCount := 0

	if err := c.walkTemplateFiles(lang, func(srcPath string, isTemplate bool) error {
		fileCount++

		relative := strings.TrimPrefix(srcPath, lang+"/")
		output := strings.TrimSuffix(relative, tmplSuffix)
		files = append(files, TemplateFile{
			Source:     relative,
			Output:     output,
			IsTemplate: isTemplate,
		})

		if !isTemplate {
			return nil
		}

		templateCount++
		content, err := fs.ReadFile(c.fsys, srcPath)
		if err != nil {
			return err
		}

		templateVars, err := extractTemplateVars(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", srcPath, err)
		}

		for _, name := range templateVars {
			vars[name] = struct{}{}
		}

		return nil
	}); err != nil {
		return TemplateDetails{}, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Output < files[j].Output
	})

	var variableList []string
	for name := range vars {
		variableList = append(variableList, name)
	}
	sort.Strings(variableList)

	return TemplateDetails{
		TemplateSummary: TemplateSummary{
			Name:          lang,
			FileCount:     fileCount,
			TemplateCount: templateCount,
			Variables:     variableList,
		},
		Files: files,
	}, nil
}

func (c *Creator) walkTemplateFiles(dir string, visit func(srcPath string, isTemplate bool) error) error {
	entries, err := fs.ReadDir(c.fsys, dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := path.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := c.walkTemplateFiles(srcPath, visit); err != nil {
				return err
			}
			continue
		}

		if err := visit(srcPath, strings.HasSuffix(entry.Name(), tmplSuffix)); err != nil {
			return err
		}
	}

	return nil
}

func extractTemplateVars(content []byte) ([]string, error) {
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
