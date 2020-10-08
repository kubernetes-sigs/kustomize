package utils

import (
	"log"
	"os"
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
