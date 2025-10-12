package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Bundle struct {
	mu				sync.RWMutex
	defaultLaguage	string
	langs			map[string]map[string]string
}

func Load(dir string, defaultLaguage string) (*Bundle, error) {
	b := &Bundle{
		defaultLaguage: strings.ToLower(defaultLaguage),
		langs:       make(map[string]map[string]string),
	}

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil { return err }
		if d.IsDir() { return nil }
		if !strings.HasSuffix(d.Name(), ".json") { return nil }

		lang := strings.TrimSuffix(d.Name(), ".json")
		lang = strings.ToLower(lang)

		raw, err := os.ReadFile(path)
		if err != nil { return fmt.Errorf("read %s: %w", path, err) }

		var m map[string]string
		if err := json.Unmarshal(raw, &m); err != nil {
			return fmt.Errorf("unmarshal %s: %w", path, err)
		}

		norm := make(map[string]string, len(m))
		for k, v := range m {
			norm[strings.TrimSpace(k)] = v
		}

		b.mu.Lock()
		b.langs[lang] = norm
		b.mu.Unlock()
		return nil
	})
	if err != nil { return nil, err }

	return b, nil
}


func (b *Bundle) BestLang(requested string) string {
	req := strings.ToLower(strings.TrimSpace(requested))
	b.mu.RLock()
	defer b.mu.RUnlock()

	if _, ok := b.langs[req]; ok {
		return req
	}
	
	if i := strings.IndexByte(req, '-'); i > 0 {
		base := req[:i]
		if _, ok := b.langs[base]; ok {
			return base
		}
	}
	
	if _, ok := b.langs[b.defaultLaguage]; ok {
		return b.defaultLaguage
	}
	
	for k := range b.langs {
		return k
	}
	return b.defaultLaguage
}

func (b *Bundle) T(lang, key string, data map[string]any) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	val := ""
	if lm, ok := b.langs[lang]; ok {
		val = lm[key]
	}
	if val == "" {
		if dm, ok := b.langs[b.defaultLaguage]; ok {
			val = dm[key]
		}
	}
	if val == "" {
		val = key
	}
	if len(data) == 0 {
		return val
	}

	re := regexp.MustCompile(`\{([a-zA-Z0-9_.-]+)\}`)
	return re.ReplaceAllStringFunc(val, func(s string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(s, "{"), "}")
		if v, ok := data[name]; ok && v != nil {
			return fmt.Sprint(v)
		}
		return s
	})
}