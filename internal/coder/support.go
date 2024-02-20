package coder

import "strings"

func normalizeVersion(ver string) string {
	return strings.ReplaceAll(ver, "-", "_")
}
