package service

import (
	"github.com/Zilliqa/gozilliqa-sdk/keytools"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/transaction"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"strconv"
	"testing"
	"time"
)

func TestNonceManager_Run(t *testing.T) {
	zilClient := provider.NewProvider("https://dev-api.zilliqa.com/")
	privateKey := "e53d1c3edaffc7a7bab5418eb836cf75819a82872b4a1a0f1c7fcf5c3e020b89"
	addr := keytools.GetAddressFromPrivateKey(util.DecodeHex(privateKey))
	zilSender := &ZilSender{
		zilSdk:     zilClient,
		address:    addr,
		privateKey: privateKey,
	}
	balAndNonce, _ := zilClient.GetBalance(zilSender.address)
	nonceAndSender := &NonceAndSender{
		Sender:     zilSender,
		LocalNonce: balAndNonce.Nonce,
	}

	zilSenderMap := make(map[string]*NonceAndSender)
	zilSenderMap[privateKey] = nonceAndSender

	senderPrivateKeys := []string{privateKey}

	nm := NonceManager{
		UpdateInterval:        30,
		ZilClient:             zilClient,
		SentTransactions:      make(map[string]map[string]TransactionWithAge),
		ConfirmedTransactions: make(map[string][]string),
		ZilSenderMap:          zilSenderMap,
		SenderPrivateKeys:     senderPrivateKeys,
		CurrentIndex:          0,
	}

	go nm.Run()

	txn := &transaction.Transaction{
		Version:  strconv.FormatInt(int64(util.Pack(333, 1)), 10),
		ToAddr:   "4BAF5faDA8e5Db92C3d3242618c5B47133AE003C",
		Amount:   "10000000",
		GasPrice: "2000000000",
		GasLimit: "1",
		Code:     "",
		Data:     "",
		Priority: false,
	}

	for i := 0; i < 10; i++ {
		nm.send(txn)
	}

	time.Sleep(time.Second * 5)

	for i := 0; i < 20; i++ {
		nm.send(txn)
	}

	WaitToExit()
}
