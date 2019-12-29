#!/bin/bash

DISTILL_BASE_URL=https://distill.plus

# colors
red='\e[0;31m'
green='\e[0;32m'
yellow='\e[0;33m'
reset='\e[0m'

echoRed() { echo -e "${red}$1${reset}"; }
echoGreen() { echo -e "${green}$1${reset}"; }
echoYellow() { echo -e "${yellow}$1${reset}"; }

check_api_key_set() {
  if [ -z ${DISTILL_API_KEY+x} ]; then echo "The DISTILL_API_KEY env var is not set!"; exit 1; fi
}

demo() {
  check_api_key_set
  curl $DISTILL_BASE_URL/demo -w "Response time:%{time_starttransfer}\n"
  cat data.json 
  curl -X POST $DISTILL_BASE_URL/api/short -d "@data.json" -H "X-API-KEY: $DISTILL_API_KEY" -H "Content-type: application/json"  -w  "Response time:%{time_starttransfer}\n"
  curl $DISTILL_BASE_URL/demo -w "Response time:%{time_starttransfer}\n"
  sleep 5
  curl $DISTILL_BASE_URL/demo -w "Response time:%{time_starttransfer}\n"
}

site() {
  echo "Setting up url for the site"
  check_api_key_set
  for file in $(ls | grep site-data); 
    do 
      echo "Processing $file whit id $(cat $file | jq -r '.id')";
      curl -X POST $DISTILL_BASE_URL/api/short -d "@$file" -H "X-API-KEY: $DISTILL_API_KEY" -H "Content-type: application/json"  -w  "Response time:%{time_starttransfer}\n";
      curl $DISTILL_BASE_URL/$(cat $file | jq -r '.id') -w "Response time:%{time_starttransfer}\n"
  done
}

case "$1" in
demo)
   demo 
;;

site)
   site
   exit 0
;;

*)
    echo "Usage: $0 {demo|site}"
    exit 1
esac
