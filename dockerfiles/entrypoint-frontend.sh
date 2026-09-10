#!/bin/sh
set -eu
envsubst '${OIDC_URL} ${OIDC_CLIENT_ID} ${OIDC_CLIENT_SECRET}' \
  < /usr/share/nginx/html/env.js.tmpl \
  > /usr/share/nginx/html/env.js

