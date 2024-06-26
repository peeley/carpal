# Carpal

A small and configurable [WebFinger](https://webfinger.net/) server.

## Quick Start

A Docker image is provided, and can be run like so:

``` sh
docker run --rm -it \
    -v ./config/resources:/etc/carpal/resources \
    -p 8008:8008 \
    peeley/carpal:latest
```

...and in the local `./config/resources` directory, you can place your resource
files that look something like the following:

``` yaml
# config/resources/acct:bob@foobar.com

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

...and you can verify that everything is working with the following command:

``` sh
$ curl "localhost:8008/?resource=acct%3Abob%40foobar.com"

{
  "subject": "acct:bob@foobar.com",
  "aliases": [
    "mailto:bob@foobar.com",
    "https://mastodon/bob"
  ],
  "properties": {
    "http://webfinger.example/ns/name": "Bob Smith"
  },
  "links": [
    {
      "rel": "http://webfinger.example/rel/profile-page",
      "href": "https://www.example.com/~bob/"
    },
    {
      "rel": "http://webfinger.example/rel/businesscard",
      "href": "https://www.example.com/~bob/bob.vcf"
    }
  ]
}
```

## Configuration

The configuration file defaults to `/etc/carpal/config.yml`, and looks like this
by default:

``` yaml
# /etc/carpal/config.yml

driver: file
file:
  directory: /etc/carpal/resources/
```

You can change the location of the configuration file with the `CONFIG_FILE`
environment variable.

Carpal allows for the configuration of multiple different types of data sources.
By default, the `file` driver is used, but an `ldap` driver is also available
for fetching users from an LDAP directory.

### [File Driver](#file-driver)

The example file configures the file driver by default. The file driver simply
reads a YAML file representing a resource from a specified directory, converts
it to JSON, and returns it as an HTTP response to the client.

Resource files should be named after the resource they describe. For example,
the data for a resource named `acct:bob@foobar.com` should reside in
`/etc/carpal/resources/acct:bob@foobar.com` (or the corresponding `directory`
value in the config file). The resource file might look like the following:

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

For a complete example of the file driver, see the [example
configuration](configs/examples/file) provided.

### [LDAP Driver](#ldap-driver)

Carpal can also be configured to read resources from an LDAP directory.

Given the following configuration files:

``` yaml
# /etc/carpal/config.yml

driver: ldap
ldap:
  url: ldap://myldapserver
  bind_user: uid=myadmin,ou=people,dc=foobar,dc=com
  bind_pass: myadminpassword
  basedn: ou=people,dc=foobar,dc=com
  filter: (uid=*)
  user_attr: uid
  attributes:
    - uid
    - mail
    - cn
  template: /etc/carpal/ldap.gotempl
```

``` yaml
# /etc/carpal/ldap.gotempl

aliases:
  - "mailto:{{ index . "mail" }}"
  - "https://mastodon/{{ index . "uid" }}"
properties:
  'http://webfinger.example/ns/name': '{{ index . "cn" }}'
links:
  - rel: "http://webfinger.example/rel/profile-page"
    href: "https://www.example.com/~{{ index . "uid" }}/"
```

Carpal, when sent requests for a resource like `acct:bob@foobar.com`, will look
for any LDAP resource within `ou=people,dc=foobar,dc=com` with the `uid` of
`bob`. When a resource is found, any fields it contains matching the list of
`attributes` given is then substituted in the specified `.gotempl` file,
converted to JSON, and returned in the HTTP response to the client.

For the moment, only `acct:` WebFinger resources are supported; additional
resource types _may_ be supported in the future. Also note that the
`@foobar.com` of the resource name from the request is discarded when searching
for a resource in LDAP. As such, no verification is done to ensure the LDAP
resource resides in a domain name matching the WebFinger resource's.

For a complete example of the LDAP driver, see the [example
configuration](configs/examples/ldap) provided.

## Features
- [x] Serve WebFinger resources over HTTP
- [x] Serve resources from static YAML files
- [x] Serve resources from LDAP
- [ ] Serve resources from an SQL database
