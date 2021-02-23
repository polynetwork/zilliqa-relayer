package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/bech32"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/polynetwork/poly/common"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/zilliqa-relayer/tools"
	log "github.com/sirupsen/logrus"
	"math/big"
	"strconv"
	"strings"
	"time"
)

/**
 * handle new block:
 * 1. commit tx block and ds block
 * 2. filter deposit event, put into local database
 * 3. from database, handle deposit event, get proof and commit to poly
 */
func (s *ZilliqaSyncManager) MonitorChain() {
	log.Infof("ZilliqaSyncManager MonitorChain - start scan block at height: %d\n", s.currentHeight)
	fetchBlockTicker := time.NewTicker(time.Duration(s.cfg.ZilConfig.ZilMonitorInterval) * time.Second)
	var blockHandleResult bool
	for {
		select {
		case <-fetchBlockTicker.C:
			txBlock, err := s.zilSdk.GetLatestTxBlock()
			if err != nil {
				log.Errorf("ZilliqaSyncManager MonitorChain - cannot get node hight, err: %s\n", err.Error())
				continue
			}
			log.Infof("ZilliqaSyncManager MonitorChain - current tx block height: %s\n", txBlock.Header.BlockNum)
			blockNumber, err2 := strconv.ParseUint(txBlock.Header.BlockNum, 10, 32)
			if err2 != nil {
				log.Errorf("ZilliqaSyncManager MonitorChain - cannot parse block height, err: %s\n", err2.Error())
			}
			if s.currentHeight >= blockNumber {
				log.Infof("ZilliqaSyncManager MonitorChain - current height is not changed, skip")
				continue
			}

			blockHandleResult = true
			for s.currentHeight < blockNumber {
				s.handleNewBlock(s.currentHeight + 1)
				s.currentHeight++

				// todo enable this
				if uint32(len(s.header4sync)) > s.cfg.ZilConfig.ZilHeadersPerBatch {
					log.Infof("ZilliqaSyncManager MonitorChain - commit header")
					//if res := s.commitHeader(); res != 0 {
					//	blockHandleResult = false
					//	break
					//}
				}
			}

			if blockHandleResult && len(s.header4sync) > 0 {
				// todo enable this
				// s.commitHeader()
			}

		case <-s.exitChan:
			return
		}
	}
}

func (s *ZilliqaSyncManager) handleNewBlock(height uint64) bool {
	log.Infof("ZilliqaSyncManager handle new block: %d\n", height)
	ret := s.handleBlockHeader(height)
	if !ret {
		log.Infof("ZilliqaSyncManager handleNewBlock - handleBlockHeader on height :%d failed\n", height)
		return false
	}
	ret = s.fetchLockDepositEvents(height)
	if !ret {
		log.Infof("ZilliqaSyncManager handleNewBlock - fetchLockDepositEvents on height :%d failed\n", height)
	}
	return true
}

func (s *ZilliqaSyncManager) handleBlockHeader(height uint64) bool {
	log.Infof("ZilliqaSyncManager handle new block header height is: %d\n", height)
	txBlockT, err := s.zilSdk.GetTxBlockVerbose(strconv.FormatUint(height, 10))
	if err != nil {
		log.Errorf("ZilliqaSyncManager - handleBlockHeader error: %s", err)
		return false
	}
	txBlock := core.NewTxBlockFromTxBlockT(txBlockT)

	// todo consider very special case
	// tx block revert to last ds epoch, but ds block don't
	if txBlock.BlockHeader.DSBlockNum > s.currentDsBlockNum {
		dsBlock, err := s.zilSdk.GetDsBlockVerbose(strconv.FormatUint(txBlock.BlockHeader.DSBlockNum, 10))
		if err != nil {
			log.Errorf("ZilliqaSyncManager - handleBlockHeader get ds block error: %s", err)
			return false
		}
		txBlockOrDsBlock := core.TxBlockOrDsBlock{
			DsBlock: core.NewDsBlockFromDsBlockT(dsBlock),
		}
		rawBlock, _ := json.Marshal(txBlockOrDsBlock)
		log.Infof("ZilliqaSyncManager handle new block header: %s\n", rawBlock)
		s.header4sync = append(s.header4sync, rawBlock)
		s.currentDsBlockNum++
	}

	txBlockOrDsBlock := core.TxBlockOrDsBlock{
		TxBlock: txBlock,
	}
	rawBlock, err2 := json.Marshal(txBlockOrDsBlock)
	if err2 != nil {
		log.Errorf("ZilliqaSyncManager - handleBlockHeader marshal block error: %s", err2)
		return false
	}
	log.Debugf("ZilliqaSyncManager handle new block header: %s\n", rawBlock)
	blockHash := txBlock.BlockHash[:]
	log.Infof("ZilliqaSyncManager handleBlockHeader - header hash: %s\n", util.EncodeHex(blockHash))
	raw, _ := s.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
		append(append([]byte(scom.MAIN_CHAIN), autils.GetUint64Bytes(s.cfg.ZilConfig.SideChainId)...), autils.GetUint64Bytes(height)...))
	if len(raw) == 0 || bytes.Equal(raw, blockHash) {
		s.header4sync = append(s.header4sync, rawBlock)
	}
	return true
}

// the workflow is: user -> LockProxy on zilliqa -> Cross Chain Manager -> emit event
// so here we need to filter out those transactions related to cross chain manager
// and parse the events, store them to local db, and commit them to the polynetwork
func (s *ZilliqaSyncManager) fetchLockDepositEvents(height uint64) bool {
	transactions, err := s.zilSdk.GetTxnBodiesForTxBlock(strconv.FormatUint(height, 10))
	if err != nil {
		if strings.Contains(err.Error(), "TxBlock has no transactions") {
			log.Infof("ZilliqaSyncManager no transaction in block %d\n", height)
			return true
		} else {
			log.Infof("ZilliqaSyncManager get transactions for tx block %d failed: %s\n", height, err.Error())
			return false
		}
	}

	for _, transaction := range transactions {
		events := transaction.Receipt.EventLogs
		for _, event := range events {
			toAddr, _ := bech32.ToBech32Address(event.Address)
			if toAddr == s.crossChainManagerAddress {
				log.Infof("ZilliqaSyncManager found event on cross chain manager: %+v\n", event)
				// todo parse event to struct CrossTransfer
				crossTx := &CrossTransfer{}
				for _, param := range event.Params {
					switch param.VName {
					case "txId":
						index := big.NewInt(0)
						index.SetBytes(util.DecodeHex(param.Value.(string)))
						crossTx.txIndex = tools.EncodeBigInt(index)
					case "toChainId":
						toChainId, _ := strconv.ParseUint(param.Value.(string), 10, 32)
						crossTx.toChain = uint32(toChainId)
					case "rawData":
						crossTx.value = util.DecodeHex(param.Value.(string))
					}
				}
				crossTx.height = height
				crossTx.txId = util.DecodeHex(transaction.ID)
				log.Infof("ZilliqaSyncManager parsed cross tx is: %+v\n", crossTx)
				sink := common.NewZeroCopySink(nil)
				crossTx.Serialization(sink)
				err1 := s.db.PutRetry(sink.Bytes())
				if err1 != nil {
					log.Errorf("ZilliqaSyncManager fetchLockDepositEvents - this.db.PutRetry error: %s", err)
				}
				log.Infof("ZilliqaSyncManager fetchLockDepositEvent -  height: %d", height)
			}
		}
	}

	return true
}

func (s *ZilliqaSyncManager) handleLockDepositEvents(height uint64) error {
	retryList, err := s.db.GetAllRetry()
	if err != nil {
		return fmt.Errorf("ZilliqaSyncManager - handleLockDepositEvents - this.db.GetAllRetry error: %s", err)
	}
	for _, v := range retryList {
		time.Sleep(time.Second * 1)
		crosstx := new(CrossTransfer)
		err := crosstx.Deserialization(common.NewZeroCopySource(v))
		if err != nil {
			log.Errorf("ZilliqaSyncManager - handleLockDepositEvents - retry.Deserialization error: %s", err)
			continue
		}
		heightString := new(string)
		*heightString = strconv.FormatUint(height, 10)
		proof, err := s.zilSdk.GetStateProof(s.cfg.ZilConfig.CrossChainManagerContract, "zilToPolyTxHashMap", []string{crosstx.txIndex}, heightString)
		if err != nil {
			return fmt.Errorf("ZilliqaSyncManager - handleLockDepositEvents - get proof from api error: %s", err)
		}

		zilProof := &ZILProof{
			AccountProof: proof.AccountProof,
		}

		storageProof := StorageProof{
			Key:   core.GenerateStorageKey("zilToPolyTxHashMap", []string{crosstx.txIndex}),
			Value: crosstx.value,
			Proof: proof.StateProof,
		}

		zilProof.StorageProofs = []StorageProof{storageProof}
		proofRaw, _ := json.Marshal(zilProof)

		// commit proof
		proofString, _ := json.Marshal(proof)
		log.Debugf("ZilliqaSyncManager - handleLockDepositEvents commit proof, height: %d, proof: %s, value: %s, txhash: %s", height, proofString, util.EncodeHex(crosstx.value), util.EncodeHex(crosstx.txId))
		tx, err := s.polySdk.Native.Ccm.ImportOuterTransfer(
			s.cfg.ZilConfig.SideChainId,
			crosstx.value,
			uint32(height),
			proofRaw,
			util.DecodeHex(s.polySigner.Address.ToHexString()),
			[]byte{},
			s.polySigner)

		if err != nil {
			if strings.Contains(err.Error(), "ZilliqaSyncManager - handleLockDepositEvents chooseUtxos, current utxo is not enough") {
				log.Infof("ZilliqaSyncManager - handleLockDepositEvents handleLockDepositEvents - invokeNativeContract error: %s", err)
				continue
			} else {
				if err := s.db.DeleteRetry(v); err != nil {
					log.Errorf("ZilliqaSyncManager - handleLockDepositEvents handleLockDepositEvents handleLockDepositEvents - this.db.DeleteRetry error: %s", err)
				}
				if strings.Contains(err.Error(), "tx already done") {
					log.Debugf("ZilliqaSyncManager - handleLockDepositEvents handleLockDepositEvents handleLockDepositEvents - eth_tx %s already on poly", util.EncodeHex(crosstx.txId))
				} else {
					log.Errorf("ZilliqaSyncManager handleLockDepositEvents invokeNativeContract error for eth_tx %s: %s", util.EncodeHex(crosstx.txId), err)
				}
				continue
			}
		} else {
			log.Infof("ZilliqaSyncManager - handleLockDepositEvents commitProof - send transaction to poly chain: ( poly_txhash: %s, eth_txhash: %s, height: %d )",
				tx.ToHexString(), util.EncodeHex(crosstx.txId), height)
			txHash := tx.ToHexString()
			err = s.db.PutCheck(txHash, v)
			err = s.db.PutCheck(txHash, v)
			if err != nil {
				log.Errorf("ZilliqaSyncManager handleLockDepositEvents - this.db.PutCheck error: %s", err)
			}
			err = s.db.DeleteRetry(v)
			if err != nil {
				log.Errorf("ZilliqaSyncManager handleLockDepositEvents - this.db.PutCheck error: %s", err)
			}
			log.Infof("ZilliqaSyncManager handleLockDepositEvents - syncProofToAlia txHash is %s", txHash)
		}
	}

	return nil
}

// should be the same as relayer side
type ZILProof struct {
	AccountProof  []string       `json:"accountProof"`
	StorageProofs []StorageProof `json:"storageProof"`
}

// key should be storage key (in zilliqa)
type StorageProof struct {
	Key   []byte   `json:"key"`
	Value []byte   `json:"value"`
	Proof []string `json:"proof"`
}

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

func (this *CrossTransfer) Deserialization(source *common.ZeroCopySource) error {
	txIndex, eof := source.NextString()
	if eof {
		return fmt.Errorf("Waiting deserialize txIndex error")
	}
	txId, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize txId error")
	}
	value, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Waiting deserialize value error")
	}
	toChain, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("Waiting deserialize toChain error")
	}
	height, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Waiting deserialize height error")
	}
	this.txIndex = txIndex
	this.txId = txId
	this.value = value
	this.toChain = toChain
	this.height = height
	return nil
}

func (s *ZilliqaSyncManager) commitHeader() int {
	tx, err := s.polySdk.Native.Hs.SyncBlockHeader(
		s.cfg.ZilConfig.SideChainId,
		s.polySigner.Address,
		s.header4sync,
		s.polySigner,
	)

	if err != nil {
		errDesc := err.Error()
		if strings.Contains(errDesc, "get the parent block failed") || strings.Contains(errDesc, "missing required field") {
			log.Warnf("commitHeader - send transaction to poly chain err: %s", errDesc)
			s.rollBackToCommAncestor()
			return 0
		} else {
			log.Errorf("commitHeader - send transaction to poly chain err: %s", errDesc)
			return 1
		}
	}

	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	for range tick.C {
		h, _ = s.polySdk.GetBlockHeightByTxHash(tx.ToHexString())
		curr, _ := s.polySdk.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}
	}

	log.Infof("commitHeader - send transaction %s to poly chain and confirmed on height %d", tx.ToHexString(), h)
	s.header4sync = make([][]byte, 0)
	return 0
}

func (s *ZilliqaSyncManager) rollBackToCommAncestor() {
	for ; ; s.currentHeight-- {
		raw, err := s.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(),
			append(append([]byte(scom.MAIN_CHAIN), autils.GetUint64Bytes(s.cfg.ZilConfig.SideChainId)...), autils.GetUint64Bytes(s.currentHeight)...))
		if len(raw) == 0 || err != nil {
			continue
		}
		txBlockT, err2 := s.zilSdk.GetTxBlockVerbose(strconv.FormatUint(s.currentHeight, 10))
		if err2 != nil {
			log.Errorf("rollBackToCommAncestor - failed to get header by number, so we wait for one second to retry: %v", err2)
			time.Sleep(time.Second)
			s.currentHeight++
			continue
		}
		blockHeader := core.NewTxBlockFromTxBlockT(txBlockT).BlockHeader
		if bytes.Equal(util.Sha256(blockHeader.Serialize()), raw) {
			bs, _ := json.Marshal(blockHeader)
			log.Infof("rollBackToCommAncestor - find the common ancestor: %s(number: %d)", bs, s.currentHeight)
			break
		}

		s.header4sync = make([][]byte, 0)

	}
}
