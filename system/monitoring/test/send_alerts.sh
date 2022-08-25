#!/usr/bin/env bash

name=fooAlert-$RANDOM
url='https://alertmanager.k8s.grigri/api/v1/alerts'
bold=$(tput bold)
normal=$(tput sgr0)

generate_post_data() {
  cat <<EOF
[{
  "status": "$1",
  "labels": {
    "alertname": "${name}",
    "service": "my-service",
    "severity":"warning",
    "instance": "${name}.example.net",
    "namespace": "fake-foo",
    "label_costcentre": "FOO"
  },
  "annotations": {
    "summary": "High latency is high!",
    "description": "Description example",
    "message": "Message example",
    "runbook_url": "https://example.com/$name"
  },
  "generatorURL": "https://example.com/$name"
  $2
  $3
},
{
  "status": "$1",
  "labels": {
    "alertname": "${name}2",
    "service": "my-service",
    "severity":"critical",
    "instance": "${name}2.example.com",
    "namespace": "fake-boo",
    "label_costcentre": "FOO"
  },
  "annotations": {
    "summary": "2High latency is high!",
    "description": "2Description example",
    "message": "2Message example",
    "runbook_url": "https://example.com/$name"
  },
  "generatorURL": "https://example.com/$name"
  $2
  $3
}]
EOF
}

echo "${bold}Firing alert ${name} ${normal}"
printf -v startsAt ',"startsAt" : "%s"' "$(date --rfc-3339=seconds | sed 's/ /T/')"
POSTDATA=$(generate_post_data 'firing' "${startsAt}")
curl $url --data "$POSTDATA" --trace-ascii /dev/stdout
echo -e "\n"

echo "${bold}Press enter to resolve alert ${name} ${normal}"
read -r

echo "${bold}Sending resolved ${normal}"
printf -v endsAt ',"endsAt" : "%s"' "$(date --rfc-3339=seconds | sed 's/ /T/')"
POSTDATA=$(generate_post_data 'firing' "${startsAt}" "${endsAt}")
curl $url --data "$POSTDATA" --trace-ascii /dev/stdout
echo -e "\n"
