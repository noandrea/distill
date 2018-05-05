# ilij.li

Another url shortener 

## Motivations

the available url shortener are not suitable for a private deploy use or are too complex in terms of requirements.

ilij aims to create a easy deployable short url service
that can be used for specific events.


## Features 

- chose the alphabet set for the short id
- chose the length of the short id
- update an existing short id with a different target url
- set a time to live on short ids (globally or per id) 
- set a end date on short ids (globabbly or per id)
- set a request limit on short ids (globally or per id)
- set a redirect for the `/` path
- set a redirect url for the expired ids
- backup urls as csv
- load urls via csv

## Expiration strategy

There are 3 ways to set an expiration for a short id:
 - ttl
 - epiration date
 - max requests

the three options can be configured globally or per short id, 
the value specified for the short id takes always precedence over the 
global configuration

For the *ttl* and the *expiration date* the actual expiration is selected as 
` max ( creation_date + ttl, expiration date) `

> !!! the expiration is set upon short id creation, changing global configuration 
> will not affect the short ids already set !!!


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

## ilij is available as docker image

`docker run welance/ilij`

## Development
to generate the Colfer model run 
`colf -b internal Go api/model.colf` from the project root