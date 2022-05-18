#!/bin/bash

# Load envs
if [ -f .env.local ]; then
    export $(echo $(cat .env.local | sed 's/#.*//g' | xargs) | envsubst)
fi

BASE_URL="https://raw.githubusercontent.com/haverzard/container-migration/main/internal/jobs"
url="${SCRIPT_URL:-$BASE_URL}"

# Substitution
data=$(cat $1 && echo .)
data=${data%.}
data="${data/NODE_1/$NODE_1_NAME}"
data="${data/NODE_2/$NODE_2_NAME}"
data="${data/NODE_3/$NODE_3_NAME}"
data="${data//$BASE_URL/$url}"

printf %s "$data"
