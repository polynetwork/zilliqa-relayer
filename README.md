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
  zil_start_height: 38
  zil_monitor_interval: 40
  corss_chain_manager_address: zil16vxy2u59sct5nupryxm3wfgteuhve9p0hp605f
  side_chain_id: 1
poly_config:
  poly_wallet_file: wallet.dat
  poly_wallet_pwd: dummy
  entrance_contract_address: 0300000000000000000000000000000000000000
  rest_url: http://beta1.poly.network	
```

## Other Resources

- [zilliqa cross chain manager proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManagerProxy.scilla)
- [zilliqa cross chain manager](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManager.scilla)
- [zilliqa lock proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/LockProxy.scilla)
- [polynetwork](https://github.com/polynetwork/poly)



