# zilliqa-relayer

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


## Other Resources

- [zilliqa cross chain manager proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManagerProxy.scilla)
- [zilliqa cross chain manager](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/ZilCrossChainManager.scilla)
- [zilliqa lock proxy](https://github.com/Zilliqa/zilliqa-contracts/blob/main/contracts/LockProxy.scilla)



