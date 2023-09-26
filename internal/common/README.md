# `common` package

This package should not contain any other packages inside it. All code should be in the root and grouped by functionality in separate files. If you feel that you need to create a new package for common code, it should probably be at the root of the `internal` directory.

## `mock-avs.go` guide

The `mock-avs.go` file provides utilities to interact with two GitHub repositories (the mock-avs and mock-avs-pkg repositories), fetch their latest versions, and cache this data locally. The primary goal is to retrieve the latest git tag and commit hash from the repositories and store them in a YAML file. If the data is older than an hour, it updates the file; otherwise, it uses the cached data. The current solution aims to avoid unnecessary GitHub API calls and reduce the API rate limiting impact.

The YAML files are the following:

- `/tmp/mock-avs-versions.yml`: Stores the latest versions of the repositories. This file can be manually updated to force the code to use custom versions. The code will not update this file if it is older than an hour.
- `/tmp/mock-avs-versions-cache.json`: Stores the latest versions of the repositories and the timestamp of the last update. This file is used to avoid unnecessary GitHub API calls. Must not be manually updated.

To enforce an update on those files, you can delete them and run the code again/run the `/scripts/mock-avs-versions.go` script (`go run scripts/mock-avs-versions.go`).

### Key Components:
#### Constants:
- `dataFile`: Path to the YAML file storing the latest versions of the repositories.
- `cacheFile`: Path to the cache file.
- Repository URLs for mock-avs and mock-avs-pkg.
- 
#### Data Structures:
- `Repos`: Contains a list of MockAVSData which holds information about a repository.
- `MockAVS`: Represents a GitHub repository with its URL, latest version, and commit hash.
- `MockAVSImage`: Represents a Docker image with its name and tag.

#### Functions:
- `SetMockAVSs()`: Initializes the data structures with the latest versions of the repositories. This data structures can be imported from the entire codebase, especially the tests.
- `checkCache()`: Checks if the cache is valid (less than an hour old). If not, it fetches the latest data from GitHub.
- `shouldUpdateFile()`: Determines if the `dataFile` should be updated based on its last modification time.
- `latestGitTagAndCommitHash()`: Fetches the latest git tag and commit hash from a GitHub repository.
- `writeYMLFile()`: Writes the repository data to a YAML file.
- `readFromCache()`: Reads the cached data.
- `writeToCache()`: Writes data to the cache.

### Usage:

The package is designed to be used by calling the `SetMockAVSs()` function, which sets up the data structures with the latest versions of the repositories. The data is then written to the `mock-avs-versions.yml` file. If the data is older than an hour, it fetches the latest data from GitHub and updates the file.