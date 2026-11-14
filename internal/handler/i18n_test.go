package handler

import (
	"regexp"
	"testing"
)

var placeholder = regexp.MustCompile(`%[-+ 0#]*[0-9]*(?:\.[0-9]+)?[sdvqtxf]`)

func TestCatalogsHaveTheSameKeys(t *testing.T) {
	en, ok := catalogs["en"]
	if !ok {
		t.Fatal("missing en catalog")
	}
	for lang, catalog := range catalogs {
		if len(catalog) != len(en) {
			t.Errorf("%s has %d keys, en has %d", lang, len(catalog), len(en))
		}
		for key := range en {
			if _, ok := catalog[key]; !ok {
				t.Errorf("%s is missing key %q", lang, key)
			}
		}
		for key := range catalog {
			if _, ok := en[key]; !ok {
				t.Errorf("%s has unknown key %q", lang, key)
			}
		}
	}
}

func TestCatalogPlaceholdersMatchAcrossLanguages(t *testing.T) {
	en := catalogs["en"]
	for lang, catalog := range catalogs {
		if lang == "en" {
			continue
		}
		for key, msg := range catalog {
			got := len(placeholder.FindAllString(msg, -1))
			want := len(placeholder.FindAllString(en[key], -1))
			if got != want {
				t.Errorf("%s: %q has %d placeholders, en has %d", lang, key, got, want)
			}
		}
	}
}

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"zh_CN.UTF-8": "zh-CN",
		"zh":          "zh-CN",
		"en_US.UTF-8": "en",
		"C":           "en",
		"POSIX":       "en",
	}
	for in, want := range cases {
		if got := normalize(in); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
}
