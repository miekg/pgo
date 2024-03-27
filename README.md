# PGO

Podman Gitops. This is a subsequent development (or successor?) of
<https://github.com/miekg/gitopper>. Where "gitopper" integrates with your OS, i.e. use Debian
packages, "pgo" uses a `compose.yaml` as its basis. It runs the compose via `docker compose` (or
docker-compose) (podman was dropped, see below, see https://docs.docker.com/engine/install/debian/
for docker's installation). It allows for remote interaction via an SSH interface, which `pgoctl`
makes easy to use. For this SSH interface no local users need to exist on the target system. The
compose is (usually) executed under a particular user name, those users do need an account on the
target system, they do not need shell access, because everything runs through SSH of pgod(8).

The main idea here is that developers can push stuff easier to production and that you can have some
of the goodies from Kubernetes, but not the bad stuff like the networking - the big trade-off being
that machines are still somewhat special.

Current the following compose file variants are supported: "compose.yaml", "compose.yml",
"docker-compose.yml" and "docker-compose.yaml". If you need more flexibility you can point to a
specific compose file, with the `compose` key in the config.

Each compose file runs under it's own user-account. That account can then access storage, or
databases it has access to - provisioning that stuff is out-of-scope - assuming your infra can deal
with all that. The compose file is parsed and the following settings are *disallowed*:

- privileged=true
- network_mode=host
- ipc=host
- volumes can only reference an absolute path (specified in pgo.toml).
- volumes can only be of type 'bind'

Servers running pgod(8) as still special in some regard, as a developers needs to know which server runs
their compose file. Moving services to a different machine is as easy as starting the compose there,
but you need to make sure your infra also updates external records (DNS for example).

The interface into pgod(8) is via custom SSH implementation that is separate from the machine's SSH.
The owner of the Git repo can publish public keys in a `ssh` direcotry and their repo and give
pgoctl(1)-access tp anyone with access the correspondong private key.

Note that pgod runs (as root) under systemd.

A typical config file looks like this:

``` toml
[[services]]
name = "pgo"
user = "miek"
repository = "https://github.com/miekg/pgo"
registries = [ "user:authtoken@registry" ] # or just authtoken@registry
compose = "compose.yaml"
branch = "main"
env = [ "MYENV=bla", "OTHERENV=bliep"]
urls = { "pgo.science.ru.nl" = "pgo:5007" }
networks = [ "reverseproxy" ]
# import = "Caddyfile-import"
# reload = "localhost:caddy//exec caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile"
# mount =  "nfs://server.example.org/share"
```

This file is used by `pgod` and should be updated for each project you want to onboard. To go over
this file:

- `name`: this is the name of the service, used to uniquely identify the service across machines.
- `user`: which user to use to run the docker compose under.
- `repository` and `branch`: where to find the git repo belonging to this service.
- `registries`: optional authentication for pulling the docker images from the registry. In
  "user:token" format, is user is omitted, `user` is used. This is a list because there can be more
  than one private registry. This should match any registries used in the compose file.
- `compose`: alternate compose file to use.
- `urls`: what DNS names need to be assigned to this server and to what network and port should they forward.
- `env`: specify extra environment variables in "VAR=VALUE" notation (i.e. secrets).
- `networks`: which external network can this service use. Empty means all.
- `import`: create a Caddyfile snippet with reverse proxy statements for all URLs in all services
  and write this in the directory where the repository is checked out.
- `reload`: a exec command in pgoctl(1) syntax to reload caddy when a new import file is written.
- `mount`: specific a NFS volume that will be mounted in `<datadir>/<name>`, see pgod(8). This NFS mount gets
  mounted with default options: "rw,nosuid,hard".

For non-root accounts, docker compose will be run with the normal supplementary groups to which the
*local* docker group has been added. This allows those user to transparently access the docker
socket, without going through some addgroup(8) hassle.

### Compose File Extensions

On every change to the compose file, pgod(8) will down and up your services. If you do not want this
added the following your compose file (as a top-level declaration). The default for reload is
`true`.

~~~ yaml
x-pgo:
  reload: false
~~~

## Requisites

To use "pgo" your project should have:

- A public SSH key (or keys) stored in a `ssh/` directory in your git repo. This keys can **not** have a
  passphrase protecting them. If there are no keys, or no ssh directory pgoctl(1) will not work.
- A `compose.yaml` (or any of the variants) in the top-level of your git repo.

## Quick Start

Assuming a working Go compiler you can issue a `make` to compile the binaries. Then:

Start `pgod`: `sudo /cmd/pgod/pgod -c pgo.toml -d /tmp/pgo --debug`. That will output some debug
data.

In other words: it clones the repo, pulls, and starts the containers. It then *tracks*
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

If you want to use pgo with GitLab you needs to setup
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
    - pgoctl mymachine:project//pull  # looks for PGOCTL_ID env var, with privkey contents
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

If services need a network, you'll need to set this up by yourself with Caddy, pgod(8) has support to
write a Caddyfile snippet that routes all URLs to the composer's backends. This does mean the
caddy's docker-compose must be setup in such a way that it will read that file *and* configures a
"well-known" network, where other composers can hook into. A setup you can use is having `caddy` as
the name for the service *and* the network. This is defined in the
<https://github.com/miekg/pgo-caddy> project. Other service need to reference this as an external netwerk.

The services that are exporting into the caddy snippet only need an "url" in their config. So the
`pgo-caddy` config is a normal pgo service and has this compose config:

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
      - reverseproxy

networks:
    reverseproxy:
      name: reverseproxy
~~~

Where the network created is called `reverseproxy` and the `Caddyfile-import` is written, by pgod(8), into
the directory that gets mounted as a volume inside the caddy container. The `pgo-caddy` git
repository has an .gitignore for `caddy/Caddyfile-import`.

Other users of this pgo instance only need to know the network is called `caddy` and commit their
pgo.toml config to make things work, i.e. in those `compose.yaml` they need:

~~~ yaml
networks:
  reverseproxy:
    external: true
    name: reverseproxy
  myownnetwork:
~~~

in each service that need a reverse proxy, and then _also_ a `urls` section in `pgo.toml` that ties
everything together: `urls = { "example.org": frontend:8080" }`. The lone `myownnetwork` is needed
because we want every "compose" to be in its own network. This might be enforced in the future.

## pgod

See the [manual page](./cmd/pgod/pgod.8.md) in [cmd/pgod](./cmd/pgod/).

## pgoctl

See the [manual page](./cmd/pgoctl/pgoctl.1.md) in [cmd/pgoctl](./cmd/pgoctl). Also details the
`PGOCTL_ID` mentioned above.

# Podman and Podman-Compose

Initialy pgo was using podman(-compose) to run the images, but this proved to be a challenge.
podman-compose is a seperate project and has it's own ideas on how to parse a compose.yml file (not
only his fault, the format is terrible). But using external network just didn't work, regardless
what syntax was used. Also podman kept complaining about CNI version clashes which were
undebuggable, so as much as I want to like podman, this is now using docker compose.

Also in podman 4 the networking moved away from CNI to a new thing written in Rust - which is
completely fine, but does raise the possibility that I can revist networking relatively soon again
to fix it for podman4.

Also podman-compose has not seen much releases, so the apt-get install story becomes weaker there as
well. Initial experiments with docker made stuff work out of the box.
