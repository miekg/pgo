# PGO

Podman Gitops. This is a subsequent development (or successor?) of
<https://github.com/miekg/gitopper>. Where "gitopper" integrates with your OS, i.e. use Debian
packages, "pgo" uses a `compose.yaml` as its basis. It runs the compose via `docker compose` (or
docker-compose) (podman was dropped, see below, see https://docs.docker.com/engine/install/debian/
for docker's installation). It allows for remote interaction via an SSH interface, which `pgoctl`
makes easy to use. For this SSH interface no local users need to exist on the target system. The
compose is (usually) executed under a particular user name, those users do need an account on the
target system, they do not need shell access.

Current the following compose file variants are supported: "compose.yaml", "compose.yml",
"docker-compose.yml" and "docker-compose.yaml". If you need more flexibility you can point to a
specific compose file, with the `compose` key in the config.

Each compose file runs under it's own user-account. That account can then access storage, or
databases it has access to - provisioning that stuff is out-of-scope - assuming your infra can deal
with all that stuff. And make that available on each server.

Servers running "pgod" as still special in some regard, a developers needs to know which server runs
their compose file. Moving services to a different machine is as easy as starting the compose there,
but you need to make sure your infra also updates external records (DNS for example), but maybe also
ACL on the databases and other resources.

The interface into `pgod` is via SSH, but not the normal SSH running on the server, this is a
completely seperate SSH interface implemented by both `pgod` and `pgoctl`. The owner of the Git repo
can publish new (public) keys and make `pgoctl` access available for anyone with access the private
key.

The main idea here is that developers can push stuff easier to production and that you can have some
of the goodies from Kubernetes, but not the bad stuff like the networking - the big trade-off being
that machines are still somewhat special.

Note that pgod runs (as root) under systemd.

A typical config file looks like this:

``` toml
[[services]]
name = "pgo"
user = "miek"
repository = "https://github.com/miekg/pgo"
registry = "user:authtoken" # or just authtoken
compose = "compose.yaml"
branch = "main"
ignore = false
env = [ "MYENV=bla", "OTHERENV=bliep"]
urls = { "pgo.science.ru.nl" = "pgo:5007" }
networks = [ "reverse_proxy" ]
# import = "Caddyfile-import"
```

This file is used by `pgod` and should be updated for each project you want to onboard. Our plan is
to have this go through an on boarding workflow and automate it.

To go over this file:

- `name`: this is the name of the service, used to uniquely identify the service across machines.
- `user`: which user to use to run the docker compose under.
- `repository` and `branch`: where to find the git repo belonging to this service.
- `registry`: optional authentication for pulling the docker images from the registry. In
  "user:token" format, is user is omitted, `user` is used.
- `compose`: alternate compose file to use.
- `ignore`: don't restart the containers when a compose file changes.
- `urls`: what DNS names need to be assigned to this server and to what network and port should they forward.
- `networks`: which external network can this service use. Empty means all.
- `env`: specify extra environment variables in "VAR=VALUE" notation (i.e. secrets).
- `import`: create a Caddyfile snippet with reverse proxy statements for all URLs in all services
  *with a specific* prefix and writes this in the directory where the repository is checked out.

For non-root accounts, docker compose will be run with the normal supplementary groups to which the
local docker group has been added. This allows those user to transparently access the docker socket.

## Requisites

To use "pgo" your project MUST have:

- A public SSH key (or keys) stored in a `ssh/` directory in your git repo. This keys can **not** have a
  passphrase protecting them
- A `compose.yaml` (or any of the variants) in the top-level of your git repo.

## Quick Start

Assuming a working Go compiler you can issue a `make` to compile the binaries. Then:

Start `pgod`: `sudo /cmd/pgod/pgod -c pgo.toml -d /tmp/pgo --debug`. That will output some debug
data.

In other words: it clones the repo, builds, pulls, and starts the containers. It then *tracks*
upstream and whenever `compose.yaml` changes it will do a `down` and `up`. To force changes
in that file you can use a `x-gpo-version` in the yaml and change that whenever you want to update
"pgo"

Now with `pgoctl` you can access and control this environment (well not you, because you don't have
the private key belonging to the public key that sits in the `ssh/` directory). `pgoctl` want to
see `<machine>:<name>//<operation>` string, i.e. `localhost:pgo//ps` which does a `docker compose
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
NAME                IMAGE               COMMAND                  SERVICE             CREATED             STATUS              PORTS
pgo-frontend-1      docker.io/busybox   "/bin/busybox httpd …"   frontend            9 minutes ago       Up 9 minutes        0.0.0.0:32771->8080/tcp
~~~

Currently implemented are: `up`, `down`, `pull`, `ps`, `logs` and `ping` to see if the
authentication works (replies with a "pong!" if everything works).

## Integrating with GitLab Environments

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

If composers need a network, you'll need to set this up by yourself with Caddy, pgod has support to
write a Caddyfile snippet that routes all URLs to the composer's backends. This does mean the
caddy's docker-compose must be setup in such a way that it will read that file *and* configures a
"well-known" network, where other composers can hook into. The setup we use `caddy` as the name
for the service *and* the network. This is defined in the <https://github.com/miekg/pgo-caddy> project.

The services that are exporting into the caddy snippet must begin with the service name, so if you
would like `frontend` (in the pgo service) to receive traffic from caddy, you must name the service
'pgo-frontend', and in urls you must have: "domain" = "pgo-frontend:port" for it to work.

The pgod config would look like this:
~~~ toml
[[services]]
name = "caddy"
user = "root"
repository = "https://github.com/miekg/pgo-caddy"
import = "caddy/Caddyfile-import"
~~~

And the compose.yaml:

~~~ yaml
version: '3.6'
services:
  caddy:
    image: docker.io/caddy:2.6-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./caddy/:/etc/caddy/
    networks:
      - caddynet

networks:
    caddynet:
      name: caddy
~~~

Where the network created is called `caddy` and the `Caddyfile-import` is written into the
directory that gets mounted as a volume inside the caddy container. The pgo-caddy git repository has
an .gitignore for `caddy/Caddyfile-import`.

Other users of this PGO instance only need to know the network is called `caddy` and commit their
pgo.toml config to make things work.

## pgod

See the [manual page](./cmd/pgod/pgod.8.md) in [cmd/pgod](./cmd/pgod/).

## pgoctl

See the [manual page](./cmd/pgoctl/pgoctl.1.md) in [cmd/pgoctl](./cmd/pgoctl). Also details the
`PGOCTL_ID` mentioned above.

# Podman and Podman-Compose

Initialy PGO was using podman(-compose) to run the images, but this proved to be a challenge.
podman-compose is a seperate project and has it's own ideas on how to parse a compose.yml file (not
only his fault, the format is terrible). But using external network just didn't work, regardless
what syntax was used. Also podman kept complaining about CNI version clashes which were
undebuggable, so as much as I want to like podman, this is now using docker compose.

Also in podman 4 the networking moved away from CNI to a new thing written in Rust - which is
completely fine, but does raise the possibility that I can revist networking relatively soon again
to fix it for podman4.

Also podman-compose has not seen much releases, so the apt-get install story becomes weaker there as
well. Initial experiments with docker made stuff work out of the box.
