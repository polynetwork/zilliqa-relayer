package service

import (
	"log"
	"strconv"
	"time"
)

func (s *SyncService) handleNewBlock(height uint64) bool {
	log.Printf("handle new block: %d\n", height)
	ret := s.handleBlockHeader(height)
	if !ret {
		log.Printf("handleNewBlock - handleBlockHeader on height :%d failed\n", height)
		return false
	}
	ret = s.fetchLockDepositEvents(height)
	if !ret {
		log.Printf("handleNewBlock - fetchLockDepositEvents on height :%d failed\n", height)
	}
	return true
}

func (s *SyncService) handleBlockHeader(height uint64) bool {
	// todo
	return true
}

func (s *SyncService) fetchLockDepositEvents(height uint64) bool {
	// todo
	return true
}

func (s *SyncService) MonitorChain() {
	log.Printf("start scan block at height: %d\n", s.currentHeight)
	fetchBlockTicker := time.NewTicker(time.Duration(s.cfg.ZilMonitorInterval) * time.Second)
	for {
		select {
		case <-fetchBlockTicker.C:
			txBlock, err := s.zilSdk.GetLatestTxBlock()
			if err != nil {
				log.Printf("MonitorChain - cannot get node hight, err: %s\n", err.Error())
				continue
			}
			log.Printf("current tx block height: %s\n", txBlock.Header.BlockNum)
			blockNumber, err2 := strconv.ParseUint(txBlock.Header.BlockNum, 10, 32)
			if err2 != nil {
				log.Printf("MonitorChain - cannot parse block height, err: %s\n", err2.Error())
			}
			if s.currentHeight >= blockNumber {
				log.Printf("cuurent height is not changed, skip")
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
