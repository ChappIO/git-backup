package git_backup

import (
	"slices"
	"strings"
)

func boolPointer(b bool) *bool {
	return &b
}

// A negative filter - checks if the given repo is in the excluded slice.
func isExcluded(excluded []string, repo string) bool {
	return slices.ContainsFunc(excluded, func(s string) bool {
		if strings.EqualFold(s, repo) {
			return true
		}

		if strings.Contains(s, "/") {
			return false
		}

		repoOwner := repo[:strings.Index(repo, "/")]
		return strings.EqualFold(s, repoOwner)
	})
}
