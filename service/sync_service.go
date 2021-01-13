package service

import (
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	poly "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/db"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

type SyncService struct {
	relayAccount    *poly.Account
	relaySdk        *poly.PolySdk
	relaySyncHeight uint32
	zilAccount      *account.Account

	currentHeight            uint64
	zilSdk                   *provider.Provider
	corssChainManagerAddress string
	cfg                      *config.Config
	db                       *db.BoltDB
	exitChan                 chan int
}

func NewSyncService(cfg *config.Config) *SyncService {
	if !checkIfExist(cfg.Path) {
		os.Mkdir(cfg.Path, os.ModePerm)
	}
	boltDB, err := db.NewBoltDB(cfg.Path)
	if err != nil {
		log.Fatal("cannot init bolt db")
	}

	return &SyncService{
		db:                       boltDB,
		cfg:                      cfg,
		zilSdk:                   provider.NewProvider(cfg.ZilApiEndpoint),
		currentHeight:            uint64(cfg.ZilStartHeight),
		corssChainManagerAddress: cfg.CrossChainManagerContract,
	}
}

func (s *SyncService) Run() {
	go s.MonitorChain()
	waitToExit()
}

func checkIfExist(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}

func waitToExit() {
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
