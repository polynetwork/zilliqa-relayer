package service

import (
	"encoding/json"
	"errors"
	"github.com/Zilliqa/gozilliqa-sdk/account"
	"github.com/Zilliqa/gozilliqa-sdk/crosschain/polynetwork"
	"github.com/Zilliqa/gozilliqa-sdk/crypto"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/ontio/ontology/common/log"
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

	zilSdk                 *provider.Provider
	crossChainManager      string
	crossChainManagerProxy string
	senders                []*ZilSender
}

func (p *PolySyncManager) init() bool {
	if p.currentHeight > 0 {
		log.Infof("PolySyncManager init - start height from flag: %d\n", p.currentHeight)
		return true
	}

	p.currentHeight = p.db.GetPolyHeight()
	latestHeight := p.findLatestHeight()
	if latestHeight > p.currentHeight {
		p.currentHeight = latestHeight
		log.Infof("PolyManager init - latest height from cross chain manager: %d\n", p.currentHeight)
		return true
	}

	log.Infof("PolyManager init - latest height from DB: %d\n", p.currentHeight)
	return true
}

func (p *PolySyncManager) Run(enable bool) {
	if enable {
		go p.MonitorChain()
	}
}

func NewPolySyncManager(cfg *config.Config, zilSdk *provider.Provider, polySdk *poly_go_sdk.PolySdk, boltDB *db.BoltDB, crossChainManager, crossChainManagerProxy string) (*PolySyncManager, error) {
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
			return nil, errors.New("NewPolySyncManager - there is no password for admin.keystore: " + ks.Address)
		}
		privateKey, err2 := descryptor.DecryptPrivateKey(keystore, pwd.(string))
		if err2 != nil {
			return nil, errors.New("NewPolySyncManager - descrypt admin.keystore error: " + err2.Error())
		}

		// init cross chain smart contract proxy
		wallet := account.NewWallet()
		wallet.AddByPrivateKey(privateKey)
		proxy := &polynetwork.Proxy{
			ProxyAddr:  crossChainManagerProxy,
			ImplAddr:   crossChainManager,
			Wallet:     wallet,
			Client:     zilSdk,
			ChainId:    cfg.ZilConfig.ZilChainId,
			MsgVersion: cfg.ZilConfig.ZilMessageVersion,
		}

		sender := &ZilSender{
			cfg:             cfg,
			zilSdk:          zilSdk,
			address:         ks.Address,
			privateKey:      privateKey,
			polySdk:         polySdk,
			crossChainProxy: proxy,
		}

		senders = append(senders, sender)
	}

	return &PolySyncManager{
		currentHeight:          cfg.PolyConfig.PolyStartHeight,
		polySdk:                polySdk,
		exitChan:               make(chan int),
		cfg:                    cfg,
		db:                     boltDB,
		zilSdk:                 zilSdk,
		crossChainManager:      crossChainManager,
		crossChainManagerProxy: crossChainManagerProxy,
		senders:                senders,
	}, nil
}
