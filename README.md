# Carpal

A small and configurable [WebFinger](https://webfinger.net/) server.

## Installation (Docker)

A Docker image is provided, and can be run like so:

``` sh
docker run --rm -it \
    -v ./config:/etc/carpal \
    -p 8008:8008 \
    peeley/carpal:latest
```

## Installation (Source)

TODO

## Configuration

The configuration file defaults to `/etc/carpal/config.yml`, and looks like so
by default:

``` yaml
# /etc/carpal/config.yml

driver: file
file:
  directory: /etc/carpal/resources/
```

Resource files should be named after the resource they describe. For example,
the data for a resource named `acct:bob@foobar.com` should reside in
`/etc/carpal/resources/acct:bob@foobar.com`. The resource file might look like
the following:

``` yaml
# /etc/carpal/resources/acct:bob@foobar.com

aliases:
  - "mailto:bob@foobar.com"
  - "https://mastodon/bob"
properties:
  'http://webfinger.example/ns/name': 'Bob Smith'
links:
  - rel: "http://webfinger.example/rel/profile-page"
    href: "https://www.example.com/~bob/"
  - rel: "http://webfinger.example/rel/businesscard"
    href: "https://www.example.com/~bob/bob.vcf"
```

The `aliases`, `properties`, and `links` fields describe the corresponding
fields of the resource as described in [Section 4.4 of the
RFC](https://datatracker.ietf.org/doc/html/rfc7033#section-4.4).

## Features
- [x] Serve WebFinger resources over HTTP
- [x] Serve resources from static YAML files
- [ ] Serve resources from an SQL database
- [ ] Serve resources from LDAP
