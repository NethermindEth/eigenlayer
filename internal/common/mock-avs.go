package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

const (
	mockAVSRepo             = "https://github.com/NethermindEth/mock-avs"
	mockAVSPkgRepo          = "https://github.com/NethermindEth/mock-avs-pkg"
	optionReturnerImageName = "mock-avs-option-returner"
	healthCheckerImageName  = "mock-avs-health-checker"
	pluginImageName         = "mock-avs-plugin"
)

var (
	MockAvsSrc          MockAVS
	MockAvsPkg          MockAVS
	OptionReturnerImage MockAVSImage
	HealthCheckerImage  MockAVSImage
	PluginImage         MockAVSImage
)

type MockAVS struct {
	repo       string
	version    string
	commitHash string
}

func NewMockAVS(repo string, version string, commitHash string) *MockAVS {
	return &MockAVS{
		repo:       repo,
		version:    version,
		commitHash: commitHash,
	}
}

func (m *MockAVS) Repo() string {
	return m.repo
}

func (m *MockAVS) Version() string {
	return m.version
}

func (m *MockAVS) CommitHash() string {
	return m.commitHash
}

type MockAVSImage struct {
	image string
	tag   string
}

func NewMockAVSImage(image, tag string) *MockAVSImage {
	return &MockAVSImage{
		image: image,
		tag:   tag,
	}
}

func (m *MockAVSImage) Image() string {
	return m.image
}

func (m *MockAVSImage) Tag() string {
	return m.tag
}

func (m *MockAVSImage) FullImage() string {
	return fmt.Sprintf("%s:%s", m.image, m.tag)
}

type Tag struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit"`
}

func latestGitTagAndCommitHash(repoURL string) (string, string, error) {
	// Extract the repo owner and name from the URL
	parts := strings.Split(strings.TrimRight(strings.TrimPrefix(repoURL, "https://github.com/"), "/"), "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo URL")
	}
	owner, repo := parts[0], parts[1]

	// GitHub API endpoint to get tags
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)

	var tag, commitHash string

	operation := func() error {
		resp, err := http.Get(apiURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to fetch data from GitHub API, status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var tags []Tag
		if err := json.Unmarshal(body, &tags); err != nil {
			return err
		}

		if len(tags) == 0 {
			return fmt.Errorf("no tags found in the repository")
		}

		log.Debugf("Latest Tag: %s\nCommit Hash: %s\n", tags[0].Name, tags[0].Commit.Sha)
		tag = tags[0].Name
		commitHash = tags[0].Commit.Sha
		return nil
	}

	// Using exponential backoff for retries
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 5 * time.Second
	if err := backoff.Retry(operation, bo); err != nil {
		return "", "", err
	}

	return tag, commitHash, nil
}

// SetMockAVSs set up the MockAVS and MockAVSPkg data structures with
// the latest versions of the mock-avs and mock-avs-pkg repositories.
// It also sets up the OptionReturnerImage, HealthCheckerImage and
// PluginImage data structures using as tag the latest version of the
// mock-avs repository.
func SetMockAVSs() error {
	repos := []string{mockAVSRepo, mockAVSPkgRepo}
	for _, repo := range repos {
		tag, commitHash, err := latestGitTagAndCommitHash(repo)
		if err != nil {
			return err
		}
		if repo == mockAVSRepo {
			MockAvsSrc = *NewMockAVS(repo, tag, commitHash)
		} else {
			MockAvsPkg = *NewMockAVS(repo, tag, commitHash)
		}
	}

	OptionReturnerImage = *NewMockAVSImage(optionReturnerImageName, MockAvsSrc.Version())
	HealthCheckerImage = *NewMockAVSImage(healthCheckerImageName, MockAvsSrc.Version())
	PluginImage = *NewMockAVSImage(pluginImageName, MockAvsSrc.Version())

	return nil
}
