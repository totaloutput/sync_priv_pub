package tools

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// IsUnder returns true if a is under b
func IsUnder(a, b string) bool {
	aParts := strings.Split(a, "/")
	bParts := strings.Split(b, "/")
	if len(bParts) > len(aParts) {
		return false
	}

	matches := 0
	for i := 0; i < len(bParts); i++ {
		if aParts[i] != bParts[i] {
			break
		}
		matches++
	}

	if matches == len(bParts) {
		return true
	}

	return false
}

// GoBin returns the GOBIN path
func GoBin() string {
	goBin, exists := os.LookupEnv("GOBIN")
	if exists {
		return goBin
	}

	goPath, exists := os.LookupEnv("GOPATH")
	if exists {
		return path.Join(goPath, "bin")
	}

	home, exists := os.LookupEnv("HOME")
	if exists {
		return path.Join(home, "go", "bin")
	}

	return ""
}

// ThisFile returns the name of the running process's executable
func ThisFile() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}

	_, this := filepath.Split(path)
	return this
}