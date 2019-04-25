package tools

import (
	"fmt"
	"os/exec"
)

// GoDropMod removes the module name from a go.mod file
func GoDropMod(modFile, name string, env []string) error {
	goCmd, exists := Which("go")
	if !exists {
		return fmt.Errorf("Couldn't find go command")
	}

	cmd := exec.Command(goCmd, "mod", "edit", fmt.Sprintf("-droprequire=%s", name), modFile)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	}

	return nil
}

// GoGetMod updates the go.mod file with the latest version of the module name
func GoGetMod(path, name string, env []string) error {
	goCmd, exists := Which("go")
	if !exists {
		return fmt.Errorf("Couldn't find go command")
	}

	cmd := exec.Command(goCmd, "get", "-v", name)
	cmd.Dir = path
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	}

	return nil
}


// GoSetMod sets the module name of a go.mod file
func GoSetMod(modFile, name string, env []string) error {
	goCmd, exists := Which("go")
	if !exists {
		return fmt.Errorf("Couldn't find go command")
	}

	cmd := exec.Command(goCmd, "mod", "edit", fmt.Sprintf("-module=%s", name), modFile)
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	}

	return nil
}

// GoTidyMod removes stale references from the go module files
func GoTidyMod(path string, env []string) error {
	goCmd, exists := Which("go")
	if !exists {
		return fmt.Errorf("Couldn't find go command")
	}

	cmd := exec.Command(goCmd, "mod", "tidy", "-v")
	cmd.Dir = path
	cmd.Env = env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s\n%s", output, err)
	}

	return nil
}
