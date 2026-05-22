package preprocess

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Prepare builds a temp directory with files organized and/or renamed according
// to the provided options. If neither splitByFormat nor extensionRenames are
// active, the original srcDir is returned unchanged with a no-op cleanup.
//
// When splitByFormat is true, each file is placed under a format subfolder
// named after its (uppercased) extension, inserted as the immediate parent of
// the file while preserving the surrounding directory hierarchy:
//
//	src/vacation/IMG_001.jpg  →  tempDir/vacation/JPEG/IMG_001.jpg
//	src/vacation/IMG_001.HIF  →  tempDir/vacation/HEIF/IMG_001.HEIF  (with HIF→HEIF rename)
//
// extensionRenames keys are matched case-insensitively.
func Prepare(srcDir string, splitByFormat bool, extensionRenames map[string]string) (dir string, cleanup func(), err error) {
	noop := func() {}

	needsProcessing := splitByFormat || len(extensionRenames) > 0
	if !needsProcessing {
		return srcDir, noop, nil
	}

	tempDir, err := os.MkdirTemp("", "jotta-archiver-*")
	if err != nil {
		return "", noop, err
	}

	cleanupFn := func() {
		os.RemoveAll(tempDir)
	}

	// Build an uppercase lookup map for extension renames.
	renames := make(map[string]string, len(extensionRenames))
	for from, to := range extensionRenames {
		renames[strings.ToUpper(from)] = to
	}

	err = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		relDir := filepath.Dir(rel)
		fileName := filepath.Base(rel)
		ext := filepath.Ext(fileName)
		nameNoExt := strings.TrimSuffix(fileName, ext)

		// Apply extension rename (case-insensitive match).
		extUpper := strings.ToUpper(strings.TrimPrefix(ext, "."))
		if renamed, ok := renames[extUpper]; ok {
			ext = "." + renamed
		}

		renamedFileName := nameNoExt + ext

		var destDir string
		if splitByFormat {
			// Insert format folder as immediate parent of the file.
			formatFolder := strings.ToUpper(strings.TrimPrefix(ext, "."))
			if relDir == "." {
				destDir = filepath.Join(tempDir, formatFolder)
			} else {
				destDir = filepath.Join(tempDir, relDir, formatFolder)
			}
		} else {
			if relDir == "." {
				destDir = tempDir
			} else {
				destDir = filepath.Join(tempDir, relDir)
			}
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		dest := filepath.Join(destDir, renamedFileName)
		return linkOrCopy(path, dest)
	})

	if err != nil {
		cleanupFn()
		return "", noop, err
	}

	return tempDir, cleanupFn, nil
}

// linkOrCopy attempts a hard link from src to dst, falling back to a byte copy
// when a hard link is not possible (e.g. cross-device).
func linkOrCopy(src, dst string) error {
	if err := os.Link(src, dst); err == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
