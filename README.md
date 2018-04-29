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
- set a time to live for a short id
- set a limit of how many time a short url can be view
- set a redirect for the / path

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