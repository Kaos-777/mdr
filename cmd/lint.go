package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/tirthpatell/mdr/internal/linter"
)

var lintCmd = &cobra.Command{
	Use:   "lint [files...]",
	Short: "Lint markdown files for structural issues",
	Long:  "Checks markdown files for heading hierarchy, duplicate headings, empty links, and other structural issues. Reads from stdin if no files are provided.",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		fileStyle := lipgloss.NewStyle().Bold(true)

		totalIssues := 0

		if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("no files provided and nothing on stdin")
			}
			issues, err := linter.LintReader(os.Stdin)
			if err != nil {
				return fmt.Errorf("error reading stdin: %w", err)
			}
			if len(issues) > 0 {
				fmt.Println(fileStyle.Render("<stdin>"))
				for _, issue := range issues {
					prefix := warnStyle.Render("warning")
					if issue.Severity == linter.SeverityError {
						prefix = errorStyle.Render("error")
					}
					fmt.Printf("  line %d: %s %s (%s)\n", issue.Line, prefix, issue.Message, issue.Rule)
				}
				fmt.Println()
				totalIssues += len(issues)
			}
		} else {
			for _, path := range args {
				issues, err := linter.LintFile(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error reading %s: %v\n", path, err)
					continue
				}
				if len(issues) == 0 {
					continue
				}

				fmt.Println(fileStyle.Render(path))
				for _, issue := range issues {
					prefix := warnStyle.Render("warning")
					if issue.Severity == linter.SeverityError {
						prefix = errorStyle.Render("error")
					}
					fmt.Printf("  line %d: %s %s (%s)\n", issue.Line, prefix, issue.Message, issue.Rule)
				}
				fmt.Println()
				totalIssues += len(issues)
			}
		}

		if totalIssues > 0 {
			return fmt.Errorf("found %d issue(s)", totalIssues)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
