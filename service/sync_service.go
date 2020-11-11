package service

import (
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	poly "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/zilliqa-relayer/db"
	"log"
	"os"
)

type SyncService struct {
	relayAccount    *poly.Account
	relaySdk        *poly.PolySdk
	relaySyncHeight uint32

	zilAccount *account.Account
	zilSdk     *provider.Provider

	db *db.BoltDB
}

func NewSyncService(replayAccount *poly.Account, replaySdk *poly.PolySdk, zilAccount *account.Account, zilSdk *provider.Provider, path string) *SyncService {
	if !checkIfExist(path) {
		os.Mkdir(path, os.ModePerm)
	}
	boltDB, err := db.NewBoltDB(path)
	if err != nil {
		log.Fatal("cannot init bolt db")
	}

	return &SyncService{
		relayAccount: replayAccount,
		relaySdk:     replaySdk,
		zilAccount:   zilAccount,
		zilSdk:       zilSdk,
		db:           boltDB,
	}
}

func checkIfExist(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}
