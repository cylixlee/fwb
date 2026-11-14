package mirror

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/cylixlee/fwb/internal/consts"
)

func (m *projectMirror) Fingerprint(engineRevision string) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, consts.FingerprintEngineFormat, engineRevision)
	for _, name := range []string{consts.PubspecFileName, consts.PubspecLockFileName, consts.MetadataFileName} {
		data, err := os.ReadFile(filepath.Join(m.project.Root, name))
		if err == nil {
			fmt.Fprintf(h, consts.FingerprintFileFormat, name, sha256.Sum256(data))
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	for _, rel := range m.syncRoots() {
		dir := filepath.Join(m.project.Root, rel)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			continue
		}
		var lines []string
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(m.project.Root, path)
			lines = append(lines, fmt.Sprintf(consts.FingerprintLineFormat, relPath, info.Size(), info.ModTime().UnixNano()))
			return nil
		})
		if err != nil {
			return "", err
		}
		sort.Strings(lines)
		for _, l := range lines {
			fmt.Fprintln(h, l)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
