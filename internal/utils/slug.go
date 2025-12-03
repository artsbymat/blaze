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

func SlugifyPath(path string) string {
	if path == "" || path == "." {
		return ""
	}
	
	parts := strings.Split(filepath.ToSlash(path), "/")
	slugged := make([]string, 0, len(parts))
	
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		sluggedPart := strings.ToLower(part)
		sluggedPart = reNonAlnum.ReplaceAllString(sluggedPart, "-")
		sluggedPart = strings.Trim(sluggedPart, "-")
		if sluggedPart != "" {
			slugged = append(slugged, sluggedPart)
		}
	}
	
	return filepath.Join(slugged...)
}
