%%%
title = "pgoctl 1"
area = "System Administration"
workgroup = "Docker Compose"
%%%

pgoctl
=====

## Name

pgoctl - interact remotely with pgod(8)

## Synopsis

`pgoctl [OPTION]...` *host*:*name*//*commands*

## Description

pgoctl is an utility to inspect and control pgod(8) managed containers remotely. The *host* is the
machine running pgod(8), the *name* used must be the service name as configured in the configuration
for pgod(8). And the possible *commands* are listed below.

The exit status from the docker compose is reflected in the exist status of pgoctl. Almost all
commands from docker compose are implemented. Interactive commands, like starting a shell, are not
implemented. Tailing logs is also not implemented.

The supported commands are:

* `up` run `docker-compose up -d`
* `down` run `docker-compose down`
* `stop` run `docker-compose stop`
* `start` run `docker-compose start`
* `restart` run `docker-compose restart`
* `ps` run `docker-compose ps`
* `pull` run `docker-compose pull`
* `logs` run `docker-compose logs`
* `journal `run `journalctl _UID=<uid>` - show the system logs (if any)
* `exec` run `docker-compose -T exec` - run any command in a container
* `git` **COMMAND**
    where **COMMAND** can be:
    * `pull`, perform git pull
    * `hash`, show current hash of repo

All command also support arguments `pgoctl -i id_pgo -- localhost:pgo//journal --since yesterday`
for example.

There are only a few options:

**-i value**
: identity file to use for SSH, this flag is mandatory, but if an environment variable named
"PGOCTL_ID" exists and has a value, that value will be used as the private key identity. If no
such variable exist `-i` _is_ mandatory.

**--help, -h**
:  show help

**--port, -p port**
:  remote port number to use (defaults to 2222)

**-v**
:  show version and exit

Start pgod(8) and look at some services:

~~~
% sudo cmd/pgod/pgod -c pgo.toml -d /tmp/pgo
[INFO ] [caddy]: Service "caddy" with upstream "https://github.com/miekg/pgo-caddy"
[INFO ] [pgo]: Service "pgo" with upstream "https://github.com/miekg/pgo"
[INFO ] [caddy]: Launched tracking routine for "caddy"
[INFO ] [pgo]: Launched tracking routine for "pgo"
[INFO ] Launched server on port :9112 (prometheus)
[INFO ] Launched server on port :2222 (ssh) with 2 services tracked
[WARN ] [caddy]: Failed to get public keys: open /tmp/pgo/caddy/ssh: no such file or directory
[INFO ] [caddy]: Checked out git repo in /tmp/pgo/caddy for "caddy" (branch main) with 0 configured public keys
[INFO ] [caddy]: Writing Caddy import file "caddy/Caddyfile-import"
[INFO ] [pgo]: Checked out git repo in /tmp/pgo/pgo for "pgo" (branch main) with 3 configured public keys
[INFO ] [caddy]: Pulling containers
[INFO ] [pgo]: Pulling containers
[INFO ] [caddy]: Upping services
[INFO ] [pgo]: Upping services
[INFO ] [caddy]: Tracking upstream
[INFO ] [pgo]: Tracking upstream
~~~

Then up the services, if not done already:

~~~
% cmd/pgoctl/pgoctl -i ssh/id_pgo localhost:pgo//up
Container pgo-frontend-1  Running
~~~

Looking at the `ps`:

~~~
% cmd/pgoctl/pgoctl -i ssh/id_pgo localhost:pgo//ps
NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
pgo-frontend-1      docker.io/busybox   "/bin/busybox httpd …"   frontend            14 minutes ago      Up 14 minutes       0.0.0.0:32771->8080/tcp
~~~

Or `exec` inside a container/service. Docker compose expects the service to be used here, this is the
service *as specfied in the compose.yaml*.

~~~
% cmd/pgoctl/pgoctl -i id_pgo -- localhost:pgo//exec frontend /bin/ls
bin    etc    lib    proc   run    tmp    var
dev    home   lib64  root   sys    usr
~~~

## Bugs

Streaming responses are not implemented, i.e tailing a service log is currently not possible.

## Also See

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper). And see pgod(8) docker(1).
