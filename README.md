# PGO

Podman Git Operation.

## pgod


### Config file

``` toml
[[services]]
name = "bliep"
user = "miekg"
group = "miekg"
git = "https://gitlab.science.ru.nl/bla/bliep"
urls = { "slashdot.org" = ":303" }
ports = [ "5005/5", "1025/5" ]
```


## pgoctl
