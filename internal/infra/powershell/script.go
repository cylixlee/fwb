package powershell

import (
	"bytes"
	"embed"
	"encoding/base64"
	"strings"
	"sync"
	"text/template"
	"unicode/utf16"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/cylixlee/fwb/internal/domain"
)

//go:embed scripts
var scripts embed.FS

//go:embed scripts/preamble.ps1
var preambleScript string

var (
	tmplOnce sync.Once
	tmpl     *template.Template
	tmplErr  error
)

// quote returns a single-quoted PowerShell string literal.
func quote(s string) string {
	return consts.SingleQuote + strings.ReplaceAll(s, consts.SingleQuote, consts.EscapedSingleQuote) + consts.SingleQuote
}

func renderScript(name string, data any) (string, error) {
	raw, err := scripts.ReadFile(consts.ScriptsDirectory + consts.EmbedPathSeparator + name)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(name, consts.ScriptTemplateSuffix) {
		return string(raw), nil
	}
	tmplOnce.Do(func() {
		tmpl, tmplErr = template.New(consts.TemplateRootName).
			Funcs(template.FuncMap{consts.TemplateFunctionQuote: quote}).
			ParseFS(scripts, consts.ScriptsGlob)
	})
	if tmplErr != nil {
		return "", tmplErr
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", domain.WrapError(consts.ErrorCodeWinScript, err, name)
	}
	return buf.String(), nil
}

func encodeCommand(script string) string {
	u := utf16.Encode([]rune(preambleScript + script))
	buf := make([]byte, 0, len(u)*2)
	for _, r := range u {
		buf = append(buf, byte(r), byte(r>>8))
	}
	return base64.StdEncoding.EncodeToString(buf)
}
