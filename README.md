# distill

Another url shortener 

## Motivations

the available url shortener are not suitable for a private deploy use or are too complex in terms of requirements.

*Distill* aims to create a easy deployable short url service
that can be used for specific events.


## Features 

- chose the alphabet set for the generate short id
- chose the length of the generate short id
- load existing short id <-> url mappings*
- update an existing short id with a different target url*
- set a time to live on short ids (globally or per id) 
- set a end date on short ids (globabbly or per id)
- set a request limit on short ids (globally or per id)
- set a redirect for the `/` path
- set a redirect url for the expired ids
- backup/restore urls in csv or binary format
- import data via csv

\* the alphabet and lenght can be enforced

## Expiration strategy

There are 3 ways to set an expiration for a short id:
 - ttl (seconds)
 - epiration date
 - max requests

the three options can be configured globally or per short id, 
the value specified for the short id takes always precedence over the 
global configuration

For the *ttl* and the *expiration date* the actual expiration is selected as 
` max ( creation_date + ttl, expiration_date) `

> !!! the expiration is set upon short id creation, changing global configuration 
> will not affect the short ids already set !!!

## Backup / Restore

Offline backup in csv and binary format

## Import data

minimum fields 

```url```

all fields 

```url,id,max_requests,ttl,expires_on```

the dates are expressed in RFC3339 format

example: 

```
url,id,max_requests,ttl,expires_on
https://hackernews.com,2018-05-06T22:31:41Z,500,86400,2019-05-06T22:05:18Z
https://hackernews.com,2018-05-06T22:31:43Z,500,86400,2019-05-06T22:05:18Z
https://hackernews.com,2018-05-06T22:31:56Z,,,,
https://hackernews.com,2018-05-06T22:31:56Z,,,,
```



## Configuration 





## Build targets

default 

build (build-dist)
clean 

docker (docker-build) 
docker-push 
docker-run 
lint 
test

*D#* distill is available as docker image

*D*docker run welance/distill`

## Development
to generate the Colfer model run 
`colf -b internal Go api/model.colf` from the project root