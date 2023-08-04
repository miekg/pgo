# Using PGO

## pgoctl

Any command you execute with `pgoctl ... exec` is handled directly by `pgod` and given to docker
compose exec. At this point (just before compose exec) one shell has parsed the tokens and that is
your local shell. When docker exec is execute there is also no shell, so redirecting output -- even
with the correct quoting to protect from your local shell - will not work.

Also note that interactive exec commands are not supported, but may work at some point in the
future.

## Environment Variables

Any environent variables containing secrets need to be put in the `pgo.toml` configuration file. To
have access to them inside your compose you also need to put `environment` statements in your
`compose.yaml`. After that the compose needs to be restarted to pick them up.

## Storage
