package handler

import (
	"embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/cylixlee/fwb/internal/consts"
)

//go:embed catalogs/*.json
var catalogFS embed.FS

var (
	builder *catalog.Builder
	matcher language.Matcher
	printer *message.Printer
	current language.Tag
)

// init loads the embedded catalogs and selects a language from the environment.
func init() {
	builder = catalog.NewBuilder(catalog.Fallback(language.English))
	var tags []language.Tag
	entries, err := catalogFS.ReadDir(consts.CatalogDirectory)
	if err == nil {
		for _, entry := range entries {
			tag := language.Make(strings.TrimSuffix(entry.Name(), consts.CatalogSuffix))
			data, err := catalogFS.ReadFile(consts.CatalogDirectory + consts.EmbedPathSeparator + entry.Name())
			if err != nil {
				continue
			}
			messages := map[string]string{}
			if json.Unmarshal(data, &messages) != nil {
				continue
			}
			for key, msg := range messages {
				_ = builder.SetString(tag, key, msg)
			}
			tags = append(tags, tag)
		}
	}
	matcher = language.NewMatcher(tags)
	locale, _ := lo.Coalesce(
		os.Getenv(consts.EnvLanguageOverride),
		os.Getenv(consts.EnvLocaleAll),
		os.Getenv(consts.EnvLocaleMessages),
		os.Getenv(consts.EnvLang),
	)
	SetLanguage(locale)
}

// SetLanguage selects a language, falling back to English.
func SetLanguage(locale string) {
	tag, _, _ := matcher.Match(parseLocale(locale))
	current = tag
	printer = message.NewPrinter(tag, message.Catalog(builder))
}

// Current returns the selected language tag.
func Current() string { return current.String() }

// Translate formats the message identified by key for the current language.
func Translate(key string, args ...any) string {
	if printer == nil {
		return key
	}
	return printer.Sprintf(key, args...)
}

func parseLocale(locale string) language.Tag {
	locale = strings.TrimSpace(locale)
	if i := strings.IndexByte(locale, '.'); i >= 0 {
		locale = locale[:i]
	}
	locale = strings.ReplaceAll(locale, "_", "-")
	if locale == "" {
		return language.Und
	}
	return language.Make(locale)
}
