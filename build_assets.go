package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func copyStaticAssets() error {
	subdirs := []string{"css", "js", "images"}
	for _, sd := range subdirs {
		src := filepath.Join("markdown", sd)
		dest := filepath.Join("docs", sd)
		if err := copyDir(src, dest); err != nil {
			// skip if doesn't exist
			var fsErr *fs.PathError
			if errors.Is(err, fs.ErrNotExist) || strings.Contains(err.Error(), "no such file") || errors.As(err, &fsErr) {
				continue
			}
			return err
		}
	}
	return nil
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(p, target)
	})
}

func copyFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
