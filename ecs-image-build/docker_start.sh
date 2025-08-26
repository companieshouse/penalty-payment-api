#!/bin/bash
#
# Start script for penalty-payment-api
PORT="8080"
if [[ ! -x ./penalty-payment-api ]]; then
  echo "ERROR: ./penalty-payment-api not found or not executable"
  exit 1
fi

# Read brokers and topics from environment and split on comma
IFS=',' read -ra BROKERS <<< "${KAFKA_BROKER_ADDR}"

# Ensure we only populate the broker address and topic via application arguments
unset KAFKA_BROKER_ADDR

# Read kafka 3 brokers from environment and split on comma
IFS=',' read -ra KAFKA3_BROKERS <<< "${KAFKA3_BROKER_ADDR}"

# Ensure we only populate the kafka 3 broker address via application arguments
unset KAFKA3_BROKER_ADDR

exec ./penalty-payment-api "-bind-addr=:${PORT}" $(for broker in "${BROKERS[@]}"; do echo -n "-broker-addr=${broker} "; done) $(for broker in "${KAFKA3_BROKERS[@]}"; do echo -n "-kafka3-broker-addr=${broker} "; done)
