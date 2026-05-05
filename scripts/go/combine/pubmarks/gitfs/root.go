package gitfs

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveDatasetsDir returns the datasets directory containing stocks/.
// Uses PUBMARKS_ROOT if set (expects <repoRoot>/datasets), else searches ancestors of cwd for datasets/stocks.
func ResolveDatasetsDir() (string, error) {
	if root := os.Getenv("PUBMARKS_ROOT"); root != "" {
		d := filepath.Join(root, "datasets")
		if err := stocksDirAccessible(d); err != nil {
			return "", fmt.Errorf("gitfs: PUBMARKS_ROOT datasets/stocks not found under %q: %w", d, err)
		}
		return d, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("gitfs: getwd: %w", err)
	}
	for {
		candidate := filepath.Join(dir, "datasets")
		if err := stocksDirAccessible(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf(`gitfs: could not find datasets/stocks (set PUBMARKS_ROOT or run from repoCheckout)`)
}

func stocksDirAccessible(datasetsRoot string) error {
	st := filepath.Join(datasetsRoot, "stocks")
	fi, err := os.Stat(st)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not a directory")
	}
	return nil
}
