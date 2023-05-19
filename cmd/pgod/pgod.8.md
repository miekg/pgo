%%%
title = "pgod 8"
area = "System Administration"
workgroup = "Podman Compose"
%%%

pgod
=====

## Name

pgod - watch a git repository, pull changes and restart the podman compose service

## Synopsis

`pgod [OPTION]...` `-c` **CONFIG**

## Description

`pgod` clones and pulls all repositories that are defined in the config file. It then exposes a SSH
interface (on port 2222) which you can interact with using pgoctl(1) or plain ssh(1) (not tested).

It then directs podman-compose to pull, build and start the containers defined in the
`docker-compose.yml` file. With pgoctl(1) you can then interact with these services. You can "up",
"down", "ps", "pull", "logs", and "ping" currently. The syntax exposed is
`<servicename>//<command>`, i.e. `pgo//ps`.

The options are:

**-c, --config string**
:  config file to read

**-s, --ssh string**
:  ssh address to listen on (default ":2222")

**-d, --debug**
:  enable debug logging

**-r, --restart**
:   send SIGHUP to ourselves when config changes

**-o, --root**
:  require root permission, setting to false can aid in debugging (default true)

**-t, --duration duration**
:  default duration between pulls (default 5m0s)

## Config File

`pgod` requires a TOML config file where the services are defined, an example config file looks like
this:

~~~ toml
[[services]]
name = "pgo"
user = "miek"  # under which user to run the podman
group = "miek" # which group to run the podman // not used atm
repository = "https://github.com/miekg/pgo"
branch = "main"
urls = { "example.org" = ":5006" }
ports = [ "5005/5", "1025/5" ]
~~~

Here we define:

name
: `pgo`, how to address this service on this machine.

user
: `miek`, run podman under this user. This username only need to exist on the target machine and has
no relation to the SSH user connecting to `pgod`. I.e. it could be `nobody`.

group
: `miek`, run podman with this group. Not used at the moment, the primary group of "user" is used.

repository *and* branch
: `https://github.com/miekg/pgo` and `main`, where to clone and pull from.

urls
: `{ "example.org" = ":5006" }` how to setup any forwarding to the listening ports. This isn't used yet,
but when the containers go up this should connect the url `example.org` to `<thismachine>:5006`.

ports
: `[ "5005/5", "1025/5" ]`, this service can bind to ports nummbers: 5005-5010 and 1025-1030. This
is checked by parsing the `docker-compose-yml`.

## Authentication

All remote access is authenticated and encrypted using SSH. The **public** keys you use *MUST* be
put in `ssh` subdirectory in the top level of your repository. The **private** is used in
combination with pgoctl(1).

The generated key can't have a passphrase, to generate use: `ssh-keygen -t ed25519 -f ssh/id_pgo`.
And add and commit `id_pgo.pub`.

## Metrics

There are no metrics yet.

## Exit Code

Pgod has following exit codes:

0 - normal exit
1 - error seen (log.Fatal())
2 - SIGHUP seen (signal to systemd to restart us)

## See Also

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper). And see pgoctl(1) podman-compose(1).
