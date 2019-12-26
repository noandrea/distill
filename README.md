# Distill

Another url shortener

[![Build Status](https://travis-ci.com/noandrea/distill.svg?branch=master)](https://travis-ci.com/noandrea/distill) [![codecov](https://codecov.io/gh/noandrea/distill/branch/master/graph/badge.svg)](https://codecov.io/gh/noandrea/distill) [![GoDoc](https://godoc.org/github.com/noandrea/distill?status.svg)](https://godoc.org/github.com/noandrea/distill) [![Go Report Card](https://goreportcard.com/badge/github.com/noandrea/distill)](https://goreportcard.com/report/github.com/noandrea/distill)

## Motivations

Existing url shorteners are not suitable for a private deploy use or are too complex in terms of requirements.

_Distill_ aims to create a easy deployable short url service
that can be used for specific events.

## Features

- Choose the alphabet set for the generate short id
- Choose the length of the generate short id
- Load existing short id <-> url mappings\*
- Overwrite an existing short id with a different target url\*
- Set a time to live on short ids (globally or per id)
- Set a expiration date on short ids (globabbly or per id)
- Set a request limit on short ids (globally or per id)
- Set a redirect for the `/` path
- Set a redirect url for exhausted ids (request limit reached)
- Set a redirect url for expired ids (ttl/end date reached)
- Backup/restore urls in csv or binary format
- Import data via csv
- Get statistics both globally and for short id

\* the alphabet and lenght can be enforced

## Expiration strategy

There are 3 ways to set an expiration for a short id:

- TTL (seconds)
- Epiration date
- Max requests

The three options can be configured globally or per short id,
the value specified for the short id takes always precedence over the
global configuration.

For the _TTL_ and the _expiration date_ the actual expiration is selected as
`max ( creation_date + ttl, expiration_date)`

> !!! the expiration is set upon short id creation, changing global configuration
> will not affect the short ids already set !!!

For redirects, the expiration url redirect takes precedence over the exhaustion url redirect.

If no redirects are set for exhausted / expired url then a `404` is returned.

## Api Doc

### Register a redirect

Miniaml request is:

```
POST http://localhost:1804/api/short
X-API-KEY: 123123_changeme_changeme
Content-Type: application/json

{
    "url": "https://example.com/target_url"
}
```

Repsonse:

```
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Origin: *
Content-Type: application/json
Date: Sun, 17 Mar 2019 21:18:16 GMT
Content-Length: 16
Connection: close

{
  "id": "wBNaqx"
}
```

A request can contain additional fields:

```
POST http://localhost:1804/api/short
X-API-KEY: 123123_changeme_changeme
Content-Type: application/json

{
    "id": "myid"
    "url": "https://example.com/target_url",
    "max_requests": 20,
    "url_exhausted" : "https://example.com/max_requests_reached_url",
    "ttl": 0,
    "expire_on": "2039-03-17T22:05:28+01:00",
    "url_expired" : "https://example.com/ttl_or_epiration_reached_url"
}
```

Repsonse:

```
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: *
Access-Control-Allow-Methods: *
Access-Control-Allow-Origin: *
Content-Type: application/json
Date: Sun, 17 Mar 2019 21:18:16 GMT
Content-Length: 16
Connection: close

{
  "id": "myid"
}
```

## Backup / Restore

Offline backup in csv and binary format

## Import data

required fields

```
url
```

all fields

```
url,id,max_requests,url_exhausted,ttl,expires_on,url_expired
```

the dates are expressed in RFC3339 format

## Configuration

TODO

## Example API request

TODO

## Build targets

default

build (build-dist)
clean

docker (docker-build)
docker-push
docker-run
lint
test

## Development

- to generate the Colfer model run
  `colf -b internal Go api/model.colf` from the project root

- to enable coverage badge use `^coverage:\s(\d+(?:\.\d+)?%)` as regexp in gilab configuration

## Hints

To generate an API Token randomly use the `make gen-secret` (linux/mac only):

```
make gen-secret 
WiYS8DauSwVIMeNGIp63ScmY-pgA1ECA7ai7Oce7
```
