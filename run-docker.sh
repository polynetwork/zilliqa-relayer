#!/bin/sh
docker stop zilliqa-relayer
docker rm zilliqa-relayer

docker run -d \
-v $(pwd)/persistence:/app/persistence \
-v $(pwd)/secrets/config.local.yaml:/app/config.local.yaml \
-v $(pwd)/secrets/target_contracts.json:/app/target_contracts.json \
-v $(pwd)/secrets/poly.wallet:/app/poly.wallet \
-v $(pwd)/secrets/zilliqa.wallet:/app/zilliqa.wallet \
--name zilliqa-relayer polynetwork/zilliqa-relayer 
