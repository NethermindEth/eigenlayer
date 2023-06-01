package commands

import (
	"io"
	"sync"
)

// goCopy concurrently copies data from a source Reader to a destination WriteCloser.
// It returns a channel for errors that might occur during the copy operation.
//
// The function takes a WaitGroup for synchronizing the completion of the copy operation,
// a destination WriteCloser to write data to, a source Reader to read data from, and
// a boolean flag indicating whether the destination is a standard input stream.
//
// If the destination is a standard input stream, goCopy will close it after the copy operation.
//
// The function adds to the WaitGroup before starting the copy operation, and calls Done on the WaitGroup
// after the copy operation is complete or if an error occurs.
//
// The error channel returned by the function is closed after the copy operation is complete, which can be used to check
// for errors that occurred during the operation.
func goCopy(wait *sync.WaitGroup, dst io.WriteCloser, src io.Reader, isStdin bool) <-chan error {
	errChan := make(chan error)
	wait.Add(1)
	go func() {
		if _, err := io.Copy(dst, src); err != nil {
			errChan <- err
			return
		}
		if isStdin {
			if err := dst.Close(); err != nil {
				errChan <- err
				return
			}
		}
		close(errChan)
		wait.Done()
	}()
	return errChan
}
