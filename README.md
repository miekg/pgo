# PGO

Podman Git Operation. This is a subsequent development (or successor?) of
<https://github.com/miekg/gitopper>. Where "gitopper" integrates with your OS, i.e. use Debian
packages, "pgo" uses a `docker-compose.yml` as it's basis. It runs the compose via `podman-compose`
(Debian package exists). It allows for remote interaction via an SSH interface, which `pgoctl` makes
easy to use.

Each compose file runs under it's own user-account. That account can then access storage, or
databases it has access too - provisioning that stuff is out-of-scope - assuming your infra can deal
with all that stuff. And make that available on each server.

Servers running "pgo" as still special in some regard, a developers needs to know which server runs
their compose file *and* you need to administrate who own which port numbers. Moving services to a
different machine is as easy as starting the compose there, but you need to make sure your infra
also updates externals records (DNS for example).

The main idea here is that developers can push stuff easier to production and that you can have some
of the goodies from Kubernetes, but not that bad stuff like the networking - the big trade-off being
you need to administrate port numbers *and* still run some proxy to forward URLs to the correct
backend.

A typical config file looks like this:

``` toml
[[services]]
name = "pgo"
user = "miek"
repository = "https://github.com/miekg/pgo"
branch = "main"
urls = { "pgo.science.ru.nl" = ":5007" }
ports = [ "5005/5", "1025/5" ]
```

This file is used by `pgod` and should be updated for each project you want to onboard. Our plan is
to have this go through an onboarding workflow.

To go over this file:

- `name`: this is the name of the service, used to uniquely identify the service across machines.
- `user`: which user to use to run the podman-compose under.
- `repository` and `branch`: where to find the git repo belonging to this service
- `urls`: what DNS names need to be assigned to this server and to what port should they forward.
- `ports`: which ports can this service bind to.

## Requisites

To use "pgo" your project MUST have:

- Public SSH keys stored in a `ssh/` directory in your git repo.
- A `docker-compose.yml` in the top-level of your git repo.

## Quick Start

Assuming a working Go compiler you can issue a `make` to compile the binaries. Then.

Start `pgod`: `sudo ./cmd/pgod/pgod -c config.toml -d`. That will output some debug data, condensed
here:

~~~ txt
2023/05/17 20:13:20 [INFO ] Service "pgo" with upstream "https://github.com/miekg/pgo"
2023/05/17 20:13:20 [INFO ] Launched tracking routine for "pgo"
2023/05/17 20:13:20 [INFO ] Launched servers on port :2222 (ssh)
2023/05/17 20:13:20 [DEBUG] running in "/tmp/pgo-3809413984" as "miek" [git clone -b main https://github.com/miekg/pgo /tmp/pgo-3809413984]
2023/05/17 20:13:20 [DEBUG] Cloning into '/tmp/pgo-3809413984'...
2023/05/17 20:13:20 [INFO ] Checked out git repo in /tmp/pgo-3809413984 for "pgo"
2023/05/17 20:13:20 [DEBUG] running in "/tmp/pgo-3809413984" as "miek" [podman-compose build] (env: [HOME=/home/miek PATH=/usr/sbin:/usr/bin:/sbin:/bin])
2023/05/17 20:13:21 [DEBUG] ['podman', '--version', '']
using podman version: 3.4.4
2023/05/17 20:13:21 [DEBUG] running in "/tmp/pgo-3809413984" as "miek" [podman-compose up -d] (env: [HOME=/home/miek PATH=/usr/sbin:/usr/bin:/sbin:/bin])
~~~

In other words: it clones the repo, builds, pulls, and starts the containers. It then *tracks*
upstream and whenever `docker-compose.yml` changes it will do an `down` and `up`. To force changes
in that file you can use a `x-gpo-version` in the yaml and change that whenever you want to update
"pgo"

Now with `pgoctl` you can access and control this environment (well not you, because you don't have
the private key belonging to the public key that sits in the `ssh/` directory). `pgoctl` want to
see `<machine>:<name>//<operation>` string, i.e. `localhost:pgo//ps` which does a `podman-compose
ps` for our stuff:

~~~
# ask for the status of pgo - denied because the correct key is not found in the repos
% ./cmd/pgoctl/pgoctl -i ~/id_pgo2 localhost:pgo//ps
Unauthorized: Key for user "miek" does not match any for name pgo
2023/05/17 20:21:08 [ERROR] Process exited with status 401
~~~

Once our committed keys get pulled:
~~~
% ./cmd/pgoctl/pgoctl -i ~/id_pgo2 localhost:pgo//ps
CONTAINER ID  IMAGE                             COMMAND               CREATED        STATUS            PORTS                    NAMES
4fe30f61c4db  docker.io/library/busybox:latest  /bin/busybox http...  3 seconds ago  Up 3 seconds ago  0.0.0.0:40475->8080/tcp  pgo-609353550_frontend_1
['podman', '--version', '']
using podman version: 3.4.4
podman ps -a --filter label=io.podman.compose.project=pgo-609353550
exit code: 0%
~~~

Currently implemented are: `up`, `down`, `pull`, `ps`, `logs` and `ping` to see if the
authentication works.

## pgod

See the (soon to be created) manual page in cmd/pgod.

## pgoctl

See the (soon to be created) manual page in cmd/pgoctl.

# TODO

- Tailing logs with -f.
