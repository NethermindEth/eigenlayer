package install

import "fmt"

// RepositoryNotFoundError is returned when the given repository URL is not found.
type RepositoryNotFoundError struct {
	URL string
}

func (e RepositoryNotFoundError) Error() string {
	return fmt.Sprintf("repository %s not found", e.URL)
}

// RepositoryNotFoundOrPrivateError is returned when the specified repository URL
// cannot be found or accessed due to its private status. This error typically occurs
// when no credentials are provided and the repository is either private or does not exist.
type RepositoryNotFoundOrPrivateError struct {
	URL string
}

func (e RepositoryNotFoundOrPrivateError) Error() string {
	return fmt.Sprintf("repository %s not found or private", e.URL)
}

// TagNotFoundError is returned when the specified git tag cannot be found.
type TagNotFoundError struct {
	Tag string
}

func (e TagNotFoundError) Error() string {
	return fmt.Sprintf("tag %s not found", e.Tag)
}
