package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func toPascal(s string) string {
	parts := splitName(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

func toCamel(s string) string {
	p := toPascal(s)
	if len(p) == 0 {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

func toSnake(s string) string {
	parts := splitName(s)
	return strings.ToLower(strings.Join(parts, "_"))
}

func splitName(s string) []string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "-", "_")
	return strings.Split(s, "_")
}

type Generator struct {
	Name     string
	Type     string
	BasePath string
	Package  string
}

func (g Generator) writeFile(ext string, content string) error {
	_ = os.MkdirAll(g.BasePath, 0755)
	filename := filepath.Join(g.BasePath, toPascal(g.Name)+ext)
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}
	fmt.Printf("  ✅ Created %s: %s\n", g.Type, filename)
	return nil
}

func ModuleName() string {
	b, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/rancago/framework"
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "github.com/rancago/framework"
}
