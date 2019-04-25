package tools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// DiffR performs a recursive diff between two directories, and returns whether or not their contents are identical
func DiffR(a, b string, exclude ...string) (bool, error) {
	diff, exists := Which("diff")
	if !exists {
		return false, fmt.Errorf("Couldn't find diff command")
	}

	args := []string{
		"--recursive",
		"--brief",
	}
	for _, p := range exclude {
		args = append(args, fmt.Sprintf("--exclude=%s", p))
	}
	args = append(args, a)
	args = append(args, b)

	cmd := exec.Command(diff, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("%s\n%s", output, err)
	}

	return true, nil
}

// ReplaceR replaces all non-overlapping occurrences of old with new
// in all files under the path, except for those under the exclude paths.
func ReplaceR(path, old, new string, exclude ...string) (int, error) {
	count := 0
	oldBytes := []byte(old)
	newBytes := []byte(new)

	rename := func(n string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(path, n)
		if err != nil {
			return fmt.Errorf("Can't determine relative path for %s: %s", n, err)
		}

		// Skip renaming files that are under an excluded directory
		for _, x := range exclude {
			if IsUnder(rel, x) {
				// The file is under the excluded path, so skip it
				//fmt.Println("Skipping", rel)
				return nil
			}
		}

		in, err := ioutil.ReadFile(n)
		if err != nil {
			return err
		}

		out := bytes.ReplaceAll(in, oldBytes, newBytes)
		if bytes.Equal(in, out) {
			// No replacements made, so no need to re-write the file
			return nil
		}

		err = ioutil.WriteFile(n, out, info.Mode())
		//fmt.Println("Modified", rel)
		count += 1

		return err
	}

	err := filepath.Walk(path, rename)
	return count, err
}