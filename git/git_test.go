package git

import (
	"testing"
)

func TestDiffStatOK(t *testing.T) {
	g := New("test", "", "", ".", "")

	data := []byte(`remote: Enumerating objects: 10, done.
remote: Counting objects: 100% (10/10), done.
remote: Compressing objects: 100% (9/9), done.
remote: Total 9 (delta 4), reused 0 (delta 0), pack-reused 0
Unpacking objects: 100% (9/9), 4.80 KiB | 1.20 MiB/s, done.
From deb.atoom.net:/git/miek/docs
 * branch            master     -> FETCH_HEAD
   37a1ec8..7e019a1  master     -> origin/master
Updating 37a1ec8..7e019a1
Fast-forward
 my/stuff/file.md | 139 +++++++++++++++++++++++++++
 1 file changed, 139 insertions(+)
 create mode 100644 provisioning-systems.md
`)

	if !g.OfInterest(data, []string{"my/stuff/file.md"}) {
		t.Fatal("Expected to find paths of interest, got none")
	}
}

func TestDiffStatFail(t *testing.T) {
	g := New("test", "", "", ".", "")

	data := []byte(`remote: Enumerating objects: 10, done.
remote: Counting objects: 100% (10/10), done.
remote: Compressing objects: 100% (9/9), done.
remote: Total 9 (delta 4), reused 0 (delta 0), pack-reused 0
Unpacking objects: 100% (9/9), 4.80 KiB | 1.20 MiB/s, done.
From deb.atoom.net:/git/miek/docs
 * branch            master     -> FETCH_HEAD
   37a1ec8..7e019a1  master     -> origin/master
Updating 37a1ec8..7e019a1
Fast-forward
 my/stuff/file.md | 139 +++++++++++++++++++++++++++
 1 file changed, 139 insertions(+)
 create mode 100644 provisioning-systems.md
`)

	if g.OfInterest(data, []string{"/other/stuff"}) {
		t.Fatal("Expected to find _no_ paths of interest, but got some")
	}
}
