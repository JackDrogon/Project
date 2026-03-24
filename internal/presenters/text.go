package presenters

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/JackDrogon/project/internal/app/catalog"
)

func writeTextLangs(w io.Writer, langs []string) error {
	var b strings.Builder
	for _, lang := range langs {
		fmt.Fprintln(&b, lang)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeTextSummaries(w io.Writer, summaries []catalog.Summary) error {
	var b strings.Builder
	for _, summary := range summaries {
		vars := "(none)"
		if len(summary.Variables) > 0 {
			vars = strings.Join(summary.Variables, ", ")
		}

		inputs := "(none)"
		if len(summary.InputNames) > 0 {
			inputs = strings.Join(summary.InputNames, ", ")
		}

		repoAssets := "(none)"
		if len(summary.RepoAssets) > 0 {
			repoAssets = strings.Join(summary.RepoAssets, ", ")
		}

		description := summary.Description
		if description == "" {
			description = "(none)"
		}

		fmt.Fprintf(&b, "%s\tdesc=%q\tmanifest=v%d\tinputs=[%s]\tfiles=%d\ttemplates=%d\tvars=[%s]\trepo=[%s]\trepo_files=%d	governance=%s\n", summary.Name, description, summary.ManifestVersion, inputs, summary.FileCount, summary.TemplateCount, vars, repoAssets, summary.RepoFileCount, summary.GovernanceTier)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeCompactTextSummaries(w io.Writer, summaries []catalog.Summary) error {
	var b strings.Builder
	for i, summary := range summaries {
		if i > 0 {
			fmt.Fprintln(&b)
		}
		vars := "(none)"
		if len(summary.Variables) > 0 {
			vars = strings.Join(summary.Variables, ", ")
		}
		inputs := "(none)"
		if len(summary.InputNames) > 0 {
			inputs = strings.Join(summary.InputNames, ", ")
		}
		repoAssets := "(none)"
		if len(summary.RepoAssets) > 0 {
			repoAssets = strings.Join(summary.RepoAssets, ", ")
		}
		description := summary.Description
		if description == "" {
			description = "(none)"
		}

		fmt.Fprintf(&b, "%s [%s]\n", summary.Name, summary.GovernanceTier)
		fmt.Fprintf(&b, "  desc: %s\n", description)
		fmt.Fprintf(&b, "  counts: files=%d templates=%d repo_files=%d manifest=v%d\n", summary.FileCount, summary.TemplateCount, summary.RepoFileCount, summary.ManifestVersion)
		fmt.Fprintf(&b, "  inputs: %s\n", inputs)
		fmt.Fprintf(&b, "  vars: %s\n", vars)
		fmt.Fprintf(&b, "  repo: %s\n", repoAssets)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeTableTextSummaries(w io.Writer, summaries []catalog.Summary) error {
	headers := []string{"NAME", "GOVERNANCE", "REPO_FILES", "FILES", "TEMPLATES", "INPUTS", "REPO_ASSETS"}
	rows := make([][]string, 0, len(summaries))
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}

	for _, summary := range summaries {
		inputs := "-"
		if len(summary.InputNames) > 0 {
			inputs = strings.Join(summary.InputNames, ",")
		}
		repoAssets := "-"
		if len(summary.RepoAssets) > 0 {
			repoAssets = strings.Join(summary.RepoAssets, ",")
		}
		row := []string{
			summary.Name,
			summary.GovernanceTier,
			strconv.Itoa(summary.RepoFileCount),
			strconv.Itoa(summary.FileCount),
			strconv.Itoa(summary.TemplateCount),
			inputs,
			repoAssets,
		}
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
		rows = append(rows, row)
	}

	var b strings.Builder
	writeTableRow(&b, headers, widths)
	separator := make([]string, len(headers))
	for i, width := range widths {
		separator[i] = strings.Repeat("-", width)
	}
	writeTableRow(&b, separator, widths)
	for _, row := range rows {
		writeTableRow(&b, row, widths)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func writeTextInspection(w io.Writer, inspection catalog.Inspection) error {
	var b strings.Builder
	vars := "(none)"
	if len(inspection.Variables) > 0 {
		vars = strings.Join(inspection.Variables, ", ")
	}
	inputs := "(none)"
	if len(inspection.Inputs) > 0 {
		parts := make([]string, 0, len(inspection.Inputs))
		for _, input := range inspection.Inputs {
			parts = append(parts, fmt.Sprintf("%s->%s", input.Name, input.TemplateVar))
		}
		inputs = strings.Join(parts, ", ")
	}
	description := inspection.Description
	if description == "" {
		description = "(none)"
	}
	repoAssets := "(none)"
	if len(inspection.RepoAssets) > 0 {
		repoAssets = strings.Join(inspection.RepoAssets, ", ")
	}

	fmt.Fprintf(&b, "name: %s\n", inspection.Name)
	fmt.Fprintf(&b, "description: %s\n", description)
	fmt.Fprintf(&b, "manifest_version: %d\n", inspection.ManifestVersion)
	fmt.Fprintf(&b, "inputs: %s\n", inputs)
	fmt.Fprintf(&b, "files: %d\n", inspection.FileCount)
	fmt.Fprintf(&b, "templates: %d\n", inspection.TemplateCount)
	fmt.Fprintf(&b, "variables: %s\n", vars)
	fmt.Fprintf(&b, "repo_assets: %s\n", repoAssets)
	fmt.Fprintf(&b, "mode: %s\n", inspection.Mode)
	fmt.Fprintf(&b, "shown: %d\n", inspection.ShownCount())
	writeTextInspectionSection(&b, "repo_files", inspection.RepoFiles())
	writeTextInspectionSection(&b, "language_files", inspection.LanguageFiles())

	_, err := io.WriteString(w, b.String())
	return err
}

func writeCompactTextInspection(w io.Writer, inspection catalog.Inspection) error {
	var b strings.Builder
	vars := "(none)"
	if len(inspection.Variables) > 0 {
		vars = strings.Join(inspection.Variables, ", ")
	}
	inputs := "(none)"
	if len(inspection.Inputs) > 0 {
		parts := make([]string, 0, len(inspection.Inputs))
		for _, input := range inspection.Inputs {
			parts = append(parts, fmt.Sprintf("%s->%s", input.Name, input.TemplateVar))
		}
		inputs = strings.Join(parts, ", ")
	}
	description := inspection.Description
	if description == "" {
		description = "(none)"
	}
	repoAssets := "(none)"
	if len(inspection.RepoAssets) > 0 {
		repoAssets = strings.Join(inspection.RepoAssets, ", ")
	}

	fmt.Fprintf(&b, "%s — %s\n", inspection.Name, description)
	fmt.Fprintf(&b, "  manifest: v%d | mode: %s | shown: %d | files: %d | templates: %d\n", inspection.ManifestVersion, inspection.Mode, inspection.ShownCount(), inspection.FileCount, inspection.TemplateCount)
	fmt.Fprintf(&b, "  inputs: %s\n", inputs)
	fmt.Fprintf(&b, "  vars: %s\n", vars)
	fmt.Fprintf(&b, "  repo_assets: %s\n", repoAssets)
	writeCompactTextInspectionSection(&b, "repo_files", inspection.RepoFiles())
	writeCompactTextInspectionSection(&b, "language_files", inspection.LanguageFiles())

	_, err := io.WriteString(w, b.String())
	return err
}

func writeTableRow(b *strings.Builder, cells []string, widths []int) {
	for i, cell := range cells {
		if i > 0 {
			b.WriteString("  ")
		}
		fmt.Fprintf(b, "%-*s", widths[i], cell)
	}
	b.WriteByte('\n')
}

func writeTextInspectionSection(b *strings.Builder, title string, files []catalog.InspectionFile) {
	fmt.Fprintf(b, "%s:\n", title)
	if len(files) == 0 {
		fmt.Fprintln(b, "- (none)")
		return
	}
	for _, file := range files {
		fmt.Fprintf(b, "- %s -> %s (%s)\n", file.Source, file.Output, file.Action)
	}
}

func writeCompactTextInspectionSection(b *strings.Builder, title string, files []catalog.InspectionFile) {
	if len(files) == 0 {
		fmt.Fprintf(b, "  %s: (none)\n", title)
		return
	}
	items := make([]string, 0, len(files))
	for _, file := range files {
		items = append(items, fmt.Sprintf("%s -> %s (%s)", file.Source, file.Output, file.Action))
	}
	fmt.Fprintf(b, "  %s:\n", title)
	for _, item := range items {
		fmt.Fprintf(b, "    - %s\n", item)
	}
}
