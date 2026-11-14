package handler

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/cylixlee/fwb/internal/consts"
)

var placeholder = regexp.MustCompile(`%[-+ 0#]*[0-9]*(?:\.[0-9]+)?[sdvqtxf]`)

func loadCatalog(t *testing.T, name string) map[string]string {
	t.Helper()
	data, err := catalogFS.ReadFile(consts.CatalogDirectory + consts.EmbedPathSeparator + name + consts.CatalogSuffix)
	if err != nil {
		t.Fatalf("read catalog %s: %v", name, err)
	}
	messages := map[string]string{}
	if err := json.Unmarshal(data, &messages); err != nil {
		t.Fatalf("parse catalog %s: %v", name, err)
	}
	return messages
}

func TestCatalogsHaveTheSameKeys(t *testing.T) {
	en := loadCatalog(t, "en")
	for _, name := range []string{"zh-CN"} {
		catalog := loadCatalog(t, name)
		if len(catalog) != len(en) {
			t.Errorf("%s has %d keys, en has %d", name, len(catalog), len(en))
		}
		for key := range en {
			if _, ok := catalog[key]; !ok {
				t.Errorf("%s is missing key %q", name, key)
			}
		}
		for key := range catalog {
			if _, ok := en[key]; !ok {
				t.Errorf("%s has unknown key %q", name, key)
			}
		}
	}
}

func TestCatalogPlaceholdersMatchAcrossLanguages(t *testing.T) {
	en := loadCatalog(t, "en")
	for _, name := range []string{"zh-CN"} {
		catalog := loadCatalog(t, name)
		for key, msg := range catalog {
			got := len(placeholder.FindAllString(msg, -1))
			want := len(placeholder.FindAllString(en[key], -1))
			if got != want {
				t.Errorf("%s: %q has %d placeholders, en has %d", name, key, got, want)
			}
		}
	}
}

func TestLanguageSelection(t *testing.T) {
	SetLanguage("zh-CN")
	if got := Current(); got != "zh-CN" {
		t.Fatalf("Current() = %q, want zh-CN", got)
	}
	zh := Translate(consts.KeyDeviceLabel)
	if !strings.Contains(zh, "Windows") {
		t.Fatalf("unexpected zh label: %q", zh)
	}

	SetLanguage("en")
	if got := Current(); got != "en" {
		t.Fatalf("Current() = %q, want en", got)
	}
	en := Translate(consts.KeyDeviceLabel)
	if zh == en {
		t.Fatalf("zh and en translations are identical: %q", en)
	}
}

func TestTranslateFormatsArguments(t *testing.T) {
	SetLanguage("en")
	got := Translate(consts.KeyDeviceInstalled, "fwb", "/tmp/device.json")
	if !strings.Contains(got, "fwb") || !strings.Contains(got, "/tmp/device.json") {
		t.Fatalf("Translate did not substitute arguments: %q", got)
	}
}

func TestTranslateErrorVerb(t *testing.T) {
	SetLanguage("en")
	got := Translate("doctor.wsl.fail", errors.New("boom"))
	if !strings.Contains(got, "boom") {
		t.Fatalf("Translate did not format %%v: %q", got)
	}
}
