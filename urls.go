package boltrouter

import (
	"fmt"
	"os"
	"strings"
)

func Path(root string, id string, suffixes ...string) string {
	parts := []string{root, id}
	parts = append(parts, suffixes...)
	parts = filterEmptyStrings(parts)
	return "/" + strings.Join(parts, "/")
}

func filterEmptyStrings(parts []string) []string {
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
func Url(root string, id string, suffixes ...string) string {
	host := os.Getenv("HOST")
	env := os.Getenv("ENV")
	protocol := "http"
	if env == "production" {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s%s", protocol, host, Path(root, id, suffixes...))
}
