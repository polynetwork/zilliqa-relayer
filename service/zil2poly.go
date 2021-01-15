package service

import (
	"encoding/hex"
	"github.com/Zilliqa/gozilliqa-sdk/bech32"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/polynetwork/poly/common"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
	"time"
)

type CrossTransfer struct {
	txIndex string
	txId    []byte
	value   []byte
	toChain uint32
	height  uint64
}

func (this *CrossTransfer) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.txIndex)
	sink.WriteVarBytes(this.txId)
	sink.WriteVarBytes(this.value)
	sink.WriteUint32(this.toChain)
	sink.WriteUint64(this.height)
}

func (s *SyncService) handleNewBlock(height uint64) bool {
	log.Infof("handle new block: %d\n", height)
	ret := s.handleBlockHeader(height)
	if !ret {
		log.Infof("handleNewBlock - handleBlockHeader on height :%d failed\n", height)
		return false
	}
	ret = s.fetchLockDepositEvents(height)
	if !ret {
		log.Infof("handleNewBlock - fetchLockDepositEvents on height :%d failed\n", height)
	}
	return true
}

func (s *SyncService) handleBlockHeader(height uint64) bool {
	// todo
	return true
}

func EncodeBigInt(b *big.Int) string {
	if b.Uint64() == 0 {
		return "00"
	}
	return hex.EncodeToString(b.Bytes())
}

// the workflow is: user -> LockProxy on zilliqa -> Cross Chain Manager -> emit event
// so here we need to filter out those transactions related to cross chain manager
// and parse the events, store them to local db, and commit them to the polynetwork
func (s *SyncService) fetchLockDepositEvents(height uint64) bool {
	transactions, err := s.zilSdk.GetTxnBodiesForTxBlock(strconv.FormatUint(height, 10))
	if err != nil {
		log.Infof("get transactions for tx block %d failed: %s\n", height, err.Error())
		return false
	}

	for _, transaction := range transactions {
		events := transaction.Receipt.EventLogs
		for _, event := range events {
			toAddr, _ := bech32.ToBech32Address(event.Address)
			log.Infof("to address: %s\n", toAddr)
			if toAddr == s.corssChainManagerAddress {
				log.Infof("found event on cross chain manager: %+v\n", event)
				// todo parse event to struct CrossTransfer
				crossTx := &CrossTransfer{}
				for _, param := range event.Params {
					switch param.VName {
					case "txId":
						index := big.NewInt(0)
						index.SetBytes(util.DecodeHex(param.Value.(string)))
						crossTx.txIndex = EncodeBigInt(index)
					case "toChainId":
						toChainId,_ := strconv.ParseUint(param.Value.(string),10,32)
						crossTx.toChain = uint32(toChainId)
					case "rawData":
						crossTx.value = []byte(param.Value.(string))
					}
				}
				crossTx.height = height
				crossTx.txId = util.DecodeHex(transaction.ID)
				log.Infof("parsed cross tx is: %+v\n",crossTx)
				sink := common.NewZeroCopySink(nil)
				crossTx.Serialization(sink)
				err1 := s.db.PutRetry(sink.Bytes())
				if err1 != nil {
					log.Errorf("fetchLockDepositEvents - this.db.PutRetry error: %s", err)
				}
				log.Infof("fetchLockDepositEvent -  height: %d", height)
			}
		}
	}

	return true
}

func (s *SyncService) MonitorChain() {
	log.Infof("start scan block at height: %d\n", s.currentHeight)
	fetchBlockTicker := time.NewTicker(time.Duration(s.cfg.ZilMonitorInterval) * time.Second)
	for {
		select {
		case <-fetchBlockTicker.C:
			txBlock, err := s.zilSdk.GetLatestTxBlock()
			if err != nil {
				log.Infof("MonitorChain - cannot get node hight, err: %s\n", err.Error())
				continue
			}
			log.Infof("current tx block height: %s\n", txBlock.Header.BlockNum)
			blockNumber, err2 := strconv.ParseUint(txBlock.Header.BlockNum, 10, 32)
			if err2 != nil {
				log.Infof("MonitorChain - cannot parse block height, err: %s\n", err2.Error())
			}
			if s.currentHeight >= blockNumber {
				log.Infof("current height is not changed, skip")
				continue
			}

			for s.currentHeight < blockNumber {
				s.handleNewBlock(s.currentHeight + 1)
				s.currentHeight++
			}

		case <-s.exitChan:
			return

		}

	}
}
