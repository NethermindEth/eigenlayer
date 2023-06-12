package utils

import (
	"io"
	"os"
)

func Copy(from, to string) (err error) {
	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() {
		err = fromFile.Close()
	}()

	toFile, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		err = toFile.Close()
	}()

	_, err = io.Copy(toFile, fromFile)
	return err
}
