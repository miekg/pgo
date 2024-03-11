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

`pgod` clones and pulls all repositories that are defined in the config file. It then starts all
containers using the compose file in each repository. Further more it exposes an SSH interface (on
port 2222) which you can interact with using pgoctl(1). Authentication is done via ssh public keys
that are available in each repository. Each *compose* runs under it's own user-account

Servers running pgod(8) as still special in some regard, a developers needs to know which server
runs their compose file. Moving services to a different machine is as easy as starting the compose
there, but you need to make sure your infra also updates externals records (DNS for example).

The interface into `pgod` is via SSH, but not OpenSSH, this is a completely seperate SSH interface
implemented by both `pgod` and `pgoctl`.

For each repository it directs docker compose to pull and start the containers defined in the
`compose.yaml` file. Whenever this compose file changes this is redone. Current the following
compose file variants are supported: "compose.yaml", "compose.yml", "docker-compose.yml" and
"docker-compose.yaml".

With pgoctl(1) you can then interact with these services. You can "up", "down", "ps", "pull",
"logs", and "ping" currently. The syntax exposed is `<servicename>//<command>`, i.e. `pgo//ps`.

On startup pgod(8) will down and remove any services that exist, but are not defined in the
confguration file.

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
compose = "my-compose.yaml"
env = [ "MYVAR=VALUE" ]
urls = { "example.org" = "pgo:5006" }
networks = [ "reverse_proxy" ]
import = "Caddyfile-import"
reload = "localhost:caddy//exec caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile"
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
services that are defined. If you have an `import` you also want to have a `reload`.

reload:
: `localhost:caddy//exec caddy --config Caddyfile --adapter caddyfile`, this is a pgoctl(1) exec
command line that runs `docker compose exec caddy caddy --config...`, to reload caddy in its
container. Note that the machine (here `localhost`) is not used, and could be anything.

## Reverse Proxy

Usually a Caddy server is run on the host port 443 (and 80 for Let's Encrypt TLS certificates
validation). This Caddy need to be able to reads the "import" file to be able to function. Only one
service needs to specify this file, and usually this is the caddy service (that can also be managed
by pgod).

## Authentication

All remote access is authenticated and encrypted using SSH. The **public** keys you use *MUST* be
put in `ssh` subdirectory in the top level of your repository. The **private** key is used in
combination with pgoctl(1) and should *never be checked in*.

The generated key can't have a passphrase, to generate use: `ssh-keygen -t ed25519 -f ssh/id_pgo`.
And add and commit `ssh/id_pgo.pub`, and use `ssh/id_pgo` for authentication.

## Metrics

Two metrics are exported:

* `pgo_command_count`: total of commands executed
* `pgo_command_error_count`: count of errors resulting from command execution

## Exit Code

pgod(8) has following exit codes:

0 - normal exit
1 - error seen (log.Fatal())
2 - SIGHUP seen (signal to systemd to restart us)

## See Also

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper). And see pgoctl(1) docker(1).
