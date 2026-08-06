package envelope

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const sep = "__"

func Marshal(environ []string, prefix string) ([]byte, error) {
	doc, err := Parse(environ, prefix)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(doc)
}

func Parse(environ []string, prefix string) (any, error) {
	root := tree{}
	sort.Strings(environ)
	for _, kv := range environ {
		name, value, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		path, err := parsePath(strings.TrimPrefix(name, prefix))
		if err == nil {
			err = root.set(path, scalar(value))
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	return root.value()
}

type tree map[any]any

func parsePath(name string) ([]any, error) {
	segments := strings.Split(name, sep)
	path := make([]any, 0, len(segments))
	for _, s := range segments {
		switch {
		case s == "":
			return nil, fmt.Errorf("empty path segment")
		case s[0] == '_':
			path = append(path, s[1:])
		case isIndex(s):
			n, _ := strconv.Atoi(s)
			path = append(path, n)
		default:
			path = append(path, s)
		}
	}
	return path, nil
}

func isIndex(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func (t tree) set(path []any, val any) error {
	for _, k := range path[:len(path)-1] {
		child, ok := t[k].(tree)
		if !ok {
			if _, taken := t[k]; taken {
				return fmt.Errorf("%v is already a value, nothing can nest under it", k)
			}
			child = tree{}
			t[k] = child
		}
		t = child
	}
	last := path[len(path)-1]
	if _, taken := t[last]; taken {
		return fmt.Errorf("%v already holds a nested block, it cannot also be a value", last)
	}
	t[last] = val
	return nil
}

func (t tree) value() (any, error) {
	indexes := make([]int, 0, len(t))
	for k := range t {
		if n, ok := k.(int); ok {
			indexes = append(indexes, n)
		}
	}
	if len(indexes) > 0 && len(indexes) != len(t) {
		return nil, fmt.Errorf("mixes list indexes and names in the same block")
	}
	if len(indexes) > 0 {
		sort.Ints(indexes)
		list := make([]any, len(indexes))
		for i, n := range indexes {
			v, err := resolve(t[n])
			if err != nil {
				return nil, fmt.Errorf("%d: %w", n, err)
			}
			list[i] = v
		}
		return list, nil
	}
	m := make(map[string]any, len(t))
	for k, child := range t {
		v, err := resolve(child)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", k, err)
		}
		m[k.(string)] = v
	}
	return m, nil
}

func resolve(v any) (any, error) {
	if sub, ok := v.(tree); ok {
		return sub.value()
	}
	return v, nil
}

func scalar(s string) any {
	if s == "" {
		return ""
	}
	var v any
	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	switch c := v.(type) {
	case []any:
		if len(c) > 0 {
			return s
		}
	case map[string]any:
		if len(c) > 0 {
			return s
		}
	}
	return v
}
