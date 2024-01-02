rm -rf ./deployments/baobab_test
rm -rf ./migration/baobab_test

node ./scripts/v0.1/admin-aggregator/generate-aggregator-deployments.cjs --pairs [\"BTC-USDT\"] --chain baobab_test

curl get "https://config.orakl.network/aggregator/default/btc-usdt.aggregator.json" > /app/contracts/scripts/v0.1/tmp/btc-usdt.aggregator.json

jq --argjson input $(yarn hardhat deploy --network baobab_test --deploy-scripts deploy/Aggregator | tail -n 2 | head -n 1 ) '.address = $input["Aggregator_BTC-USDT"]' /app/contracts/scripts/v0.1/tmp/btc-usdt.aggregator.json > /app/contracts/scripts/v0.1/tmp/updated-btc-usdt.aggregator.json
jq '.bulk[0].aggregatorSource = "/app/tmp/updated-btc-usdt.aggregator.json"' /app/contracts/scripts/v0.1/tmp/bulk.json > /app/contracts/scripts/v0.1/tmp/updated_bulk.json