package service

import (
	"github.com/polynetwork/zilliqa-relayer/config"
	log "github.com/sirupsen/logrus"
	"time"
)

func (p *PolySyncManager) MonitorChain() {
	log.Infof("PolySyncManager MonitorChain - start scan block at height: %d\n", p.currentHeight)
	monitorTicker := time.NewTicker(time.Duration(p.cfg.PolyConfig.PolyMonitorInterval) * time.Second)
	var blockHandleResult bool
	for {
		select {
		case <-monitorTicker.C:
			latestHeight, err := p.polySdk.GetCurrentBlockHeight()
			if err != nil {
				log.Errorf("PolySyncManager MonitorChain - cannot get node hight, err: %s\n", err.Error())
				continue
			}
			latestHeight--
			if latestHeight-p.currentHeight < config.OntUsefulBlockNum {
				continue
			}
			log.Infof("PolySyncManager MonitorChain - poly chain current height: %d", latestHeight)
			blockHandleResult = true
			for p.currentHeight <= latestHeight-config.OntUsefulBlockNum {
				blockHandleResult = p.handleDepositEvents(p.currentHeight)
				if blockHandleResult == false {
					break
				}
				p.currentHeight++
			}
			if err = p.db.UpdatePolyHeight(p.currentHeight - 1); err != nil {
				log.Errorf("MonitorChain - failed to save height of poly: %v", err)
			}
		case <-p.exitChan:
			return

		}
	}
}

func (p *PolySyncManager) handleDepositEvents(height uint32) bool {
	// todo
	return true
}

func (p *PolySyncManager) findLatestHeight() uint32 {
	// todo get substate of cross chain manager
	return 0
}
