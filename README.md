# zilliqa-relayer

*This program is still under developing!*

Zilliqa Relayer is an important character of Poly cross-chain interactive protocol which is responsible for relaying cross-chain transaction from and to Zilliqa.

## Build From Source

### Prerequisites

- [Golang](https://golang.org/doc/install) version 1.14 or later

### Build

```shell
git clone https://github.com/polynetwork/zil-relayer.git
cd zil-relayer
./build.sh
```

After building the source code successfully,  you should see the executable program `zilliqa-relayer`.

### Build Docker Image

```
docker build -t polynetwork/zilliqa-relayer .
```

This command will copy config.yaml to /app/config.yaml in the image. So you need to prepare config.yaml before running this command and you should start the zilliqa-relayer in container basing on the configuration in /app/config.yaml.


## Run Relayer
Before you can run the relayer you will need to create a wallet file of PolyNetwork by running(build Poly CLI first):

```shell
./poly account add -d
```

After creation, you need to register it as a Relayer to Poly net and get consensus nodes approving your registeration. And then you can send transaction to Poly net and start relaying.

Before running, you need feed the configuration file `config.yaml`.

```yaml
zil_config:
  zil_api: https://api.zilliqa.com
  zil_chain_id: 111
  zil_message_version: 1
  zil_force_height: 0
  zil_monitor_interval: 10
  zil_headers_per_batch: 2
  corss_chain_manager_address: zil1tjru7m5zdn3x6k0t72nzmmpz62e5qds62nte9t
  cross_chain_manager_proxy_address: zil1n7wkwr0xxslwsrhnqtjrwlus80dp5ncnlpaw93
  side_chain_id: 85
  key_store_path: zilliqa.wallet
  key_store_pwd_set:
    6c89b62d65dc632e259b96f7ae2f1d68a27e3383: ""
poly_config:
  poly_wallet_file: poly.wallet
  poly_wallet_pwd:
  poly_start_height: 0
  poly_monitor_interval: 2
  entrance_contract_address: "0300000000000000000000000000000000000000"
  rest_url: http://poly.com
target_contracts: target_contracts.json
db_path: persistence
```

A sample keystore file could be:

```text
{"address":"7d48043742a1103042d327111746531ca26be9be","id":"6cd445ed-8f5f-4565-af2a-cc2306a82b73","version":3,"crypto":{"cipher":"aes-128-ctr","ciphertext":"d136660a4e5664709031ebc162616556e8c812ab37d0157ea3276aa08d0a6c2d","kdf":"pbkdf2","mac":"b30dd459f1fd9d99c0b2f3452ccd2bf11414ad92d32ac70d1d7b52f17281b4e5","cipherparams":{"iv":"6a14f95c8cbafe7d1f317bec88e9d1b8"},"kdfparams":{"n":8192,"c":262144,"r":8,"p":1,"dklen":32,"salt":"c4939e7cead32935d1972a2cd06d249dd501181e6ad2d1872fa0eb397d7fea20"}}}
```
# Relayer Container Administration
## Running the Relayer Container 
### Prerequisites:
If you are running via docker-compose, you'll need to install docker-compose first via the following guide:
```
https://docs.docker.com/compose/install/
```

You need the following files and folders created: <br />
If persistence is already created:
```
./persistence/bolt.bin
```
If persistence folder is not created:
```
create a folder named 'persistence'.
```
Configuration Files:
```
secrets/config.local.yaml
secrets/target_contracts.json
secrets/poly.wallet
secrets/zilliqa.wallet
```
### Running Relayer via Docker Container
```
./docker-run
```

### Running Relayer via Docker Compose
```
docker-compose up -d
```
## Stopping the Relayer
### Stopping Relayer via Docker Container
```
docker stop zilliqa-relayer && docker rm zilliqa-relayer
```
### Stopping Relayer via Docker Compose
```
docker-compose down
```
## Getting logs
```
docker logs -f zilliqa-relayer
```


## Other Resources

- [zilliqa cross chain manager proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManagerProxy.scilla)
- [zilliqa cross chain manager](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManager.scilla)
- [zilliqa lock proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/LockProxy.scilla)
- [polynetwork](https://github.com/polynetwork/poly)




