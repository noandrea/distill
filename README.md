# Distill

Another url shortener

[![Build Status](https://travis-ci.com/noandrea/distill.svg?branch=master)](https://travis-ci.com/noandrea/distill) [![codecov](https://codecov.io/gh/noandrea/distill/branch/master/graph/badge.svg)](https://codecov.io/gh/noandrea/distill) [![GoDoc](https://godoc.org/github.com/noandrea/distill?status.svg)](https://godoc.org/github.com/noandrea/distill) [![Go Report Card](https://goreportcard.com/badge/github.com/noandrea/distill)](https://goreportcard.com/report/github.com/noandrea/distill)

[![Docker](https://img.shields.io/badge/docker-noandrea/distill-blue)]

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
- Set a expiration date on short ids (globally or per id)
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
- Expiration date
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

Response:

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

- to enable coverage badge use `^coverage:\s(\d+(?:\.\d+)?%)` as regexp in gitlab configuration

## Hints

To generate an API Token randomly use the `make gen-secret` (linux/mac only):

```
make gen-secret 
WiYS8DauSwVIMeNGIp63ScmY-pgA1ECA7ai7Oce7
```

## Installation

Distill is distributed via different channels

### Docker

Distill is available on [docker hub](https://hub.docker.com/r/noandrea/distill), 
as for the images:

- `latest` tag is built from the [`develop`](https://github.com/noandrea/distill/tree/develop) git branch and contains the latest changes
- tags in the form `x.y.z` are built from [git tags](https://github.com/noandrea/distill/releases) and are considered stable

to run distill in docker use the following command:

```
docker run -p 1804:1804 noandrea/distill
```

the default configuration for the docker image is available

- [here](https://github.com/noandrea/distill/blob/master/configs/settings.docker.yaml) for stable releases
- [here](https://github.com/noandrea/distill/blob/develop/configs/settings.docker.yaml) for `latest` releases

#### Mount points

To override the configuration mount a volume in `/settings.docker.yaml` path.
For the data folder the mount point is `/data`.

#### Docker Compose

A [`docker-compose`](https://docs.docker.com/compose/) example is available in the [`examples/docker`](https://github.com/noandrea/distill/blob/master/examples/docker)

### Systemd

Distill can be run via `systemd`, check the [example](https://github.com/noandrea/distill/blob/master/examples/systemd) configuration.


# Structure

the datastore package exposes an interface that has the following

datastore
  Put(key string, data protoreflect.ProtoMessage) err
  Get(key string, data *protoreflect.ProtoMessage) (found bool, err error)
  
  ConunterPlus(key) err
  ConunterMinus(key) err

  Hit(key string) (URLInfo, error)
  Peek(key string) (URLInfo, error)
  Insert(key string, u URLInfo) err
  Upsert(key string, u URLInfo) err
  Delete(key string) err


distill
  deal with settings
  deal with stats
  deal with urls 

