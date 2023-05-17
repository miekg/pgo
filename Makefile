all:
	( cd cmd/pgod; go build )
	( cd cmd/pgoctl; go build )

.PHONY: man
man:
	mmark -man cmd/pgod/pgod.8.md > cmd/pgod/pgod.8
	mmark -man cmd/pgoctl/pgoctl.1.md > cmd/pgoctl/pgoctl.1
