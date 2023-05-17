// Package gitcmd has a bunch of convience functions to work with Git.
// Each machine should use it's own Git.
package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/miekg/pgo/osutil"
	"go.science.ru.nl/log"
)

type Git struct {
	upstream string // upstream git repo
	branch   string // specific branch to get, 'main' is not specified
	user     string // what user to use
	dir      string // where to put it
}

// New returns a pointer to an intialized Git.
func New(upstream, branch, user, directory string) *Git {
	g := &Git{
		upstream: upstream,
		user:     user,
		branch:   branch,
		dir:      directory,
	}
	return g
}

func (g *Git) run(args ...string) ([]byte, error) {
	ctx := context.TODO()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.dir
	cmd.Env = []string{"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null"}
	uid, gid := osutil.User(g.user)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}

	log.Debugf("running in %q as %q %v", cmd.Dir, g.user, cmd.Args)

	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debug(string(out))
	}

	return bytes.TrimSpace(out), err
}

func (g *Git) IsCheckedOut() bool {
	info, err := os.Stat(path.Join(g.dir, ".git"))
	if err != nil {
		return false
	}
	return info.Name() == ".git" && info.IsDir()
}

// Checkout will do the initial check of the git repo. If the g.dir directory already exist and has
// a .git subdirectory, it will assume the checkout has been done during a previuos run.
func (g *Git) Checkout() error {
	if g.IsCheckedOut() {
		return nil
	}

	if err := os.MkdirAll(g.dir, 0775); err != nil {
		log.Errorf("Directory %q can not be created", g.dir)
		return fmt.Errorf("failed to create directory %q: %s", g.dir, err)
	}

	if os.Geteuid() == 0 { // set g.dir to the correct owner, if we are root
		uid, gid := osutil.User(g.user)
		if err := os.Chown(g.dir, int(uid), int(gid)); err != nil {
			log.Errorf("Directory %q can not be chown-ed to %q: %s", g.dir, g.user, err)
			return fmt.Errorf("failed to chown directory %q to %q: %s", g.dir, g.user, err)
		}
	}

	_, err := g.run("clone", "-b", g.branch, g.upstream, g.dir)
	if err != nil {
		return err
	}

	return err
}

// Pull pulls from upstream. If the returned bool is true there were updates if on the files named in names.
func (g *Git) Pull(names []string) (bool, error) {
	if err := g.Stash(); err != nil {
		return false, err
	}

	out, err := g.run("pull", "--stat", "origin", g.branch)
	if err != nil {
		return false, err
	}
	return g.OfInterest(out, names), nil
}

// Hash returns the git hash of HEAD in the repo in g.dir. Empty string is returned in case of an error.
// The hash is always truncated to 8 hex digits.
func (g *Git) Hash() string {
	out, err := g.run("rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	if len(out) < 8 {
		return ""
	}
	return string(out)[:8]
}

// Rollback checks out commit <hash>, and return nil if no errors are encountered.
func (g *Git) Rollback(hash string) error {
	if err := g.Stash(); err != nil {
		return err
	}
	_, err := g.run("checkout", hash)
	return err
}

func (g *Git) Stash() error { _, err := g.run("stash"); return err }
