package utils

import (
	"log"
	"os"
	"strings"
)

func DirExists(name string) bool {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func SliceToSet(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, x := range slice {
		if _, ok := result[x]; ok {
			log.Fatalf("programmer error - repeated value: %s", x)
		} else {
			result[x] = true
		}
	}
	return result
}

func ExtractModule(m string) string {
	k := strings.Index(m, " => ")
	if k < 0 {
		return m
	}
	return m[:k]
}

func SliceContains(slice []string, target string) bool {
	for _, x := range slice {
		if x == target {
			return true
		}
	}
	return false
}
