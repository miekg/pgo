all:
	( cd cmd/pgod; go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" )
	( cd cmd/pgoctl; go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" )

.PHONY: man
man:
	mmark -man cmd/pgod/pgod.8.md > cmd/pgod/pgod.8
	mmark -man cmd/pgoctl/pgoctl.1.md > cmd/pgoctl/pgoctl.1
