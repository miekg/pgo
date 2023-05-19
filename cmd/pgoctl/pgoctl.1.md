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

# EVERY BELOW HERE NEEDS UPDATING FOR PGOCTL - NOW SHOWS GITOPPER

~~~
./gitopperctl -i ~/.ssh/id_ed25519_gitopper list machine @<host>
./gitopperctl list service @<host>
./gitopperctl list service  @<host> <service>
~~~

In order:

1. List all machines defined in the config file for gitopper running on `<host>`.
2. List all services that are controlled on `<host>`.
3. List a specific service on `<host>`.

Each will output a simple table with the information:

~~~
./gitopperctl list service @localhost grafana-server
SERVICE         HASH      STATE  INFO  SINCE
grafana-server  606eb576  OK           2022-11-18 13:29:44.824004812 +0000 UTC
~~~

Use `--help` to show implemented subcommands.

### Manipulating Services

## Example

This is a small example of this tool interacting with the daemon.

- check current service

~~~
./gitopperctl list service @localhost grafana-server
SERVICE         HASH      STATE  INFO  SINCE
grafana-server  606eb576  OK           0001-01-01 00:00:00 +0000 UTC
~~~

-  rollback

~~~
./gitopperctl do rollback @localhost grafana-server 8df1b3db679253ba501d594de285cc3e9ed308ed
~~~

- check
~~~
./gitopperctl list service @localhost grafana-server
SERVICE         HASH      STATE     INFO                                      SINCE
grafana-server  606eb576  ROLLBACK  8df1b3db679253ba501d594de285cc3e9ed308ed  2022-11-18 13:28:42.619731556 +0000 UTC
~~~

- check do, rollback done. Now state is FREEZE

~~~
./gitopperctl list service @localhost grafana-server
SERVICE         HASH      STATE   INFO                                                      SINCE
grafana-server  8df1b3db  FREEZE  ROLLBACK: 8df1b3db679253ba501d594de285cc3e9ed308ed  2022-11-18 13:29:17.92401403 +0000 UTC
~~~

- unfreeze and let it pick up changes again

~~~
./gitopperctl do unfreeze @localhost grafana-server
~~~

- check the service

~~~
./gitopperctl list service @localhost grafana-server
SERVICE         HASH      STATE  INFO  SINCE
grafana-server  8df1b3db  OK           2022-11-18 13:29:44.824004812 +0000 UTC
~~~

- and updated to new hash

~~~
./gitopperctl list service @localhost grafana-server
SERVICE         HASH      STATE  INFO  SINCE
grafana-server  606eb576  OK           2022-11-18 13:29:44.824004812 +0000 UTC
~~~

## Also See

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper). And see pgod(8) podman-compose(1).
