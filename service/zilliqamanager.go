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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/core"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	poly "github.com/polynetwork/poly-go-sdk"
	sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	autils "github.com/polynetwork/poly/native/service/utils"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/db"
	log "github.com/sirupsen/logrus"
	"strconv"
)

/**
 * currentHeight's source is either from poly remote storage or forceHeight
 */
type ZilliqaSyncManager struct {
	polySigner               *poly.Account
	polySdk                  *poly.PolySdk
	relaySyncHeight          uint32
	zilAccount               *account.Account
	currentHeight            uint64
	currentDsBlockNum        uint64
	forceHeight              uint64
	zilSdk                   *provider.Provider
	crossChainManagerAddress string
	cfg                      *config.Config
	db                       *db.BoltDB
	exitChan                 chan int
	header4sync              [][]byte
}

func NewZilliqaSyncManager(cfg *config.Config, zilSdk *provider.Provider, polysdk *sdk.PolySdk, boltDB *db.BoltDB) (*ZilliqaSyncManager, error) {
	var wallet *sdk.Wallet
	var err error
	if !common.FileExisted(cfg.PolyConfig.PolyWalletFile) {
		wallet, err = polysdk.OpenWallet(cfg.PolyConfig.PolyWalletFile)
		if err != nil {
			return nil, err
		}
	} else {
		wallet, err = polysdk.OpenWallet(cfg.PolyConfig.PolyWalletFile)
		if err != nil {
			log.Errorf("NewZilliqaSyncManager - wallet open error: %s", err.Error())
			return nil, err
		}
	}
	signer, err := wallet.GetDefaultAccount([]byte(cfg.PolyConfig.PolyWalletPassword))
	if err != nil || signer == nil {
		signer, err = wallet.NewDefaultSettingAccount([]byte(cfg.PolyConfig.PolyWalletPassword))
		if err != nil {
			log.Errorf("NewETHManager - wallet password error")
			return nil, err
		}

		err = wallet.Save()
		if err != nil {
			return nil, err
		}
	}
	log.Infof("NewZilliqaSyncManager - poly address: %s", signer.Address.ToBase58())
	zilliqaSyncManager := &ZilliqaSyncManager{
		db:                       boltDB,
		cfg:                      cfg,
		exitChan:                 make(chan int),
		zilSdk:                   zilSdk,
		forceHeight:              cfg.ZilConfig.ZilForceHeight,
		crossChainManagerAddress: cfg.ZilConfig.CrossChainManagerContract,
		polySigner:               signer,
		polySdk:                  polysdk,
	}

	err = zilliqaSyncManager.init()
	if err != nil {
		return nil, err
	} else {
		return zilliqaSyncManager, nil
	}
}

func (s *ZilliqaSyncManager) Run(enable bool) {
	if enable {
		go s.MonitorChain()
	}
}

func (s *ZilliqaSyncManager) init() error {
	// get latest tx block from remote poly storage, thus we can know current tx block num and ds block num
	latestHeight := s.findLatestTxBlockHeight()
	log.Infof("ZilliqaSyncManager init - get latest tx block from poly, tx block height is: %d\n", latestHeight)

	if latestHeight == 0 {
		return fmt.Errorf("ZilliqaSyncManager init - the genesis block has not synced!")
	}

	if s.forceHeight > 0 && s.forceHeight < latestHeight {
		s.currentHeight = s.forceHeight
	} else {
		s.currentHeight = latestHeight
	}
	log.Infof("ZilliqaSyncManager init - start height: %d", s.currentHeight)
	s.getGenesisHeader()
	s.getMainChain(latestHeight)
	txBlockT, err := s.zilSdk.GetTxBlockVerbose(strconv.FormatUint(s.currentHeight, 10))
	if err != nil {
		return fmt.Errorf("ZilliqaSyncManager init - get tx block error: %s", err.Error())
	}
	dsBlockNum, _ := strconv.ParseUint(txBlockT.Header.DSBlockNum, 10, 64)
	s.currentDsBlockNum = dsBlockNum
	log.Infof("ZilliqaSyncManager init - current ds block height is: %d\n", s.currentDsBlockNum)
	return nil
}

func (s *ZilliqaSyncManager) getGenesisHeader() {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], s.cfg.ZilConfig.SideChainId)
	key := append([]byte(scom.GENESIS_HEADER), sideChainIdBytes[:]...)
	contractAddress := autils.HeaderSyncContractAddress
	result, err := s.polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil {
		log.Infof("cannot obtain latest genesis header\n")
	}
	if result == nil || len(result) == 0 {
		log.Infof("0-length result from gensis header query\n")
	}
	// Turns out the genesis header is a string.
	block := new(core.TxBlock)
	genesisString := string(result[:])
	log.Infof("----- GENESIS HEADER ----- \n")
	log.Infof("%s", genesisString)
	log.Infof("====== END GENESIS HEADER ===== \n")
	err = json.Unmarshal(result, block)
	if err != nil {
		log.Infof("!!!!! Could not unmarshal genesis - %s", err)
	} else {
		log.Infof("Get main chain for %d", block.BlockHeader.BlockNum)
		s.getMainChain(block.BlockHeader.BlockNum)
		log.Infof("GetHeaderIndex for %d:%x", block.BlockHeader.BlockNum, block.BlockHash)
		s.getHeaderIndex(block.BlockHeader.BlockNum, block.BlockHash[:])
		dsBlockNum := block.BlockHeader.DSBlockNum
		log.Infof("Get DS Block %d from chain", dsBlockNum)
		dsb, err := s.zilSdk.GetDsBlock(strconv.Itoa(int(dsBlockNum)))
		if err != nil {
			log.Infof("Could not get DS Block - %s", err)
		} else {
			dsbval := core.NewDsBlockFromDsBlockT(dsb)
			log.Infof("getDsBlockHeader %x", dsbval.BlockHash)
			s.getDsBlockHeader(dsBlockNum, dsbval.BlockHash[:])
		}
		log.Infof("XXX Done")
	}
}

// Get the DS Block Header
func (s *ZilliqaSyncManager) getDsBlockHeader(dsBlkNum uint64, hash []byte) {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], s.cfg.ZilConfig.SideChainId)
	key := append([]byte(scom.HEADER_INDEX), sideChainIdBytes[:]...)
	key = append(key, hash...)
	contractAddress := autils.HeaderSyncContractAddress
	result, err := s.polySdk.GetStorage(contractAddress.ToHexString(), key)
	log.Infof("---- DS Header hash with blknum %d hash %x", dsBlkNum, hash)
	if err != nil {
		log.Infof("==== FAILED %s", err)
	}
	if result == nil || len(result) == 0 {
		log.Infof("==== 0-length or empty result")
	} else {
		log.Infof("==== Retrieved: %x", result)
	}

}

// Given a tx block number, try to obtain MAIN_CHAIN
func (s *ZilliqaSyncManager) getMainChain(blknum uint64) {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], s.cfg.ZilConfig.SideChainId)
	key := append([]byte(scom.MAIN_CHAIN), sideChainIdBytes[:]...)
	var blkNumBytes [8]byte
	binary.LittleEndian.PutUint64(blkNumBytes[:], blknum)
	key = append(key, blkNumBytes[:]...)
	contractAddress := autils.HeaderSyncContractAddress
	result, err := s.polySdk.GetStorage(contractAddress.ToHexString(), key)
	log.Infof("---- MAIN_CHAIN with blknum %d", blknum)
	if err != nil {
		log.Infof("==== FAILED %s", err)
	}
	if result == nil || len(result) == 0 {
		log.Infof("==== 0-length or empty result")
	} else {
		log.Infof("==== Retrieved: %x", result)
	}
}

// Given a tx block hash, see if the header index is in storage.
func (s *ZilliqaSyncManager) getHeaderIndex(blknum uint64, hash []byte) {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], s.cfg.ZilConfig.SideChainId)
	key := append([]byte(scom.HEADER_INDEX), sideChainIdBytes[:]...)
	key = append(key, hash...)
	contractAddress := autils.HeaderSyncContractAddress
	result, err := s.polySdk.GetStorage(contractAddress.ToHexString(), key)
	log.Infof("---- Header hash with blknum %d hash %x", blknum, hash)
	if err != nil {
		log.Infof("==== FAILED %s", err)
	}
	if result == nil || len(result) == 0 {
		log.Infof("==== 0-length or empty result")
	} else {
		log.Infof("==== Retrieved: %x", result)
	}
}

func (s *ZilliqaSyncManager) checkDSBlockInStorage(blk uint64) {
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], s.cfg.ZilConfig.SideChainId)
	var blkc [8]byte
	binary.LittleEndian.PutUint64(blkc[:], blk)
	key := append(sideChainIdBytes[:], []byte("dsComm")...)
	key = append(key, blkc[:]...)
	result, err := s.polySdk.GetStorage(autils.HeaderSyncContractAddress.ToHexString(), key)
	if err != nil {
		log.Infof("Couldn't retrieve polynet storage for DS Block %d: %s", blk, err.Error())
	}
	if result == nil || len(result) == 0 {
		log.Infof("no DSC stored for ds block %d", blk)
	} else {
		log.Infof("Something there for ds block %d", blk)
	}
}

func (s *ZilliqaSyncManager) findLatestTxBlockHeight() uint64 {
	// try to get key
	var sideChainIdBytes [8]byte
	binary.LittleEndian.PutUint64(sideChainIdBytes[:], s.cfg.ZilConfig.SideChainId)
	contractAddress := autils.HeaderSyncContractAddress
	key := append([]byte(scom.CURRENT_HEADER_HEIGHT), sideChainIdBytes[:]...)
	// try to get storage
	result, err := s.polySdk.GetStorage(contractAddress.ToHexString(), key)
	if err != nil {
		log.Infof("get latest tx block from poly failed,err: %s\n", err.Error())
		return 0
	}
	if result == nil || len(result) == 0 {
		return 0
	} else {
		return binary.LittleEndian.Uint64(result)
	}
}
