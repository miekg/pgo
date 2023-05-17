# PGO

Podman Git Operation.

## pgod


### Config file

``` toml
[[services]]
name = "bliep"
user = "miekg"
group = "miekg"
repository = "https://gitlab.science.ru.nl/bla/bliep"
branch = "bla"
urls = { "slashdot.org" = ":303" }
ports = [ "5005/5", "1025/5" ]
```

SSh keys must be put in a ssh/ directory.


## pgoctl
