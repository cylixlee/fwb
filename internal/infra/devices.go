package infra

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
	"github.com/cylixlee/fwb/pkg/fsx"
)

type deviceRepository struct{}

// NewDeviceRepository persists custom devices in Flutter's config file.
func NewDeviceRepository() domain.DeviceRepository { return deviceRepository{} }

func customDevicesPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, consts.FlutterConfigDirName, consts.CustomDevicesFileName), nil
}

func (deviceRepository) Install(spec domain.DeviceSpec) (string, error) {
	path, err := customDevicesPath()
	if err != nil {
		return "", err
	}
	doc := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return "", domain.WrapError(consts.ErrorCodeDeviceParse, err, path)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	list := []any{}
	if raw, ok := doc[consts.JSONCustomDevices]; ok {
		if arr, ok := raw.([]any); ok {
			list = arr
		}
	}
	entry := map[string]any{
		consts.JSONID:                spec.ID,
		consts.JSONLabel:             spec.Label,
		consts.JSONSDKNameAndVersion: spec.SDKName,
		consts.JSONPlatform:          spec.Platform,
		consts.JSONEnabled:           true,
	}
	for hook, cmd := range spec.Hooks() {
		entry[hook] = cmd
	}
	replaced := false
	for i, item := range list {
		if m, ok := item.(map[string]any); ok && fmt.Sprint(m[consts.JSONID]) == spec.ID {
			list[i] = entry
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, entry)
	}
	doc[consts.JSONCustomDevices] = list
	if err := fsx.WriteJSONAtomic(path, doc, consts.FilePermissions); err != nil {
		return "", err
	}
	return path, nil
}

func (deviceRepository) Remove(id string) (string, error) {
	path, err := customDevicesPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return "", domain.WrapError(consts.ErrorCodeDeviceParse, err, path)
	}
	raw, _ := doc[consts.JSONCustomDevices].([]any)
	list := make([]any, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok && fmt.Sprint(m[consts.JSONID]) == id {
			continue
		}
		list = append(list, item)
	}
	doc[consts.JSONCustomDevices] = list
	if err := fsx.WriteJSONAtomic(path, doc, consts.FilePermissions); err != nil {
		return "", err
	}
	return path, nil
}
