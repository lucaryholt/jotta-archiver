package archive

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

// Archive starts the archiving process and returns the upload ID
func Archive(ctx context.Context, folderPath, remotePath, archiveName string, debug bool) (string, error) {
	fullRemote := fmt.Sprintf("%s/%s", remotePath, archiveName)
	cmd := exec.CommandContext(ctx, "jotta-cli", "archive", folderPath, fmt.Sprintf("--remote=%s", fullRemote))
	
	if debug {
		fmt.Printf("[DEBUG] Running: jotta-cli archive %s --remote=%s\n", folderPath, fullRemote)
	}
	
	// Always capture output to get the upload ID
	output, err := cmd.CombinedOutput()
	if err != nil {
		if debug {
			fmt.Printf("[DEBUG] Archive command error: %v\n", err)
			fmt.Printf("[DEBUG] Output: %s\n", string(output))
		}
		return "", fmt.Errorf("failed to start archive command: %w\nOutput: %s", err, string(output))
	}
	
	if debug {
		fmt.Printf("[DEBUG] Archive command output: %s\n", string(output))
	}
	
	// Extract upload ID from output
	// Looking for line like: "started upload 434e709656"
	uploadID := extractUploadID(string(output))
	if uploadID == "" {
		return "", fmt.Errorf("failed to extract upload ID from output: %s", string(output))
	}
	
	if debug {
		fmt.Printf("[DEBUG] Extracted upload ID: %s\n", uploadID)
	}
	
	return uploadID, nil
}

// extractUploadID extracts the upload ID from jotta-cli archive output
func extractUploadID(output string) string {
	// Look for "started upload <id>"
	re := regexp.MustCompile(`started upload ([a-f0-9]+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Observe launches the jotta-cli observe command to monitor upload progress
func Observe(uploadID string, debug bool) error {
	if debug {
		fmt.Printf("[DEBUG] Running: jotta-cli observe --uploadid=%s\n", uploadID)
	}
	
	cmd := exec.Command("jotta-cli", "observe", fmt.Sprintf("--uploadid=%s", uploadID))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

