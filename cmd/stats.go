package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/tirthpatell/mdr/internal/stats"
)

var statsCmd = &cobra.Command{
	Use:   "stats [file]",
	Short: "Show word count and document statistics for a markdown file",
	Long:  "Displays word count, line count, heading count, link count, and estimated reading time. Reads from stdin if no file is provided.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var s stats.Stats
		var name string
		var err error

		if len(args) == 0 || args[0] == "-" {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) != 0 {
				return fmt.Errorf("no file provided and nothing on stdin")
			}
			s, err = stats.FromReader(os.Stdin)
			name = "<stdin>"
		} else {
			s, err = stats.FromFile(args[0])
			name = args[0]
		}
		if err != nil {
			return err
		}

		titleStyle := lipgloss.NewStyle().Bold(true)
		fmt.Println(titleStyle.Render(name))
		fmt.Println(s.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
