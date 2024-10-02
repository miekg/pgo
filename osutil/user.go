package osutil

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"go.science.ru.nl/log"
)

// User looks up the username u and return the uid and gid. If the username can't be found 0, 0 is returned.
func User(u string) (uint32, uint32) {
	u1, err := user.Lookup(u)
	if err != nil {
		return 0, 0
	}
	uid, _ := strconv.ParseInt(u1.Uid, 10, 32)
	gid, _ := strconv.ParseInt(u1.Gid, 10, 32)
	return uint32(uid), uint32(gid)
}

// User looks up the username u and return the home directory.
func Home(u string) string {
	u1, err := user.Lookup(u)
	if err != nil {
		return ""
	}
	switch u1.HomeDir {
	case "/":
		fallthrough
	case "/bestaat-niet":
		fallthrough
	case "/non-existent":
		// Create a home dir, to store docker login credentials.
		home := filepath.Join(filepath.Join(os.TempDir(), u))
		dockerhome := filepath.Join(home, ".docker")
		if _, err := os.Stat(dockerhome); os.IsNotExist(err) {
			if err := os.MkdirAll(dockerhome, 750); err != nil {
				log.Errorf("Failed to create %q: %s", dockerhome, err)
			}
		}
		uid, gid := User(u)
		os.Chown(home, int(uid), int(gid))
		os.Chown(dockerhome, int(uid), int(gid))
		return home
	}
	return u1.HomeDir
}

// Dockergroup returns the gid of the docker group.
func DockerGroup() uint32 {
	g, err := user.LookupGroup("docker")
	if err != nil {
		return 0
	}
	gid, _ := strconv.ParseInt(g.Gid, 10, 32)
	return uint32(gid)
}

// Groups returns the supplement groups the user is a member of.
func Groups(u string) []uint32 {
	u1, err := user.Lookup(u)
	if err != nil {
		return nil
	}
	groups, err := u1.GroupIds()
	if err != nil {
		return nil
	}

	gids := make([]uint32, len(groups))
	for i := range groups {
		g1, _ := strconv.ParseInt(u1.Uid, 10, 32)
		gids[i] = uint32(g1)
	}
	return gids
}
