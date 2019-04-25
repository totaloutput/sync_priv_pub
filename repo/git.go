package repo

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/soterium/sync_priv_pub/tools"
)

type GitRepo struct {
	// The path to the repository (ex: github.com/user/repo_name)
	Path string
}

// Clone clones the repo to the dst
func (g *GitRepo) Clone(dst string, useHttps bool) error {
	git, exists := tools.Which("git")
	if !exists {
		return fmt.Errorf("Couldn't find git command")
	}

	var src string
	if useHttps {
		src = g.HTTPSURL()
	} else {
		src = g.SSHUrl()
	}

	cmd := exec.Command(git, "clone", "--quiet", src, dst)
	err := cmd.Run()
	return err
}

// Host returns the host of the repo path
func (g *GitRepo) Host() string {
	parts := strings.Split(g.Path, "/")
	return parts[0]
}

// HTTPSURL returns an https url for the repo
func (g *GitRepo) HTTPSURL() string {
	return fmt.Sprintf("https://%s/%s.git", g.Host(), g.RepoPath())
}

// RepoPath returns the repository identifier part of the path (without the git host)
func (g *GitRepo) RepoPath() string {
	parts := strings.Split(g.Path, "/")
	return path.Join(parts[1:]...)
}

// SSHURL returns an ssh url for the repo
func (g *GitRepo) SSHUrl() string {
	return fmt.Sprintf("git@%s:%s.git", g.Host(), g.RepoPath())
}

// String returns a string representing the GitRepo
func (g *GitRepo) String() string {
	return g.Path
}
