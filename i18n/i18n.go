package i18n

/*
|--------------------------------------------------------------------------
| I18n
|--------------------------------------------------------------------------
|
| Implements I18n as part of the i18n package in the Gofreight framework.
| Key symbols: Translator, New, LoadDir, SetLocale, Locale, T.
| 
| The i18n package loads JSON locale files from config/locales and
| resolves translation keys in views and controllers.
| 
| Middleware can set locale from session or Accept-Language; helpers
| mirror Laravel-style __() usage in GFT.
| 
| Symbols defined here include: Translator (exported type); New (New
| creates a translator with default locale.); LoadDir (LoadDir loads JSON
| translation files from a directory (e.g. config/locales/en.json).);
| SetLocale (SetLocale sets the active locale.); Locale (Locale returns
| the active locale.); T (T translates a key with optional :placeholder
| replacements.); FormatNumber (FormatNumber formats a number for the
| locale (basic grouping).).
| 
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translator loads and resolves translation keys.
type Translator struct {
	mu       sync.RWMutex
	locale   string
	fallback string
	catalog  map[string]map[string]string
}

// New creates a translator with default locale.
func New(locale, fallback string) *Translator {
	return &Translator{
		locale:   locale,
		fallback: fallback,
		catalog:  make(map[string]map[string]string),
	}
}

// LoadDir loads JSON translation files from a directory (e.g. config/locales/en.json).
func (t *Translator) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		locale := strings.TrimSuffix(e.Name(), ".json")
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return err
		}
		var messages map[string]string
		if err := json.Unmarshal(raw, &messages); err != nil {
			return err
		}
		t.mu.Lock()
		t.catalog[locale] = messages
		t.mu.Unlock()
	}
	return nil
}

// SetLocale sets the active locale.
func (t *Translator) SetLocale(locale string) {
	t.mu.Lock()
	t.locale = locale
	t.mu.Unlock()
}

// Locale returns the active locale.
func (t *Translator) Locale() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.locale
}

// T translates a key with optional :placeholder replacements.
func (t *Translator) T(key string, replacements map[string]string) string {
	msg := t.lookup(key, t.locale)
	if msg == "" && t.fallback != "" {
		msg = t.lookup(key, t.fallback)
	}
	if msg == "" {
		return key
	}
	for k, v := range replacements {
		msg = strings.ReplaceAll(msg, ":"+k, v)
	}
	return msg
}

func (t *Translator) lookup(key, locale string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if cat, ok := t.catalog[locale]; ok {
		if msg, ok := cat[key]; ok {
			return msg
		}
	}
	return ""
}

// FormatNumber formats a number for the locale (basic grouping).
func (t *Translator) FormatNumber(n int64) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}
