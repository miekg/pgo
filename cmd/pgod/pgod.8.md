%%%
title = "pgod 8"
area = "System Administration"
workgroup = "Docker Compose"
%%%

pgod
=====

## Name

pgod - watch a git repository, pull changes and restart the docker compose service

## Synopsis

`pgod [OPTION]...` `-c` **CONFIG**

## Description

`pgod` clones and pulls all repositories that are defined in the config file. It then exposes a SSH
interface (on port 2222) which you can interact with using pgoctl(1) or plain ssh(1) (not tested).

Each compose file runs under it's own user-account. That account can then access storage, or
databases it has access too - provisioning that stuff is out-of-scope - assuming your infra can deal
with all that stuff. And make that available on each server.

Servers running pgod(8) as still special in some regard, a developers needs to know which server
runs their compose file. Moving services to a different machine is as easy as starting the compose
there, but you need to make sure your infra also updates externals records (DNS for example).

The interface into `pgod` is via SSH, but not the normal SSH running on the server, this is a
completely seperate SSH interface implemented by both `pgod` and `pgoctl`.

For each repository it directs docker compose to pull, build and start the containers defined in the
`compose.yaml` file. Whenever this compose file changes this is redone. Current the following
compose file variants are supported: "compose.yaml", "compose.yml", "docker-compose.yml" and
"docker-compose.yaml".

With pgoctl(1) you can then interact with these services. You can "up", "down", "ps", "pull",
"logs", and "ping" currently. The syntax exposed is `<servicename>//<command>`, i.e. `pgo//ps`.

The options are:

**-c, --config string**
:  config file to read, when not given this default to `/etc/pgo.toml`

**-d, --dir string**
:  directory where to check out the git repositories, this must be a directory that is not wiped
   when the system reboots; the directory must also be accessible for all user accounts defined
   in the configuration file; this default to `/var/lib/pgo`.

**-s, --ssh string**
:  ssh address to listen on (default ":2222")

**-t, --duration duration**
:  default duration between pulls (default 5m0s)

**--debug**
:  enable debug logging

**--restart**
:   send SIGHUP to ourselves when config changes (default true)

**-v**, **--version**
:  show version and exit

## Config File

`pgod` requires a TOML config file where the services are defined, an example config file looks like
this:

~~~ toml
[[services]]
name = "pgo"
user = "miek"
repository = "https://github.com/miekg/pgo"
branch = "main"
registries = [ "user:token@registry" ]
ignore = false
compose = "my-compose.yaml"
env = [ "MYVAR=VALUE" ]
urls = { "example.org" = "pgo:5006" }
networks = [ "reverse_proxy" ]
import = "Caddyfile-import"
~~~

Here we define:

name
: `pgo`, how to address this service on this machine.

user
: `miek`, run docker under this user. This username only need to exist on the target machine and has
no relation to the SSH user connecting to `pgod`. I.e. it could be `nobody`.

repository *and* branch
: `https://github.com/miekg/pgo` and `main`, where to clone and pull from. If branch is not
specified `main` is assumed.

registries:
: `user:token@registry`, docker login credentials, this is used to login the registry and pull the
containers. Note that different credentials for the same user in different services might lead to
race conditions (and failed pulls). If the user part is not specifiied, the user from the `user`
keyword is used. This is a list because multiple private repositories are allowed.

ignore
: `false`, ignore changes to the compose yaml files and *do not* restart containers.

compose
: `my-compose.yaml`, specify an alternate compose file to use, outside of the supported variants.

env
: `"MYVAR=VALUE"`, specify environment variables to be exposed to the service.

urls
: `{ "example.org" = "pgo:5006" }` how to setup any forwarding to the listening ports.
but when the containers go up this should connect the url `example.org` to `<service>:5006`.

networks:
: `[ "reverse_proxy" ]`, allowed external networks. If empty all networks are allowed to be used.

import:
: `"Caddyfile-import"`, generate a Caddy (import) file that sets up the reverse proxies for *all*
services that are defined.

## Authentication

All remote access is authenticated and encrypted using SSH. The **public** keys you use *MUST* be
put in `ssh` subdirectory in the top level of your repository. The **private** key is used in
combination with pgoctl(1).

The generated key can't have a passphrase, to generate use: `ssh-keygen -t ed25519 -f ssh/id_pgo`.
And add and commit `ssh/id_pgo.pub`, and use `ssh/id_pgo` for authentication.

## Metrics

There are no metrics yet.

## Exit Code

Pgod has following exit codes:

0 - normal exit
1 - error seen (log.Fatal())
2 - SIGHUP seen (signal to systemd to restart us)

## See Also

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper). And see pgoctl(1) docker(1).
