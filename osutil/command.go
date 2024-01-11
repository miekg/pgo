package osutil

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func EnvVars(env []string) []string {
	envnames := make([]string, len(env))
	for i := range env {
		fs := strings.Split(env[i], "=")
		envnames[i] = fs[0]
	}
	return envnames
}

func RunAs(cmd *exec.Cmd, user string) error {
	uid, gid := User(user)
	if uid == 0 && gid == 0 && user != "root" {
		return fmt.Errorf("failed to resolve user %q to uid/gid", user)
	}
	dgid := DockerGroup()
	if dgid == 0 {
		return fmt.Errorf("failed to resolve docker to gid")
	}
	groups := Groups(user)
	groups = append(groups, dgid)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: groups}

	path := "/usr/sbin:/usr/bin:/sbin:/bin"
	cmd.Env = []string{env("HOME", Home(user)), env("PATH", path)}
	return nil
}

func env(k, v string) string { return k + "=" + v }
