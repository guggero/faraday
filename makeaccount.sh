#!/bin/bash

ACCOUNT_PERMISSIONS="info:read offchain:read offchain:write onchain:read invoices:read invoices:write"

frcli --network regtest createaccount --balance $1

id=$(frcli --network regtest listaccounts | jq -r '.accounts[0].id')

lncli --lnddir $HOME/.lnd-dev-zane --network regtest --rpcserver localhost:10019 bakemacaroon --save_to /tmp/accounts.macaroon --custom_caveat_name account --custom_caveat_condition $id $ACCOUNT_PERMISSIONS

lncli printmacaroon --macaroon_file /tmp/accounts.macaroon

lndconnect --host localhost --port 10019 --tlscertpath $HOME/.lnd-dev-zane/tls.cert --adminmacaroonpath /tmp/accounts.macaroon -j | grep -v lnd.conf

