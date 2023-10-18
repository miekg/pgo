all:
	( cd cmd/pgod; CGO_ENABLED=0 go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" )
	( cd cmd/pgoctl; CGO_ENABLED=0 go build -ldflags "-X main.version=`git tag --sort=-version:refname | head -n 1`" )

test:
	go test -v ./...

.PHONY: man
man:
	mmark -man cmd/pgod/pgod.8.md > cmd/pgod/pgod.8
	mmark -man cmd/pgoctl/pgoctl.1.md > cmd/pgoctl/pgoctl.1
