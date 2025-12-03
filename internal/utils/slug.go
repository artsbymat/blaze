package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

var reNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func PathToSlug(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.ToLower(name)
	name = reNonAlnum.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}
