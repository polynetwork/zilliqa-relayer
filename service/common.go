package service

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func WaitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("Zilliqa Relayer received exit signal: %v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
