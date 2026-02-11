package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tirthpatell/mdr/internal/markdown"
	"github.com/tirthpatell/mdr/internal/viewer"
)

var viewRaw bool

var viewCmd = &cobra.Command{
	Use:   "view [file]",
	Short: "Render and view a markdown file in the terminal",
	Long:  "Renders a markdown file with rich formatting. Reads from stdin if no file is provided.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var rendered string
		var err error

		if len(args) == 0 {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("no file provided and nothing on stdin")
			}
			rendered, err = markdown.RenderFromReader(os.Stdin)
		} else {
			rendered, err = markdown.RenderFile(args[0])
		}
		if err != nil {
			return err
		}

		if viewRaw {
			fmt.Print(rendered)
			return nil
		}

		m := viewer.NewModel(rendered)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	viewCmd.Flags().BoolVar(&viewRaw, "raw", false, "Print rendered output without TUI (useful for piping)")
	rootCmd.AddCommand(viewCmd)

	// Make `mdr <file>` work as shorthand for `mdr view <file>`
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			info, err := os.Stat(args[0])
			if err == nil && !info.IsDir() {
				return viewCmd.RunE(viewCmd, args)
			}
		}
		// Check stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			rendered, err := markdown.RenderFromReader(os.Stdin)
			if err != nil {
				return err
			}
			fmt.Print(rendered)
			return nil
		}
		return cmd.Help()
	}
	rootCmd.Args = cobra.ArbitraryArgs
}
