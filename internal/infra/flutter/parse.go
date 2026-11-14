package flutter

import (
	"encoding/json"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

// Parse extracts the version JSON object from `flutter --version --machine`
// output, which may be prefixed with unrelated lines.
func Parse(out string) (domain.FlutterVersion, error) {
	text := out
	if i := strings.Index(text, consts.JSONObjectStart); i > 0 {
		text = text[i:]
	}
	var raw struct {
		FrameworkVersion  string `json:"frameworkVersion"`
		Channel           string `json:"channel"`
		FrameworkRevision string `json:"frameworkRevision"`
		EngineRevision    string `json:"engineRevision"`
		DartSdkVersion    string `json:"dartSdkVersion"`
	}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return domain.FlutterVersion{}, domain.WrapError(consts.ErrorCodeFlutterParse, err, strings.TrimSpace(out))
	}
	if raw.FrameworkVersion == "" || raw.EngineRevision == "" {
		return domain.FlutterVersion{}, domain.NewError(consts.ErrorCodeFlutterIncomplete, strings.TrimSpace(out))
	}
	return domain.FlutterVersion{
		Framework:         raw.FrameworkVersion,
		Channel:           raw.Channel,
		FrameworkRevision: raw.FrameworkRevision,
		EngineRevision:    raw.EngineRevision,
		DartSDK:           raw.DartSdkVersion,
	}, nil
}
