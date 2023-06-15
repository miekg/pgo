# PGO

Podman Gitops. This is a subsequent development (or successor?) of
<https://github.com/miekg/gitopper>. Where "gitopper" integrates with your OS, i.e. use Debian
packages, "pgo" uses a `compose.yaml` as it's basis. It runs the compose via `podman-compose`
(Debian package exists). It allows for remote interaction via an SSH interface, which `pgoctl` makes
easy to use. For this SSH interface no local users need to exist on the target system.

You can restrict which external networks can be used.

Current the following compose file variants are supported: "compose.yaml", "compose.yml",
"docker-compose.yml" and "docker-compose.yaml". If you need more flexibility you can point to a
specific compose file.

Each compose file runs under it's own user-account. That account can then access storage, or
databases it has access to - provisioning that stuff is out-of-scope - assuming your infra can deal
with all that stuff. And make that available on each server.

Servers running "pgo" as still special in some regard, a developers needs to know which server runs
their compose file *and* you need to administrate who own which port numbers. Moving services to a
different machine is as easy as starting the compose there, but you need to make sure your infra
also updates externals records (DNS for example).

The interface into `pgod` is via SSH, but not the normal SSH running on the server, this is a
completely seperate SSH interface implemented by both `pgod` and `pgoctl`.

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
compose = "compose.yaml"
branch = "main"
ignore = false
env = [ "MYENV=bla", "OTHERENV=bliep"]
urls = { "pgo.science.ru.nl" = "pgo:5007" }
networks = [ "reverse_proxy" ]
# import = "Caddyfile-import"
```

This file is used by `pgod` and should be updated for each project you want to onboard. Our plan is
to have this go through an on boarding workflow.

To go over this file:

- `name`: this is the name of the service, used to uniquely identify the service across machines.
- `user`: which user to use to run the podman-compose under.
- `repository` and `branch`: where to find the git repo belonging to this service.
- `compose`: alternate compose file to use.
- `ignore`: don't restart podman when a compose file changes.
- `urls`: what DNS names need to be assigned to this server and to what port should they forward.
- `networks`: which external network can this service use. Empty means all.
- `env`: specify extra environment variables in "VAR=VALUE" notation.
- `import`: create a Caddyfile snippet with reverse proxy statements for all URLs in all services.

## Requisites

To use "pgo" your project MUST have:

- A public SSH key (or keys) stored in a `ssh/` directory in your git repo. This keys can **not** have a
  passphrase protecting them
- A `compose.yaml` (or any of the variants) in the top-level of your git repo.

## Quick Start

Assuming a working Go compiler you can issue a `make` to compile the binaries. Then:

Start `pgod`: `sudo ./cmd/pgod/pgod -c pgo.toml -d /tmp --debug`. That will output some debug
data, condensed here:

~~~ txt
[INFO ] Service "pgo" with upstream "https://github.com/miekg/pgo"
[INFO ] Launched tracking routine for "pgo"
[INFO ] Launched servers on port :2222 (ssh)
[DEBUG] running in "/tmp/pgo-pgo" as "miek" [git clone -b main https://github.com/miekg/pgo /tmp/pgo-3809413984]
[DEBUG] Cloning into '/tmp/pgo-pgo'...
[INFO ] Checked out git repo in /tmp/pgo-pgo for "pgo"
[DEBUG] running in "/tmp/pgo-pgo" as "miek" [podman-compose build] (env: [HOME=/home/miek PATH=/usr/sbin:/usr/bin:/sbin:/bin])
[DEBUG] ['podman', '--version', '']
using podman version: 3.4.4
[DEBUG] running in "/tmp/pgo-pgo" as "miek" [podman-compose up -d] (env: [HOME=/home/miek PATH=/usr/sbin:/usr/bin:/sbin:/bin])
~~~

In other words: it clones the repo, builds, pulls, and starts the containers. It then *tracks*
upstream and whenever `compose.yaml` changes it will do a `down` and `up`. To force changes
in that file you can use a `x-gpo-version` in the yaml and change that whenever you want to update
"pgo"

Now with `pgoctl` you can access and control this environment (well not you, because you don't have
the private key belonging to the public key that sits in the `ssh/` directory). `pgoctl` want to
see `<machine>:<name>//<operation>` string, i.e. `localhost:pgo//ps` which does a `podman-compose
ps` for our stuff:

~~~
# ask for the status of pgo - denied because the correct key is not found in the repo
% ./cmd/pgoctl/pgoctl -i ~/id_pgo2 localhost:pgo//ps
Unauthorized: Key for user "miek" does not match any for name pgo
2023/05/17 20:21:08 [ERROR] Process exited with status 401
~~~

Once our committed keys get pulled:
~~~
% ./cmd/pgoctl/pgoctl -i ~/id_pgo2 localhost:pgo//ps
CONTAINER ID  IMAGE                             COMMAND               CREATED        STATUS            PORTS                    NAMES
4fe30f61c4db  docker.io/library/busybox:latest  /bin/busybox http...  3 seconds ago  Up 3 seconds ago  0.0.0.0:40475->8080/tcp  pgo-pgo_frontend_1
['podman', '--version', '']
using podman version: 3.4.4
podman ps -a --filter label=io.podman.compose.project=pgo-pgo
exit code: 0%
~~~

Currently implemented are: `up`, `down`, `pull`, `ps`, `logs` and `ping` to see if the
authentication works. With `ping` you can check if the authentication is setup correctly, you should
see a "pong!" reply if everything works.

## Integrating with GitLab

If you want to use PGO with GitLab you needs to setup
[environments](https://docs.gitlab.com/ee/ci/environments/) that allow you to deploy to
"production", here is an example `.gitlab-ci.yml` that does this:

~~~ yaml
image: "registry.science.ru.nl/cncz/sys/image/cncz-debian-go:latest"

stages:
  - deploy

deploy_production:
  resource_group: production
  stage: deploy
  environment:
    name: production
    url: https://example.com
    on_stop: stop_production
  script:
    - pgoctl mymachine:project//pull  # looks for PGOCTL_ID env var
    - pgoctl mymachine:project//build
    - pgoctl mymachine:project//up
  when: manual

stop_production:
  resource_group: production
  stage: deploy
  script:
    - pgoctl mymachine:project//down
  environment:
    name: production
    action: stop
  when: manual
~~~

With 'manual' you can still control when this actually happens.

If you want to clone a repository that is private, you can create an access token with
'read_repository' and the "developer" role. This can be then used as:

~~~ toml
repository = "https://oauth2:<token>@gitlab.science.ru.nl/..."
~~~

## Networking and Reverse Proxy

Default how to do a caddy setup.

## pgod

See the [manual page](./cmd/pgod/pgod.8.md) in [cmd/pgod](./cmd/pgod/).

## pgoctl

See the [manual page](./cmd/pgoctl/pgoctl.1.md) in [cmd/pgoctl](./cmd/pgoctl).
