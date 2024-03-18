package utils

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// CopyFile copies a file.
func CopyFile(src string, dst string) (n int64, err error) {
	if src == dst {
		return
	}

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	n, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir coyies a dir.
func CopyDir(src string, dst string) (err error) {
	src = CleanPath(src)
	dst = CleanPath(dst)
	if src == dst {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.New("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return errors.New("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := path.Join(src, entry.Name())
		dstPath := path.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Type().Perm()&os.ModeSymlink != 0 {
				continue
			}

			_, err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// ZipTo compresses the path to the io.Writer.
func ZipTo(path string, output io.Writer) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		absDir, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		archive := zip.NewWriter(output)
		defer archive.Close()

		return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Name = strings.TrimPrefix(strings.TrimPrefix(path, absDir), "/")
			if header.Name == "" {
				return nil
			}

			if info.IsDir() {
				header.Name += "/"
			} else {
				header.Method = zip.Deflate
			}

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})
	}

	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	archive := zip.NewWriter(output)
	defer archive.Close()

	gzw, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(gzw, file)
	return err
}
