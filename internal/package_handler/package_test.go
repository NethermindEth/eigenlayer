package package_handler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/package_handler/testdata"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPackageHandlerFromURL(t *testing.T) {
	type testCase struct {
		name       string
		path       string
		url        string
		pkgHandler *PackageHandler
		err        error
	}
	// TODO: add test case for private repository
	ts := []testCase{
		func() testCase {
			t.Helper()
			afs := afero.NewOsFs()
			path, err := afero.TempDir(afs, "", "test")
			require.NoError(t, err)

			return testCase{
				name: "valid package",
				path: path,
				url:  "https://github.com/NethermindEth/mock-avs-pkg",
				pkgHandler: &PackageHandler{
					path: path,
				},
				err: nil,
			}
		}(),
		func() testCase {
			t.Helper()
			afs := afero.NewOsFs()
			path, err := afero.TempDir(afs, "", "test")
			require.NoError(t, err)

			return testCase{
				name:       "invalid url",
				path:       path,
				url:        "https://github.com/NethermindEth/mock-avs-invalid",
				pkgHandler: nil,
				err: RepositoryNotFoundOrPrivateError{
					URL: "https://github.com/NethermindEth/mock-avs-invalid",
				},
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler, err := NewPackageHandlerFromURL(NewPackageHandlerOptions{
				Path:    tc.path,
				URL:     tc.url,
				GitAuth: nil,
			})
			assert.ErrorIs(t, err, tc.err)
			if err == nil {
				assert.Equal(t, tc.pkgHandler.path, pkgHandler.path)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	type testCase struct {
		name      string
		pkgFolder string
		err       error
	}
	ts := []testCase{
		func() testCase {
			return testCase{
				name:      "valid package",
				pkgFolder: setupPackage(t),
				err:       nil,
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			if err := exec.Command("rm", "-rf", filepath.Join(pkgFolder, "pkg")).Run(); err != nil {
				t.Fatal("error preparing the test: " + err.Error())
			}
			return testCase{
				name:      "pkg folder does not exist",
				pkgFolder: pkgFolder,
				err: PackageDirNotFoundError{
					dirRelativePath: "pkg",
					packagePath:     pkgFolder,
				},
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			if err := exec.Command("rm", "-rf", filepath.Join(pkgFolder, "checksum.txt")).Run(); err != nil {
				t.Fatal("error preparing the test: " + err.Error())
			}
			return testCase{
				name:      "checksum.txt file does not exist",
				pkgFolder: pkgFolder,
				err:       nil,
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			if err := exec.Command("rm", "-rf", filepath.Join(pkgFolder, "pkg", "manifest.yml")).Run(); err != nil {
				t.Fatal("error preparing the test: " + err.Error())
			}
			return testCase{
				name:      "missing file that is listed in checksum.txt produces ErrInvalidChecksum",
				pkgFolder: pkgFolder,
				err:       ErrInvalidChecksum,
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			targetFile := filepath.Join(pkgFolder, "pkg", "manifest.yml") // replace targetFile.txt with the file you want to modify

			file, err := os.OpenFile(targetFile, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				t.Fatal("error opening target file: " + err.Error())
			}
			defer file.Close()

			_, err = file.WriteString("\n")
			if err != nil {
				t.Fatal("error writing to target file: " + err.Error())
			}

			return testCase{
				name:      "invalid hash in the checksum.txt",
				pkgFolder: pkgFolder,
				err:       ErrInvalidChecksum,
			}
		}(),
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler := NewPackageHandler(tc.pkgFolder)
			err := pkgHandler.Check()
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func setupPackage(t *testing.T) string {
	t.Helper()
	pkgFolder := t.TempDir()

	mockTapRepo := "https://github.com/NethermindEth/mock-avs-pkg.git"
	tag := "v5.4.0"

	t.Logf("Cloning mock tap repo %s and tag %s into %s", mockTapRepo, tag, pkgFolder)

	if err := exec.Command("git", "clone", "--single-branch", "-b", tag, mockTapRepo, pkgFolder).Run(); err != nil {
		t.Fatal("error cloning the mock tap repo: " + err.Error())
	}
	return pkgFolder
}

func TestProfilesNames(t *testing.T) {
	afs := afero.NewOsFs()
	testDir, err := afero.TempDir(afs, "", "test")
	require.NoError(t, err)
	testdata.SetupDir(t, "manifests", testDir, afs)

	ts := []struct {
		name       string
		folderPath string
		profiles   []string
		wantError  bool
	}{
		{
			name:       "valid manifest with one",
			folderPath: "full-ok",
			profiles:   []string{"profile1"},
		},
		{
			name:       "valid manifest with multiple profiles",
			folderPath: "minimal",
			profiles:   []string{"profile1", "profile2"},
		},
		{
			name:       "invalid manifest",
			folderPath: "invalid-fields",
			profiles:   nil,
			wantError:  true,
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler := NewPackageHandler(filepath.Join(testDir, "manifests", tc.folderPath))
			profiles, err := pkgHandler.profilesNames()
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.profiles, profiles)
			}
		})
	}
}

func TestParseProfile(t *testing.T) {
	afs := afero.NewOsFs()
	testDir, err := afero.TempDir(afs, "", "test")
	require.NoError(t, err)
	testdata.SetupDir(t, "packages", testDir, afs)

	ts := []struct {
		name    string
		pkgPath string
		profile string
		err     error
	}{
		{
			name:    "valid profile",
			pkgPath: "good-profiles",
			profile: "ok",
		},
		{
			name:    "profile without options",
			pkgPath: "no-options",
			profile: "no-options",
		},
		{
			name:    "invalid yml file",
			pkgPath: "bad-profiles",
			profile: "invalid-yml",
			err:     ParsingProfileError{profileName: "invalid-yml"},
		},
		{
			name:    "no profile",
			pkgPath: "bad-profiles",
			profile: "no-profile",
			err:     ReadingProfileError{profileName: "no-profile"},
		},
		{
			name:    "invalid format",
			pkgPath: "bad-profiles",
			profile: "not-yml",
			err:     ReadingProfileError{profileName: "not-yml"},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler := NewPackageHandler(filepath.Join(testDir, "packages", tc.pkgPath))
			profile, err := pkgHandler.parseProfile(tc.profile)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
			}
		})
	}
}

func TestProfiles(t *testing.T) {
	afs := afero.NewOsFs()
	testDir, err := afero.TempDir(afs, "", "test")
	require.NoError(t, err)
	testdata.SetupDir(t, "packages", testDir, afs)

	ts := []struct {
		name    string
		pkgPath string
		want    []Profile
		err     error
	}{
		{
			name:    "good profiles",
			pkgPath: "good-profiles",
			want: []Profile{
				{
					Options: []Option{
						{
							Name:    "el-port",
							Target:  "PORT",
							Type:    "port",
							Default: "8080",
							Help:    "Port of the harbor bay crocodile in the horse window within upside Coca Cola",
						},
						{
							Name:   "graffiti",
							Target: "GRAFFITI",
							Type:   "str",
							Help:   "Graffiti code of Donatello tattoo in DevCon restroom while hanging out with a Bored Ape",
						},
					},
				},
				{
					Options: []Option{},
				},
			},
		},
		{
			name:    "bad profiles",
			pkgPath: "bad-profiles",
			want:    []Profile{},
			err:     ParsingProfileError{profileName: "invalid-yml"},
		},
		{
			name:    "no options",
			pkgPath: "no-options",
			want:    []Profile{},
			err:     InvalidConfError{message: "Invalid profile", missingFields: []string{"options"}},
		},
		{
			name:    "invalid profile",
			pkgPath: "invalid-profile",
			want:    []Profile{},
			err:     fmt.Errorf("Invalid profile: invalid options: %w: %w: invalid monitoring: %w", InvalidConfError{message: "Option #1 is invalid", invalidFields: []string{"options.default"}}, InvalidConfError{message: "Option #2 is invalid", missingFields: []string{"options.type", "options.help"}}, InvalidConfError{message: "Monitoring target #1 is invalid", missingFields: []string{"monitoring.targets.port", "monitoring.targets.path"}}),
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler := NewPackageHandler(filepath.Join(testDir, "packages", tc.pkgPath))
			profiles, err := pkgHandler.Profiles()
			if tc.err != nil {
				assert.ErrorContains(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				for i, profile := range profiles {
					assert.Equal(t, tc.want[i].Options, profile.Options)
				}
			}
		})
	}
}

func TestDotEnv(t *testing.T) {
	afs := afero.NewOsFs()
	testDir, err := afero.TempDir(afs, "", "test")
	require.NoError(t, err)
	testdata.SetupDir(t, "packages", testDir, afs)

	ts := []struct {
		name    string
		pkgPath string
		profile string
		want    map[string]string
		err     error
	}{
		{
			name:    "good dot env",
			pkgPath: "good-profiles",
			profile: "ok",
			want: map[string]string{
				"KEY1": "8000",
				"KEY2": "false",
				"KEY3": "\"foo\"",
			},
		},
		{
			name:    "empty dot env",
			pkgPath: "bad-profiles",
			profile: "no-profile",
			want:    map[string]string{},
		},
		{
			name:    "invalid dot env",
			pkgPath: "bad-profiles",
			profile: "invalid-yml",
			want: map[string]string{
				"$$$FACADE": "trueAvocado666%!",
			},
		},
		{
			name:    "no dot env",
			pkgPath: "bad-profiles",
			profile: "not-yml",
			err:     ReadingDotEnvError{profileName: "not-yml"},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler := NewPackageHandler(filepath.Join(testDir, "packages", tc.pkgPath))
			dotEnv, err := pkgHandler.DotEnv(tc.profile)
			if tc.err != nil {
				assert.ErrorContains(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tc.want, dotEnv)
			}
		})
	}
}

func TestVersions(t *testing.T) {
	type testCase struct {
		name     string
		gitTags  []string
		versions []string
		err      error
	}
	ts := []testCase{
		{
			name:     "all tags are valid versions",
			gitTags:  []string{"v0.0.0", "v0.2.0", "v4.3.0"},
			versions: []string{"v4.3.0", "v0.2.0", "v0.0.0"},
			err:      nil,
		},
		{
			name:     "no versions",
			gitTags:  []string{"v0.0", "some-tag", "0"},
			versions: nil,
			err:      ErrNoVersionsFound,
		},
		{
			name:     "some tags are valid versions",
			gitTags:  []string{"v0.0.0", "v0.2.0", "v4.3.0", "v0.0", "some-tag", "0"},
			versions: []string{"v4.3.0", "v0.2.0", "v0.0.0"},
			err:      nil,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			path := t.TempDir()

			// Add a readme file to create the first commit
			readmeFile, err := os.Create(filepath.Join(path, "readme.txt"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = readmeFile.WriteString("Test file for test " + tc.name)
			if err != nil {
				t.Fatal(err)
			}
			for _, cmd := range []*exec.Cmd{
				exec.Command("git", "-C", path, "init"),
				exec.Command("git", "-C", path, "add", "readme.txt"),
				exec.Command("git", "-C", path, "config", "user.name", "user"),
				exec.Command("git", "-C", path, "config", "user.email", "user@email.com"),
				exec.Command("git", "-C", path, "commit", "-m", "Initial commit"),
			} {
				err := cmd.Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			// Add tags
			for _, tag := range tc.gitTags {
				err = exec.Command("git", "-C", path, "tag", "-a", tag, "-m", "Version: "+tag).Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			pkgHandler := NewPackageHandler(path)
			versions, err := pkgHandler.Versions()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				assert.Len(t, versions, 0)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.versions, versions)
			}
		})
	}
}

func TestLatestVersion(t *testing.T) {
	type testCase struct {
		name          string
		gitTags       []string
		latestVersion string
		err           error
	}
	ts := []testCase{
		{
			name:          "all tags are valid versions",
			gitTags:       []string{"v0.0.0", "v0.2.0", "v4.3.0"},
			latestVersion: "v4.3.0",
			err:           nil,
		},
		{
			name:          "no versions",
			gitTags:       []string{"v0.0", "some-tag", "0"},
			latestVersion: "",
			err:           ErrNoVersionsFound,
		},
		{
			name:          "some tags are valid versions",
			gitTags:       []string{"v0.0.0", "v0.2.0", "v4.3.0", "v0.0", "some-tag", "0"},
			latestVersion: "v4.3.0",
			err:           nil,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			path := t.TempDir()

			// Add a readme file to create the first commit
			readmeFile, err := os.Create(filepath.Join(path, "readme.txt"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = readmeFile.WriteString("Test file for test " + tc.name)
			if err != nil {
				t.Fatal(err)
			}

			for _, cmd := range []*exec.Cmd{
				exec.Command("git", "-C", path, "init"),
				exec.Command("git", "-C", path, "add", "readme.txt"),
				exec.Command("git", "-C", path, "config", "user.name", "user"),
				exec.Command("git", "-C", path, "config", "user.email", "user@email.com"),
				exec.Command("git", "-C", path, "commit", "-m", "Initial commit"),
			} {
				err := cmd.Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			// Add tags
			for _, tag := range tc.gitTags {
				err = exec.Command("git", "-C", path, "tag", "-a", tag, "-m", "Version: "+tag).Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			pkgHandler := NewPackageHandler(path)
			latestVersion, err := pkgHandler.LatestVersion()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.latestVersion, latestVersion)
			}
		})
	}
}

func TestCheckoutVersion(t *testing.T) {
	ts := []struct {
		name       string
		versions   []string
		checkoutTo string
		err        error
	}{
		{
			name:       "checkout to existing version",
			versions:   []string{"v0.0.0", "v0.2.0", "v4.3.0"},
			checkoutTo: "v4.3.0",
			err:        nil,
		},
		{
			name:       "checkout to non-existing version",
			versions:   []string{"v1.0.0"},
			checkoutTo: "v2.0.0",
			err:        ErrNoVersionsFound,
		},
		{
			name:       "checkout to invalid version format",
			versions:   []string{"v1.0.0"},
			checkoutTo: "v1.0",
			err:        ErrInvalidVersion,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			path := t.TempDir()
			// Initialize git repo
			err := exec.Command("git", "-C", path, "init").Run()
			if err != nil {
				t.Fatal(err)
			}

			// Add version tags
			for i, tag := range tc.versions {
				file := fmt.Sprintf("readme-%d.txt", i)
				readmeFile, err := os.Create(filepath.Join(path, file))
				if err != nil {
					t.Fatal(err)
				}
				defer readmeFile.Close()
				_, err = readmeFile.WriteString("Test file for test " + tc.name)
				if err != nil {
					t.Fatal(err)
				}
				for _, cmd := range []*exec.Cmd{
					exec.Command("git", "-C", path, "add", file),
					exec.Command("git", "-C", path, "config", "user.name", "user"),
					exec.Command("git", "-C", path, "config", "user.email", "user@email.com"),
					exec.Command("git", "-C", path, "commit", "-m", fmt.Sprintf("Commit %d", i)),
					exec.Command("git", "-C", path, "tag", "-a", tag, "-m", "Version: "+tag),
				} {
					err := cmd.Run()
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			pkgHandler := NewPackageHandler(path)
			err = pkgHandler.CheckoutVersion(tc.checkoutTo)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				err = exec.Command("git", "-C", path, "describe", "--exact-match", "--tags", tc.checkoutTo).Run()
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckoutCommit(t *testing.T) {
	testDir := t.TempDir()
	require.NoError(t, initGitRepo(testDir), "error initializing git repo")
	var commitHashes []string

	for i := 0; i < 3; i++ {
		f, err := os.Create(filepath.Join(testDir, fmt.Sprintf("file-%d.txt", i)))
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(fmt.Sprintf("Test file %d", i))
		require.NoError(t, err)
		require.NoError(t, stageAll(testDir), "error staging file %d", i)
		commitHash, err := commit(testDir, fmt.Sprintf("Commit %d", i))
		require.NoError(t, err)
		commitHashes = append(commitHashes, commitHash)
	}

	pkgHandler := NewPackageHandler(testDir)

	for i, commitHash := range commitHashes {
		err := pkgHandler.CheckoutCommit(commitHash)
		require.NoError(t, err)
		headHash, err := headCommitHash(testDir)
		require.NoError(t, err)
		assert.Equal(t, commitHash, headHash, "checkout to commit %d failed", i)
	}
}

func TestCurrentCommitHash(t *testing.T) {
	testDir := t.TempDir()
	require.NoError(t, initGitRepo(testDir), "error initializing git repo")
	pkgHandler := NewPackageHandler(testDir)
	for i := 0; i < 3; i++ {
		f, err := os.Create(filepath.Join(testDir, fmt.Sprintf("file-%d.txt", i)))
		require.NoError(t, err)
		defer f.Close()
		_, err = f.WriteString(fmt.Sprintf("Test file %d", i))
		require.NoError(t, err)
		require.NoError(t, stageAll(testDir), "error staging file %d", i)
		expectedHash, err := commit(testDir, fmt.Sprintf("Commit %d", i))
		require.NoError(t, err)
		currentHash, err := pkgHandler.CurrentCommitHash()
		assert.NoError(t, err, "error getting current commit hash")
		assert.Equal(t, expectedHash, currentHash, "current commit hash does not match expected hash")
	}
}

func initGitRepo(path string) error {
	for _, cmd := range []*exec.Cmd{
		exec.Command("git", "-C", path, "init"),
		exec.Command("git", "-C", path, "config", "user.name", "user"),
		exec.Command("git", "-C", path, "config", "user.email", "user@email.com"),
	} {
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func stageAll(path string) error {
	return exec.Command("git", "-C", path, "add", ".").Run()
}

func commit(path string, message string) (string, error) {
	err := exec.Command("git", "-C", path, "commit", "-m", message).Run()
	if err != nil {
		return "", err
	}
	return headCommitHash(path)
}

func headCommitHash(path string) (string, error) {
	out, err := exec.Command("git", "-C", path, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}
	if len(out) != 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return string(out), nil
}

func TestCurrentVersion(t *testing.T) {
	type testCase struct {
		name    string
		path    string
		version string
		err     error
	}

	prepareTest := func(t *testing.T, path string, tags []string) {
		tFile, err := os.Create(filepath.Join(path, "readme.txt"))
		if err != nil {
			t.Fatal(err)
		}
		defer tFile.Close()
		_, err = tFile.WriteString("Test file")
		if err != nil {
			t.Fatal(err)
		}
		for _, cmd := range []*exec.Cmd{
			exec.Command("git", "-C", path, "init"),
			exec.Command("git", "-C", path, "add", "readme.txt"),
			exec.Command("git", "-C", path, "config", "user.name", "user"),
			exec.Command("git", "-C", path, "config", "user.email", "user@email.com"),
			exec.Command("git", "-C", path, "commit", "-m", "Initial commit"),
		} {
			err := cmd.Run()
			if err != nil {
				t.Fatal(err)
			}
		}
		for _, tag := range tags {
			err = exec.Command("git", "-C", path, "tag", "-a", tag, "-m", tag).Run()
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	ts := []testCase{
		func() testCase {
			path := t.TempDir()
			prepareTest(t, path, []string{"v1.0.0"})
			return testCase{
				name:    "HEAD has a only one tag, which is a version tag",
				path:    path,
				version: "v1.0.0",
				err:     nil,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			prepareTest(t, path, []string{"some-tag", "v1.0.1", "v1.2"})
			return testCase{
				name:    "HEAD has many tags, which one is a version tag",
				path:    path,
				version: "v1.0.1",
				err:     nil,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			prepareTest(t, path, []string{"v1.0.1", "v1.0.0", "some-tag"})
			return testCase{
				name:    "HEAD has many tags, which more than one is a version tag",
				path:    path,
				version: "v1.0.1",
				err:     nil,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			prepareTest(t, path, []string{"v1.0.1", "v1.0.0", "v2.0.0"})
			return testCase{
				name:    "HEAD has many tags, and all of them are version tags",
				path:    path,
				version: "v2.0.0",
				err:     nil,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			prepareTest(t, path, []string{})
			return testCase{
				name:    "HEAD has no tags",
				path:    path,
				version: "",
				err:     ErrNoVersionsFound,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			prepareTest(t, path, []string{"some-tag", "another-tag"})
			return testCase{
				name:    "HEAD has tags, but none of them are version tags",
				path:    path,
				version: "",
				err:     ErrNoVersionsFound,
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgHandler := NewPackageHandler(tc.path)
			version, err := pkgHandler.CurrentVersion()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.version, version)
			}
		})
	}
}
