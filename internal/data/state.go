package data

// import (
// 	"encoding/json"
// 	"os"
// 	"path/filepath"

// 	"github.com/gofrs/flock"
// )

// type State struct {
// 	dataDir string
// 	url     string
// 	name    string
// 	version string
// 	profile string
// 	tag     string
// 	lock    *flock.Flock
// }

// func NewState(url, name, version, profile, tag string) *State {
// 	return &State{
// 		dataDir: defaultDataDir,
// 		url:     url,
// 		name:    name,
// 		version: version,
// 		profile: profile,
// 		tag:     tag,
// 	}
// }

// func (s *State) FolderName() string {
// 	return s.name + "-" + s.tag
// }

// func (s *State) DataDir() string {
// 	return s.dataDir
// }

// func (s *State) SetDataDir(dataDir string) {
// 	s.dataDir = dataDir
// }

// func (s *State) TryLock() (bool, error) {
// 	if s.lock == nil {
// 		s.lock = flock.New(filepath.Join(s.dataDir, s.FolderName(), ".lock"))
// 	}
// 	return s.lock.TryLock()
// }

// func (s *State) Lock() error {
// 	if s.lock == nil {
// 		s.lock = flock.New(filepath.Join(s.dataDir, s.FolderName(), ".lock"))
// 	}
// 	return s.lock.Lock()
// }

// func (s *State) Unlock() error {
// 	if s.lock == nil {
// 		return nil
// 	}
// 	return s.lock.Unlock()
// }

// func (s *State) URL() string {
// 	return s.url
// }

// func (s *State) SetURL(url string) {
// 	s.url = url
// }

// func (s *State) Name() string {
// 	return s.name
// }

// func (s *State) SetName(name string) {
// 	s.name = name
// }

// func (s *State) Version() string {
// 	return s.version
// }

// func (s *State) SetVersion(version string) {
// 	s.version = version
// }

// func (s *State) Profile() string {
// 	return s.profile
// }

// func (s *State) SetProfile(profile string) {
// 	s.profile = profile
// }

// func (s *State) Tag() string {
// 	return s.tag
// }

// func (s *State) SetTag(tag string) {
// 	s.tag = tag
// }

// func (i *State) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(struct {
// 		URL     string `json:"url"`
// 		Name    string `json:"name"`
// 		Version string `json:"version"`
// 		Tag     string `json:"tag"`
// 	}{
// 		URL:     i.url,
// 		Name:    i.name,
// 		Version: i.version,
// 		Tag:     i.tag,
// 	})
// }

// func (i *State) UnmarshalJSON(data []byte) error {
// 	var tmp struct {
// 		URL     string `json:"url"`
// 		Name    string `json:"name"`
// 		Version string `json:"version"`
// 		Tag     string `json:"tag"`
// 	}
// 	if err := json.Unmarshal(data, &tmp); err != nil {
// 		return err
// 	}
// 	i.name = tmp.Name
// 	i.url = tmp.URL
// 	i.version = tmp.Version
// 	i.tag = tmp.Tag
// 	return nil
// }
