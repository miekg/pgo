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
interface which you can interact with using `pgoctl` or plain `ssh`.

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
urls = { "slashdot.org" = ":303" }
ports = [ "5005/5", "1025/5" ]
~~~

## Interface

(ssh stuff)

## Metrics

There are no metrics yet.

## Exit Code

## See Also

See [this design doc](https://miek.nl/2022/november/15/provisioning-services/), and
[gitopper](https://github.com/miekg/gitopper).
