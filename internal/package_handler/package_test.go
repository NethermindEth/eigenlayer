package package_handler

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigen-wiz/internal/package_handler/testdata"
	"github.com/stretchr/testify/assert"
)

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

	mockTapRepo := "https://github.com/NethermindEth/mock-avs.git"
	tag := "v0.1.0"

	t.Logf("Cloning mock tap repo %s and tag %s into %s", mockTapRepo, tag, pkgFolder)

	if err := exec.Command("git", "clone", "--single-branch", "-b", tag, mockTapRepo, pkgFolder).Run(); err != nil {
		t.Fatal("error cloning the mock tap repo: " + err.Error())
	}
	return pkgFolder
}

func TestProfilesNames(t *testing.T) {
	testDir := t.TempDir()
	testdata.SetupDir(t, "manifests", testDir)

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
	testDir := t.TempDir()
	testdata.SetupDir(t, "packages", testDir)

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
	testDir := t.TempDir()
	testdata.SetupDir(t, "packages", testDir)

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
							Type:   "id",
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
			err:     InvalidConfError{message: "Invalid profile.yml", missingFields: []string{"options"}},
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
