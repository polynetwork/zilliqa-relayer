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
	"strings"
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
	nonceManager           *NonceManager
}

func (p *PolySyncManager) init() bool {
	if p.currentHeight > 0 {
		log.Infof("PolySyncManager init - start height from flag: %d", p.currentHeight)
		return true
	}

	p.currentHeight = p.db.GetPolyHeight()
	log.Infof("PolySyncManager init - get poly height from local storage: %d", p.currentHeight)
	latestHeight := p.findLatestHeight()
	log.Infof("PolySyncManager init - get poly height from cross chain manager contract: %d", latestHeight)
	if latestHeight > p.currentHeight {
		p.currentHeight = latestHeight
		return true
	}
	log.Infof("PolySyncManager init - start height from flag: %d", p.currentHeight)
	return true
}

func (p *PolySyncManager) Run(enable bool) {
	if enable {
		go p.MonitorChain()
		go p.nonceManager.Run()
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
	var privateKeys []string
	zilSenderMap := make(map[string]*NonceAndSender, 0)

	for _, keystore := range keystores {
		var ks crypto.KeystoreV3
		err1 := json.Unmarshal([]byte(keystore), &ks)
		if err1 != nil {
			return nil, err1
		}
		pwd := keystorepwdset[strings.ToLower(ks.Address)]
		if pwd == nil {
			return nil, errors.New("NewPolySyncManager - there is no password for zilliqa.wallet: " + ks.Address)
		}
		privateKey, err2 := descryptor.DecryptPrivateKey(keystore, pwd.(string))
		if err2 != nil {
			return nil, errors.New("NewPolySyncManager - descrypt zilliqa.wallet error: " + err2.Error())
		}

		// init cross chain smart contract proxy
		wallet := account.NewWallet()
		wallet.AddByPrivateKey(privateKey)
		log.Infof("NewPolySyncManager get zilliqa wallet: %s", wallet.DefaultAccount.Address)
		proxy := &polynetwork.Proxy{
			ProxyAddr:  crossChainManagerProxy,
			ImplAddr:   crossChainManager,
			Wallet:     wallet,
			Client:     zilSdk,
			ChainId:    cfg.ZilConfig.ZilChainId,
			MsgVersion: cfg.ZilConfig.ZilMessageVersion,
		}

		addr := strings.Replace(ks.Address, "0x", "", 1)
		sender := &ZilSender{
			cfg:             cfg,
			zilSdk:          zilSdk,
			address:         addr,
			privateKey:      privateKey,
			polySdk:         polySdk,
			crossChainProxy: proxy,
			inUse:           false,
		}

		senders = append(senders, sender)

		balAndNonce, err3 := zilSdk.GetBalance(addr)
		if err3 != nil {
			log.Infof("NewPolySyncManager get address %s nonce error %s", addr, err3.Error())
			continue
		}

		privateKeyAndNonce := &NonceAndSender{
			Sender:     sender,
			LocalNonce: balAndNonce.Nonce,
		}

		privateKeys = append(privateKeys, privateKey)
		zilSenderMap[privateKey] = privateKeyAndNonce
	}

	nonceManager := &NonceManager{
		UpdateInterval:        30,
		ZilClient:             zilSdk,
		SentTransactions:      make(map[string]map[string]TransactionWithAge),
		ConfirmedTransactions: make(map[string][]string),
		SenderPrivateKeys:     privateKeys,
		ZilSenderMap:          zilSenderMap,
		CurrentIndex:          0,
		Cfg:                   cfg,
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
		nonceManager:           nonceManager,
	}, nil
}
