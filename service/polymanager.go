package service

import (
	"encoding/json"
	"errors"
	"github.com/Zilliqa/gozilliqa-sdk/crypto"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/db"
	"github.com/polynetwork/zilliqa-relayer/tools"
)

type PolySyncManager struct {
	currentHeight uint32
	polySdk       *poly_go_sdk.PolySdk
	exitChan      chan int
	cfg           *config.Config
	db            *db.BoltDB

	zilSdk  *provider.Provider
	senders []*ZilSender
}

func (p *PolySyncManager) Run() {

}

func NewPolySyncManager(cfg *config.Config, zilSdk *provider.Provider, polySdk *poly_go_sdk.PolySdk, boltDB *db.BoltDB) (*PolySyncManager, error) {
	keystores, err := tools.ReadLine(cfg.ZilConfig.KeyStorePath)
	keystorepwdset := cfg.ZilConfig.KeyStorePwdSet
	if err != nil {
		return nil, err
	}
	descryptor := crypto.NewDefaultKeystore()

	var senders []*ZilSender

	for _, keystore := range keystores {
		var ks crypto.KeystoreV3
		err1 := json.Unmarshal([]byte(keystore), &ks)
		if err1 != nil {
			return nil, err1
		}
		pwd := keystorepwdset[ks.Address]
		if pwd == nil {
			return nil, errors.New("NewPolySyncManager - there is no password for keystore: " + ks.Address)
		}
		privateKey, err2 := descryptor.DecryptPrivateKey(keystore, pwd.(string))
		if err2 != nil {
			return nil, errors.New("NewPolySyncManager - descrypt keystore error: " + err2.Error())
		}

		sender := &ZilSender{
			cfg:        cfg,
			zilSdk:     zilSdk,
			address:    ks.Address,
			privateKey: privateKey,
			polySdk:    polySdk,
		}

		senders = append(senders, sender)
	}

	return &PolySyncManager{
		currentHeight: cfg.PolyConfig.PolyMonitorInterval,
		polySdk:       polySdk,
		exitChan:      make(chan int),
		cfg:           cfg,
		db:            boltDB,
		zilSdk:        zilSdk,
		senders:       senders,
	}, nil
}
