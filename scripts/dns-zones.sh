#!/bin/sh

# This scripts tries to find all DNS zones with the given domain in all projects you have access to.
# It will output a list of projects and zones that matches the domain.
#
# Usage:
#  DOMAIN=intern.nav.no ./dns-zones.sh
#
# The script will create a list of projects with the DNS API enabled to speed up future runs.


domain=${DOMAIN:-"intern.nav.no"}
## Fetch all projects. This is a slow operation, so we cache the result in a file called projects_with_dns.txt
## If the file exists, we use that instead of fetching all projects again.
projects=${PROJECTS:-$(cat projects_with_dns.txt 2>/dev/null || gcloud projects list --format 'get(projectId)')}

## No edit below this line

hasProjectTxt=$(test ! -f projects_with_dns.txt)
for PROJECT in $projects; do
  if [ -z "$PROJECT" ]; then
    continue
  fi

  if ! zones=$(gcloud dns managed-zones list --format "get(dnsName)" --project "$PROJECT" --quiet 2>/dev/null); then
    continue
  fi

  if [ "$hasProjectTxt" ]; then
    echo "$PROJECT" >> projects_with_dns.txt
  fi

  for zone in $zones; do
    # If zone matches intern.nav.no using grep, output project and zone name
    if echo "$zone" | grep -q "$domain"; then
      echo "$PROJECT $zone"
    fi
  done

done
