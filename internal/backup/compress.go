package backup

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// GzipFile compresses src into dst using gzip.
func GzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src for gzip: %w", err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst for gzip: %w", err)
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	// Optional: give the original name to the gzip header
	gw.Name = src

	if _, err := io.Copy(gw, in); err != nil {
		gw.Close()
		return fmt.Errorf("copy to gzip writer: %w", err)
	}

	if err := gw.Close(); err != nil {
		return fmt.Errorf("close gzip writer: %w", err)
	}

	return nil
}
