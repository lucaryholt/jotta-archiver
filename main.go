package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/luca/jotta-archiver/archive"
	"github.com/luca/jotta-archiver/config"
	"github.com/luca/jotta-archiver/preprocess"
	"github.com/luca/jotta-archiver/tui"
	"github.com/luca/jotta-archiver/wordgen"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check if jotta-cli is installed
	if err := checkJottaCLI(); err != nil {
		return err
	}

	// Parse arguments
	if len(os.Args) < 2 {
		printUsage()
		return fmt.Errorf("missing folder argument")
	}

	folderPath := os.Args[1]

	// Validate folder exists
	info, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("folder does not exist: %s", folderPath)
		}
		return fmt.Errorf("failed to access folder: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", folderPath)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		configPath, _ := config.GetConfigPath()
		return fmt.Errorf("failed to load config from %s: %w", configPath, err)
	}

	if len(cfg.Presets) == 0 {
		return fmt.Errorf("no presets defined in configuration")
	}

	// Generate archive name
	archiveName := wordgen.Generate()

	// Collect user selections via TUI
	result, err := tui.Run(cfg.Presets, archiveName, absPath)
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	if result.Cancelled {
		return nil
	}

	// Preprocess: reorganize/rename files into a temp dir if needed
	uploadDir, cleanup, err := preprocess.Prepare(absPath, result.SplitByFormat, cfg.ExtensionRenames)
	if err != nil {
		return fmt.Errorf("failed to prepare upload directory: %w", err)
	}
	defer cleanup()

	// Start archive
	ctx := context.Background()
	uploadID, err := archive.Archive(ctx, uploadDir, result.RemotePath, result.ArchiveName, result.DebugMode)
	if err != nil {
		return fmt.Errorf("failed to start archive: %w", err)
	}

	// Monitor upload progress
	if err := archive.Observe(uploadID, result.DebugMode); err != nil {
		return fmt.Errorf("failed to observe upload: %w", err)
	}

	return nil
}

func checkJottaCLI() error {
	cmd := exec.Command("jotta-cli", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("jotta-cli not found or not working. Please ensure jotta-cli is installed and in your PATH")
	}
	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: jotta-archiver <folder>

Archive a folder using jotta-cli with an interactive TUI.

Arguments:
  <folder>    Path to the folder to archive

Configuration:
  The tool reads presets from ~/.jotta-archiver.yaml
  A default config will be created if none exists.

Example:
  jotta-archiver ~/Pictures/vacation_2025

`)
}

