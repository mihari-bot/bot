package bot

import (
	"os"
	"path/filepath"
)

func (b *Bot) pthResolve(parts ...string) (string, error) {
	relPath := filepath.Join(parts...)

	if filepath.IsAbs(relPath) {
		return relPath, nil
	}

	err := os.MkdirAll(b.baseDir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(b.baseDir, relPath), nil
}
