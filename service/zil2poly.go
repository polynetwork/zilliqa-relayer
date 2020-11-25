package service

import (
	"log"
	"time"
)

func (s *SyncService) Zil2Poly() {
	for {
		log.Println("start scan block")
		time.Sleep(time.Duration(s.cfg.ZilScanInterval) * time.Second)
	}
}
