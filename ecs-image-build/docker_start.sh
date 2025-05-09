#!/bin/bash
#
# Start script for file-delivery-api
PORT="8080"
if [[ ! -x ./penalty-payment-api ]]; then
  echo "ERROR: ./penalty-payment-api not found or not executable"
  exit 1
fi
exec ./penalty-payment-api "-bind-addr=:${PORT}"