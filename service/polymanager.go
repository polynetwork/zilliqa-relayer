package service

import (
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/db"
)

type PolySyncManager struct {
	currentHeight uint32
	polySdk       *poly_go_sdk.PolySdk
	exitChan      chan int
	cfg           *config.Config
	db            *db.BoltDB

	zilSdk *provider.Provider
}

func (p *PolySyncManager) Run() {
	waitToExit()
}
