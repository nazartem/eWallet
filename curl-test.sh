#!/usr/bin/bash

set -eux
set -o pipefail

SERVERPORT=4112
SERVERADDR=localhost:${SERVERPORT}

# Send money
curl -iL -w "\n" -X POST -H "Content-Type: application/json" --data '{"from":"--ADDRESS--","to":"--ADDRESS--", "amount":0}' ${SERVERADDR}/api/send

# Get balance by address
curl -iL -w "\n" ${SERVERADDR}/api/wallet/--ADDRESS--/balance

# Get last transaction
curl -iL -w "\n" -X GET ${SERVERADDR}/api/transactions?count=4