package main

import (
	"./stdlib/strings"
)

// Tentative implementation
// "foo" => "foo"
// "foo/bar" => "bar"
// "./stdlib/foo/bar" => "bar"
func getPackageNameInImport(importPath string) string {
	return string(getBaseNameFromImport(importPath))
}

func getBaseNameFromImport(path string) string {
	if strings.Contains(path, "/") {
		words := strings.Split(path, "/")
		r := words[len(words)-1]
		return r
	} else {
		return path
	}

}

func getIndex(item string, list []string) int {
	for id, v := range list {
		if v == item {
			return id
		}
	}
	return -1
}

func inArray(item string, list []string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
