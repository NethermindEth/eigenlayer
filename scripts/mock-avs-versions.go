package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	fileName  = "/tmp/mock-avs-versions.yml"
	cacheFile = "/tmp/mock-avs-versions-cache.yml"
)

type MockAVS struct {
	Repo       string `yml:"repo"`
	Version    string `yml:"version"`
	CommitHash string `yml:"commitHash"`
}

type Repos struct {
	Repos []MockAVS `yml:"repos"`
}

func main() {
	repos := []string{
		"https://github.com/NethermindEth/mock-avs",
		"https://github.com/NethermindEth/mock-avs-pkg",
	}

	cache, err := readFromCache()
	if err == nil && time.Since(cache.Timestamp).Hours() < 1 {
		fmt.Println("Using cached data:", cache.Data)
		return
	}

	data := make([]MockAVS, 0)
	if shouldUpdateFile(fileName) {
		for _, repo := range repos {
			tag, commitHash, err := latestGitTagAndCommitHash(repo)
			if err != nil {
				fmt.Println("Error fetching latest git tag:", err)
				return
			}

			data = append(data, MockAVS{
				Repo:       repo,
				Version:    tag,
				CommitHash: commitHash,
			})
		}

		err = writeYMLFile(fileName, Repos{Repos: data})
		if err != nil {
			fmt.Println("Error writing to yml file:", err)
			return
		}

		// Update the cache
		err = writeToCache(Repos{Repos: data})
		if err != nil {
			fmt.Println("Error updating cache:", err)
		}
	}
}

func shouldUpdateFile(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return true
	}

	modifiedTime := info.ModTime()
	currentTime := time.Now()

	return currentTime.Sub(modifiedTime).Hours() > 1
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

func writeYMLFile(filePath string, data Repos) error {
	ymlContent, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, ymlContent, 0o644)
	if err != nil {
		return err
	}

	return nil
}

type CacheData struct {
	Timestamp time.Time `json:"timestamp"`
	Data      Repos     `json:"data"`
}

func readFromCache() (CacheData, error) {
	var cache CacheData

	file, err := os.ReadFile(cacheFile)
	if err != nil {
		return cache, err
	}

	err = json.Unmarshal(file, &cache)
	return cache, err
}

func writeToCache(data Repos) error {
	cache := CacheData{
		Timestamp: time.Now(),
		Data:      data,
	}

	jsonData, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, jsonData, 0o644)
}
