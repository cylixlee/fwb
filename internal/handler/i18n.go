package handler

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cylixlee/fwb/internal/consts"
	"github.com/samber/lo"
)

//go:embed catalogs/*.json
var catalogFS embed.FS

var (
	catalogs = map[string]map[string]string{}
	current  = consts.DefaultLanguage
)

// init loads the embedded catalogs and selects a language from the environment.
func init() {
	entries, err := catalogFS.ReadDir(consts.CatalogDirectory)
	if err == nil {
		for _, e := range entries {
			name := strings.TrimSuffix(e.Name(), consts.CatalogSuffix)
			data, err := catalogFS.ReadFile(consts.CatalogDirectory + consts.EmbedPathSeparator + e.Name())
			if err != nil {
				continue
			}
			m := map[string]string{}
			if json.Unmarshal(data, &m) == nil {
				catalogs[name] = m
			}
		}
	}
	lang, _ := lo.Coalesce(
		os.Getenv(consts.EnvLanguageOverride),
		os.Getenv(consts.EnvLocaleAll),
		os.Getenv(consts.EnvLocaleMessages),
		os.Getenv(consts.EnvLang),
	)
	SetLanguage(lang)
}

// SetLanguage selects a language, falling back to English.
func SetLanguage(lang string) {
	norm := normalize(lang)
	if _, ok := catalogs[norm]; ok {
		current = norm
		return
	}
	if i := strings.IndexByte(norm, '-'); i > 0 {
		if _, ok := catalogs[norm[:i]]; ok {
			current = norm[:i]
			return
		}
	}
	current = consts.DefaultLanguage
}

func normalize(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if i := strings.IndexByte(lang, consts.LanguageEncodingSeparator); i >= 0 {
		lang = lang[:i]
	}
	lang = strings.ReplaceAll(lang, consts.LanguageUnderscore, consts.LanguageDash)
	switch lang {
	case consts.ChineseLanguageShort, consts.ChineseLanguageCN, consts.ChineseLanguageHans, consts.ChineseLanguageHansCN:
		return consts.ChineseLanguage
	case consts.EnglishLanguage, consts.EnglishLanguageUS, consts.EnglishLanguageGB, consts.LanguageC, consts.LanguagePOSIX:
		return consts.EnglishLanguage
	}
	return lang
}

// T returns the localized format string for key, formatted with args.
func Translate(key string, args ...any) string {
	msg := lookup(current, key)
	if msg == "" {
		msg = lookup(consts.DefaultLanguage, key)
	}
	if msg == "" {
		msg = key
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

func lookup(lang, key string) string {
	if m, ok := catalogs[lang]; ok {
		return m[key]
	}
	return ""
}
