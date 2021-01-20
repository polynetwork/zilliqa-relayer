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
docker build -t polynetwork/zilliqa-relayer -f Dockerfile ./
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
  zil_api: https://polynetworkcc3dcb2-5-api.dev.z7a.xyz
  zil_chain_id: 333
  zil_message_version: 1
  zil_start_height: 38
  zil_monitor_interval: 40
  corss_chain_manager_address: zil16vxy2u59sct5nupryxm3wfgteuhve9p0hp605f
  side_chain_id: 1
  key_store_path: ./keystore
  key_store_pwd_set:
    7d48043742a1103042d327111746531ca26be9be: "pwd1"
    de0a0fbe9042aa165fb21e7b5e648162bcf1e8e7: "pwd2"
poly_config:
  poly_wallet_file: wallet.dat
  poly_wallet_pwd: dummy
  poly_monitor_interval: 40
  entrance_contract_address: "0300000000000000000000000000000000000000"
  rest_url: http://beta1.poly.network
```

A sample keystore file could be:

```text
{"address":"7d48043742a1103042d327111746531ca26be9be","id":"6cd445ed-8f5f-4565-af2a-cc2306a82b73","version":3,"crypto":{"cipher":"aes-128-ctr","ciphertext":"d136660a4e5664709031ebc162616556e8c812ab37d0157ea3276aa08d0a6c2d","kdf":"pbkdf2","mac":"b30dd459f1fd9d99c0b2f3452ccd2bf11414ad92d32ac70d1d7b52f17281b4e5","cipherparams":{"iv":"6a14f95c8cbafe7d1f317bec88e9d1b8"},"kdfparams":{"n":8192,"c":262144,"r":8,"p":1,"dklen":32,"salt":"c4939e7cead32935d1972a2cd06d249dd501181e6ad2d1872fa0eb397d7fea20"}}}
{"address":"de0a0fbe9042aa165fb21e7b5e648162bcf1e8e7","id":"e6ed9aba-7be3-40ca-8fab-7f9438fdf7cf","version":3,"crypto":{"cipher":"aes-128-ctr","ciphertext":"b96b59ea5861c7048f62cf15c219faa9e1494495030f42f021b6277622ab819f","kdf":"pbkdf2","mac":"ad06d0f0f5df29ad0947a62954ed08b084c2fc11aec66a36ab2c79eb1398768c","cipherparams":{"iv":"ef08dc8dbca886b1141e13fb7e988817"},"kdfparams":{"n":8192,"c":262144,"r":8,"p":1,"dklen":32,"salt":"26bc16e8bc0dc2749c1cde2e62aa5c9990898c5c4d3f78c822fed2988c0ab682"}}}
```

## Other Resources

- [zilliqa cross chain manager proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManagerProxy.scilla)
- [zilliqa cross chain manager](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManager.scilla)
- [zilliqa lock proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/LockProxy.scilla)
- [polynetwork](https://github.com/polynetwork/poly)



