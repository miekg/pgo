%%%
title = "pgoctl 1"
area = "System Administration"
workgroup = "Podman Compose"
%%%

pgoctl
=====

## Name

pgoctl - interact remotely with pgod(8)

## Synopsis

`pgoctl [OPTION]...` *host*:*name*//*commands*

## Description

pgoctl is an utility to inspect and control pgod(8) remotely.

There are only a few options:

**-i value**
: identity file to use for SSH, this flag is mandatory. You can use `'$VAR'` then the private key
will be read from the environment variable `VAR`. Note the "$" must be the first character of value
and must be quoted, i.e. it should not be expanded by the shell.

**--help, -h**
:  show help

**--port, -p port**
:  remote port number to use (defaults to 2222)

Start pgod(8) and look at some services:

~~~
% sudo ./cmd/pgod/pgod -c config.toml
2023/05/19 08:34:47 [INFO ] Service "pgo" with upstream "https://github.com/miekg/pgo"
2023/05/19 08:34:47 [INFO ] Launched tracking routine for "pgo"
2023/05/19 08:34:47 [INFO ] Launched servers on port :2222 (ssh)
~~~

Then up the services, if not done already:

~~~
% cmd/pgoctl/pgoctl -i ssh/id_pgo4 localhost:pgo//up
61380c3c0cbe9827f335b5d6e7690d3a366317f755d87f969fcd9b1cb4b2254c
['podman', '--version', '']
using podman version: 3.4.4
** excluding:  set()
['podman', 'network', 'exists', 'pgo-3493677287_default']
podman run --name=pgo-3493677287_frontend_1 ..... -p 8080 -w / busybox /bin/busybox httpd -f -p 8080
exit code: 0%
~~~

Looking at the `ps`:

~~~
% cmd/pgoctl/pgoctl -i ssh/id_pgo4 localhost:pgo//ps
CONTAINER ID  IMAGE                             COMMAND               CREATED             STATUS                 PORTS                    NAMES
61380c3c0cbe  docker.io/library/busybox:latest  /bin/busybox http...  About a minute ago  Up About a minute ago  0.0.0.0:36391->8080/tcp  pgo-3493677287_frontend_1
['podman', '--version', '']
using podman version: 3.4.4
podman ps -a --filter label=io.podman.compose.project=pgo-3493677287
exit code: 0%
~~~

## Also See

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper). And see pgod(8) podman-compose(1).
