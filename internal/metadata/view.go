package metadata

import (
	"strconv"
	"strings"
)

type View struct {
	store   *Store
	modelID string
}

func (v View) ModelID() string { return v.modelID }

func (v View) Raw(key Key) (any, bool) {
	if v.store == nil {
		return nil, false
	}
	return v.store.Get(v.modelID, key)
}

func (v View) Present(key Key) bool {
	x, ok := v.Raw(key)
	if !ok || x == nil {
		return false
	}
	switch t := x.(type) {
	case string:
		return strings.TrimSpace(t) != ""
	case []string:
		return len(t) > 0
	default:
		return true
	}
}

func (v View) String(key Key) (string, bool) {
	x, ok := v.Raw(key)
	if !ok || x == nil {
		return "", false
	}
	s, ok := x.(string)
	if !ok {
		return "", false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	return s, true
}

func (v View) Strings(key Key) ([]string, bool) {
	x, ok := v.Raw(key)
	if !ok || x == nil {
		return nil, false
	}
	switch t := x.(type) {
	case []string:
		out := make([]string, 0, len(t))
		for _, s := range t {
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return nil, false
		}
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) == 0 {
			return nil, false
		}
		return out, true
	default:
		return nil, false
	}
}

func (v View) Bool(key Key) (bool, bool) {
	x, ok := v.Raw(key)
	if !ok || x == nil {
		return false, false
	}
	switch t := x.(type) {
	case bool:
		return t, true
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(t))
		if err != nil {
			return false, false
		}
		return b, true
	default:
		return false, false
	}
}

func (v View) Int(key Key) (int, bool) {
	x, ok := v.Raw(key)
	if !ok || x == nil {
		return 0, false
	}
	switch t := x.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	case string:
		i, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return 0, false
		}
		return i, true
	default:
		return 0, false
	}
}
