/*
 * Copyright (C) 2021 Zilliqa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package service

import (
	"encoding/json"
	"github.com/Zilliqa/gozilliqa-sdk/keytools"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/transaction"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	polytypes "github.com/polynetwork/poly/core/types"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/tools"
	log "github.com/sirupsen/logrus"
	"math"
	"strconv"
	"sync"
	"time"
)

const AuditLogFile = "audit.log"

type NonceAndSender struct {
	Sender     *ZilSender
	LocalNonce int64
}

type TransactionWithAge struct {
	Txn          *transaction.Transaction
	StartTxBlock uint64
	Age          int
}

type NonceManager struct {
	UpdateInterval int64
	ZilClient      *provider.Provider
	// address => hash => transaction
	SentTransactions    map[string]map[string]TransactionWithAge
	LockSentTransaction sync.Mutex
	// address => list of transaction hash
	ConfirmedTransactions map[string][]string
	// private key => nonce and sender
	ZilSenderMap      map[string]*NonceAndSender
	SenderPrivateKeys []string
	CurrentIndex      int
	Cfg               *config.Config
}

func (nm *NonceManager) Run() {
	log.Infof("starting nonce manager...")
	for {
		time.Sleep(time.Second * time.Duration(nm.UpdateInterval))
		nm.stat()
	}
}

func (nm *NonceManager) commitHeader(hdr *polytypes.Header) bool {
	nm.LockSentTransaction.Lock()
	defer nm.LockSentTransaction.Unlock()
	currentSenderPrivateKey := nm.SenderPrivateKeys[nm.CurrentIndex]
	nm.CurrentIndex++
	if nm.CurrentIndex > len(nm.SenderPrivateKeys)-1 {
		nm.CurrentIndex = 0
	}
	currentSender := nm.ZilSenderMap[currentSenderPrivateKey].Sender
	log.Infof("NonceManager - commitHeader use sender %s", currentSender.address)
	nonce := strconv.FormatUint(uint64(nm.ZilSenderMap[currentSenderPrivateKey].LocalNonce+1), 10)
	txn, err := currentSender.commitHeaderWithNonce(hdr, nonce)
	if err != nil {
		log.Warnf("NonceManager - commitHeader error %s", err.Error())
		return false
	}

	hash, _ := txn.Hash()

	// handle nonce
	nm.ZilSenderMap[currentSenderPrivateKey] = &NonceAndSender{
		Sender:     nm.ZilSenderMap[currentSenderPrivateKey].Sender,
		LocalNonce: nm.ZilSenderMap[currentSenderPrivateKey].LocalNonce + 1,
	}

	outerMap := nm.SentTransactions[currentSender.address]
	if outerMap == nil {
		outerMap = make(map[string]TransactionWithAge)
	}
	outerMap[util.EncodeHex(hash)] = TransactionWithAge{
		Txn: txn,
		Age: 0,
	}
	nm.SentTransactions[currentSender.address] = outerMap
	return true
}

type TransactionAuditLog struct {
	Time    time.Time
	PayLoad string
	Error   string
}

func (nm *NonceManager) commitDepositEventsWithHeader(header *polytypes.Header, param *common2.ToMerkleValue, headerProof string, anchorHeader *polytypes.Header, polyTxHash string, rawAuditPath []byte) bool {
	nm.LockSentTransaction.Lock()
	defer nm.LockSentTransaction.Unlock()
	currentSenderPrivateKey := nm.SenderPrivateKeys[nm.CurrentIndex]
	nm.CurrentIndex++
	if nm.CurrentIndex > len(nm.SenderPrivateKeys)-1 {
		nm.CurrentIndex = 0
	}

	currentSender := nm.ZilSenderMap[currentSenderPrivateKey].Sender
	log.Infof("NonceManager - commitDepositEventsWithHeader use sender %s", currentSender.address)
	nonce := strconv.FormatUint(uint64(nm.ZilSenderMap[currentSenderPrivateKey].LocalNonce+1), 10)
	txn, err := currentSender.commitDepositEventsWithHeaderWithNonce(header, param, headerProof, anchorHeader, polyTxHash, rawAuditPath, nonce)
	var auditLog TransactionAuditLog
	if err != nil {
		auditLog = TransactionAuditLog{
			Time:  time.Now(),
			Error: err.Error(),
		}
	} else {
		transactionRaw, _ := json.Marshal(txn)
		auditLog = TransactionAuditLog{
			Time:    time.Now(),
			PayLoad: string(transactionRaw),
		}
	}

	auditLogRaw, _ := json.Marshal(auditLog)
	tools.AppendToFile(AuditLogFile, string(auditLogRaw))

	if err != nil {
		log.Warnf("NonceManager - commitDepositEventsWithHeaderWithNonc e error %s", err.Error())
		return false
	}

	hash, _ := txn.Hash()

	// handle nonce
	nm.ZilSenderMap[currentSenderPrivateKey] = &NonceAndSender{
		Sender:     nm.ZilSenderMap[currentSenderPrivateKey].Sender,
		LocalNonce: nm.ZilSenderMap[currentSenderPrivateKey].LocalNonce + 1,
	}

	outerMap := nm.SentTransactions[currentSender.address]
	if outerMap == nil {
		outerMap = make(map[string]TransactionWithAge)
	}
	outerMap[util.EncodeHex(hash)] = TransactionWithAge{
		Txn: txn,
		Age: 0,
	}
	nm.SentTransactions[currentSender.address] = outerMap
	return true
}

func (nm *NonceManager) stat() {
	txBlock, err := nm.ZilClient.GetLatestTxBlock()
	if err != nil {
		log.Warnf("NonceManager - get current tx block number error: %s", err.Error())
		return
	}
	currentTxEpoch, _ := strconv.ParseUint(txBlock.Header.BlockNum, 10, 64)
	log.Infof("NonceManager - current tx block number is: %s", txBlock.Header.BlockNum)
	nm.LockSentTransaction.Lock()
	defer nm.LockSentTransaction.Unlock()
	for _, key := range nm.SenderPrivateKeys {
		addr := keytools.GetAddressFromPrivateKey(util.DecodeHex(key))
		balAndNonce, err := nm.ZilClient.GetBalance(addr)
		if err != nil {
			log.Warnf("NonceManager - get nonce for address %s error %s", addr, err.Error())
			continue
		}

		// print some stat info about this address
		log.Infof("NonceManager - address %s, local nonce = %d, remote nonce = %d", addr, nm.ZilSenderMap[key].LocalNonce, balAndNonce.Nonce)
		log.Infof("NonceManager - sent transactions: %+v", nm.SentTransactions[addr])
		log.Infof("NonceManager - confimred transactions: %+v", len(nm.ConfirmedTransactions[addr]))

		// check sent transactions
		log.Infof("NonceManager - check sent transactions")
		var confirmedTxn []string

		sentTransactionMap := nm.SentTransactions[addr]
		for hash, txn := range sentTransactionMap {
			log.Infof("NonceManager - check transaction %s", hash)
			_, err := nm.ZilClient.GetTransaction(hash)

			if err == nil {
				log.Infof("NonceManager - transaction %s confirmed", hash)
				confirmedTxn = append(confirmedTxn, hash)
			} else {
				// if start block is 0, try to give it a number first
				if sentTransactionMap[hash].StartTxBlock == 0 {
					log.Infof("NonceManager - stat try to determine start tx block for hash: %s", hash)
					sentTransactionMap[hash] = TransactionWithAge{
						Txn:          txn.Txn,
						StartTxBlock: currentTxEpoch,
						Age:          0,
					}
				} else {
					log.Warnf("NonceManager - stat already has inserted epoch, update age, hash is %s", hash)
					age := 0
					if currentTxEpoch > sentTransactionMap[hash].StartTxBlock {
						age = int(currentTxEpoch - sentTransactionMap[hash].StartTxBlock)
					}
					sentTransactionMap[hash] = TransactionWithAge{
						Txn:          txn.Txn,
						StartTxBlock: txn.StartTxBlock,
						Age:          age,
					}
				}
			}

		}

		for _, hash := range confirmedTxn {
			delete(sentTransactionMap, hash)
			nm.ConfirmedTransactions[addr] = append(nm.ConfirmedTransactions[addr], hash)
		}
		nm.SentTransactions[addr] = sentTransactionMap

		// print some stat info about this address again
		log.Infof("NonceManager - sent transactions: %+v", nm.SentTransactions[addr])
		log.Infof("NonceManager - confimred transactions: %+v", len(nm.ConfirmedTransactions[addr]))

		// detect dead transactions
		log.Infof("NonceManager - stat start to detect dead transactions")
		currentNonce := uint64(math.MaxUint64)
		for hash, txn := range nm.SentTransactions[addr] {
			if txn.Age > nm.Cfg.ZilConfig.MaxExistTxEpoch {
				log.Warnf("NonceManager - stat found dead transaction, hash: %s, nonce is %s", hash, txn.Txn.Nonce)
				log.Warnf("NonceManager - stat current nonce is: %d", currentNonce)
				nonce, _ := strconv.ParseUint(txn.Txn.Nonce, 10, 64)
				if currentNonce > nonce {
					log.Warnf("NonceManager - stat replace current nonce with it")
					currentNonce = nonce
				}
			}
		}

		if currentNonce == math.MaxUint64 {
			log.Infof("NonceManager - stat no dead found")
		} else {
			log.Infof("NonceManager - stat dead transaction, bad nonce is: %d, start to resend transactions", currentNonce)
			for hash, txn := range nm.SentTransactions[addr] {
				nonce, _ := strconv.ParseUint(txn.Txn.Nonce, 10, 64)
				if nonce >= currentNonce {
					log.Infof("NonceManager - stat start to resend transaction %s, nonce %d", hash, nonce)
					// todo handle error
					nm.ZilClient.CreateTransaction(txn.Txn.ToTransactionPayload())
					nm.SentTransactions[addr][hash] = TransactionWithAge{
						Txn:          txn.Txn,
						StartTxBlock: currentTxEpoch,
						Age:          0,
					}
				}
			}

		}
	}
}
