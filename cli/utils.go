package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
)

func validatePkgURL(urlStr string) error {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("%w: %s", ErrInvalidURL, "URL must be HTTP or HTTPS")
	}
	return nil
}

func buildTar(srcPath string, dst *os.File) error {
	gw := gzip.NewWriter(dst)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// walk through every file in the folder
	err := filepath.Walk(srcPath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(srcPath, file)
		if err != nil {
			return err
		}

		// write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
