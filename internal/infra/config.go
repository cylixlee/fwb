package infra

import (
	"encoding/json"
	"errors"
	"github.com/cylixlee/fwb/internal/consts"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/pkg/fsx"
)

// Defaults for the settings keys declared in domain.
var defaults = map[string]string{
	consts.ConfigKeyWindowsFlutter:    "",
	consts.ConfigKeyWindowsMirrorRoot: "",
	consts.ConfigKeyWindowsPubCache:   "",
	consts.ConfigKeyWindowsStorageURL: "",
	consts.ConfigKeySyncExtra:         "",
	consts.ConfigKeyDeviceID:          consts.DefaultDeviceID,
	consts.ConfigKeyLogLevel:          consts.DefaultLogLevel,
	consts.ConfigKeyLanguage:          "",
}

type configStore struct {
	path   string
	values map[string]string
}

// NewConfigStore loads ~/.config/fwb/config.json.
func NewConfigStore() (domain.ConfigStore, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, consts.AppDirectoryName, consts.ConfigFileName)
	st := &configStore{path: path, values: map[string]string{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return nil, domain.WrapError(consts.ErrorCodeConfigRead, err, path)
	}
	if len(data) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(data, &st.values); err != nil {
		return nil, domain.WrapError(consts.ErrorCodeConfigParse, err, path)
	}
	return st, nil
}

func (s *configStore) Path() string { return s.path }

func (s *configStore) Get(key string) string {
	if v, ok := os.LookupEnv(envName(key)); ok {
		return v
	}
	if v, ok := s.values[key]; ok {
		return v
	}
	return defaults[key]
}

func (s *configStore) Set(key, value string) error {
	if key == "" {
		return domain.NewError(consts.ErrorCodeConfigEmptyKey)
	}
	if value == "" {
		delete(s.values, key)
	} else {
		s.values[key] = value
	}
	return s.save()
}

func (s *configStore) Unset(key string) error {
	delete(s.values, key)
	return s.save()
}

func (s *configStore) All() map[string]string {
	out := map[string]string{}
	for k := range defaults {
		out[k] = s.Get(k)
	}
	for k := range s.values {
		out[k] = s.Get(k)
	}
	return out
}

func (s *configStore) Keys() []string {
	seen := map[string]bool{}
	for k := range defaults {
		seen[k] = true
	}
	for k := range s.values {
		seen[k] = true
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *configStore) save() error {
	if err := fsx.WriteJSONAtomic(s.path, s.values, consts.FilePermissions); err != nil {
		return domain.WrapError(consts.ErrorCodeConfigWrite, err, s.path)
	}
	return nil
}

// Source reports where a value comes from (for `fwb config list`).
func (s *configStore) Source(key string) string {
	if _, ok := os.LookupEnv(envName(key)); ok {
		return consts.EnvSourcePrefix + envName(key)
	}
	if _, ok := s.values[key]; ok {
		return consts.ConfigSourceFile
	}
	if _, ok := defaults[key]; ok {
		return consts.ConfigSourceDefault
	}
	return consts.ConfigSourceUnset
}

func envName(key string) string {
	replacer := strings.NewReplacer(consts.EnvKeySeparator, consts.EnvKeyReplacement, consts.EnvKeyDash, consts.EnvKeyReplacement)
	return consts.EnvPrefix + strings.ToUpper(replacer.Replace(key))
}
