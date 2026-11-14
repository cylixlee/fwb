package handler

import (
	"embed"
	"encoding/json"
	"os"
	"strings"

	"github.com/jeandeaual/go-locale"
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
				builder.SetString(tag, key, msg)
			}
			tags = append(tags, tag)
		}
	}
	matcher = language.NewMatcher(tags)
	if override := os.Getenv(consts.EnvLanguageOverride); override != "" {
		SetLanguage(override)
		return
	}
	locales, err := locale.GetLocales()
	if err != nil || len(locales) == 0 {
		SetLanguage(language.English.String())
		return
	}
	SetLanguage(locales...)
}

// SetLanguage selects the first matching language, falling back to English.
func SetLanguage(locales ...string) {
	tags := make([]language.Tag, 0, len(locales))
	for _, locale := range locales {
		if locale != "" {
			tags = append(tags, language.Make(locale))
		}
	}
	if len(tags) == 0 {
		tags = append(tags, language.English)
	}
	tag, _, _ := matcher.Match(tags...)
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
