package utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const tarBlockSize = 512

var (
	ErrTarPrepareToAppend       = errors.New("failed preparing to append")
	ErrInitializingEmptyTarFile = errors.New("failed initializing empty tar file")
)

func CompressToTarGz(srcDir string, tarFile io.Writer) error {
	gw := gzip.NewWriter(tarFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// walk through every file in the folder
	err := filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// generate tar header
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(srcDir, file)
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

func DecompressTarGz(tarFile io.Reader, destDir string) error {
	log.Debugf("Decompressing tar file to %s", destDir)
	gr, err := gzip.NewReader(tarFile)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		target := filepath.Join(destDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			targetInfo, err := os.Stat(target)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					err = os.MkdirAll(target, 0o755)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			} else if !targetInfo.IsDir() {
				return fmt.Errorf("cannot decompress tar file: %s is not a directory", target)
			}
		case tar.TypeReg:
			targetDir := filepath.Dir(target)
			err = os.MkdirAll(targetDir, 0o755)
			if err != nil {
				return err
			}
			targetF, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer func() {
				closeErr := targetF.Close()
				if err == nil {
					err = closeErr
				}
			}()
			_, err = io.Copy(targetF, tr)
			if err != nil {
				return err
			}
		}
	}
}

// TarInit creates an empty tar file. The tar file is created with 2 empty blocks
// because 2 empty blocks denote the end of the tar file following the specification
// https://www.gnu.org/software/tar/manual/html_node/Standard.html. It is not required
// by the specification, but it is a common practice.
func TarInit(fs afero.Fs, path string) error {
	tarFile, err := fs.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	n, err := tarFile.Write(make([]byte, 2*tarBlockSize))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInitializingEmptyTarFile, err)
	}
	if n != 2*tarBlockSize {
		return fmt.Errorf("%w: %s", ErrInitializingEmptyTarFile, path)
	}
	return nil
}

// TarPrepareToAppend prepares a tar file for appending new files. The tar file
// should have 2 empty blocks (1024 bytes) at the end of the file following the
// specification https://www.gnu.org/software/tar/manual/html_node/Standard.html.
// Removes the last 2 blocks (1024 bytes) if they are all 0. Returns an error if
// the tar file is not empty or if the last 1024 bytes are not all 0.
func TarPrepareToAppend(tarFile afero.File) error {
	// Prepare tar for append
	stats, err := tarFile.Stat()
	if err != nil {
		return err
	}
	if stats.Size() == 0 {
		return nil
	}
	if stats.Size() < 2*tarBlockSize {
		return fmt.Errorf("%w: tar file is not empty but has less than 2 blocks (1024 bytes)", ErrTarPrepareToAppend)
	}

	// Check if the last 1024 bytes are all 0
	d := make([]byte, 1024)
	n, err := tarFile.ReadAt(d, stats.Size()-1024)
	if err != nil {
		return err
	}
	if n != 1024 {
		return fmt.Errorf("%w: read %d bytes instead of 1024", ErrTarPrepareToAppend, n)
	}
	for _, b := range d {
		if b != 0 {
			return fmt.Errorf("%w: last 1024 bytes are not all 0", ErrTarPrepareToAppend)
		}
	}
	// Seek last 1024 bytes
	_, err = tarFile.Seek(-1024, io.SeekEnd)
	return err
}

// TarAddDir add a directory to a tar file. The directory is added with a prefix
// path
func TarAddDir(srcPath, prefix string, tarFile io.Writer) error {
	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()
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

		fileRelPath, err := filepath.Rel(srcPath, file)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(prefix, fileRelPath)

		// write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// if not a dir, write file content
		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
			err = data.Close()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
