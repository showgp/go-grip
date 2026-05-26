package cmd

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/showgp/go-grip/internal"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-grip [file]",
	Short: "Render markdown document as html",
	Args:  cobra.MatchAll(cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		exportFile, _ := cmd.Flags().GetString("export")
		outputFile, _ := cmd.Flags().GetString("output")

		if exportFile != "" {
			data, err := os.ReadFile(exportFile)
			if err != nil {
				return fmt.Errorf("read %q: %w", exportFile, err)
			}

			parser := internal.NewParser()
			rendered, err := parser.Render(data)
			if err != nil {
				return fmt.Errorf("render %q: %w", exportFile, err)
			}

			// Use the export file's directory as root for image resolution
			exportDir := filepath.Dir(exportFile)
			if exportDir == "." {
				absDir, err := filepath.Abs(".")
				if err == nil {
					exportDir = absDir
				}
			} else {
				absDir, err := filepath.Abs(exportDir)
				if err == nil {
					exportDir = absDir
				}
			}

			htmlContent, err := internal.BuildExportHTML(template.HTML(rendered.Content), false, exportDir)
			if err != nil {
				return fmt.Errorf("build export HTML: %w", err)
			}

			if outputFile != "" {
				if err := os.WriteFile(outputFile, []byte(htmlContent), 0o644); err != nil {
					return fmt.Errorf("write %q: %w", outputFile, err)
				}
			} else {
				fmt.Print(htmlContent)
			}
			return nil
		}

		browser, _ := cmd.Flags().GetBool("browser")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		boundingBox, _ := cmd.Flags().GetBool("bounding-box")
		noReload, _ := cmd.Flags().GetBool("no-reload")
		recursive, _ := cmd.Flags().GetBool("recursive")

		var file string
		if len(args) == 1 {
			file = args[0]
		}

		parser := internal.NewParser()
		server := internal.NewServerWithOptions(internal.ServerOptions{
			Host:         host,
			Port:         port,
			BoundingBox:  boundingBox,
			Browser:      browser,
			EnableReload: !noReload,
			StrictPort:   cmd.Flags().Changed("port"),
			Recursive:    recursive,
			Parser:       parser,
		})
		return server.Serve(file)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("browser", "b", true, "Open new browser tab")
	rootCmd.Flags().StringP("host", "H", "localhost", "Host to use")
	rootCmd.Flags().IntP("port", "p", 6419, "Port to use")
	rootCmd.Flags().Bool("bounding-box", true, "Add bounding box to HTML")
	rootCmd.Flags().Bool("no-reload", false, "Disable automatic browser reload on file changes")
	rootCmd.Flags().BoolP("recursive", "r", false, "Include nested Markdown files in directory sidebar")
	rootCmd.Flags().String("export", "", "Export a Markdown file as standalone HTML and exit")
	rootCmd.Flags().String("output", "", "Output file path (used with --export; default: stdout)")
}
