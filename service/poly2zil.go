package service

import (
	"encoding/hex"
	"encoding/json"
	"github.com/Zilliqa/gozilliqa-sdk/bech32"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/consensus/vbft/config"
	polytypes "github.com/polynetwork/poly/core/types"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/tools"
	log "github.com/sirupsen/logrus"
	"strconv"
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
				log.Errorf("PolySyncManager MonitorChain - failed to save height of poly: %v", err)
			}
		case <-p.exitChan:
			return

		}
	}
}

func (p *PolySyncManager) handleDepositEvents(height uint32) bool {
	lastEpoch := p.findLatestHeight()
	hdr, err := p.polySdk.GetHeaderByHeight(height + 1)
	if err != nil {
		log.Errorf("PolySyncManager handleBlockHeader - GetNodeHeader on height :%d failed", height)
		return false
	}
	isCurr := lastEpoch < height+1
	info := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(hdr.ConsensusPayload, info); err != nil {
		log.Errorf("PolySyncManager failed to unmarshal ConsensusPayload for height %d: %v", height+1, err)
		return false
	}
	isEpoch := hdr.NextBookkeeper != common.ADDRESS_EMPTY && info.NewChainConfig != nil
	var (
		anchor *polytypes.Header
		hp     string
	)
	if !isCurr {
		anchor, _ = p.polySdk.GetHeaderByHeight(lastEpoch + 1)
		proof, _ := p.polySdk.GetMerkleProof(height+1, lastEpoch+1)
		hp = proof.AuditPath
	} else if isEpoch {
		anchor, _ = p.polySdk.GetHeaderByHeight(height + 2)
		proof, _ := p.polySdk.GetMerkleProof(height+1, height+2)
		hp = proof.AuditPath
	}
	cnt := 0
	events, err := p.polySdk.GetSmartContractEventByBlock(height)
	for err != nil {
		log.Errorf("PolySyncManager handleDepositEvents - get block event at height:%d error: %s", height, err.Error())
		return false
	}
	for _, event := range events {
		for _, notify := range event.Notify {
			if notify.ContractAddress == p.cfg.PolyConfig.EntranceContractAddress {
				states := notify.States.([]interface{})
				method, _ := states[0].(string)
				if method != "makeProof" {
					continue
				}
				if uint64(states[2].(float64)) != p.cfg.ZilConfig.SideChainId {
					continue
				}
				proof, err := p.polySdk.GetCrossStatesProof(hdr.Height-1, states[5].(string))
				if err != nil {
					log.Errorf("handleDepositEvents - failed to get proof for key %s: %v", states[5].(string), err)
					continue
				}
				auditpath, _ := hex.DecodeString(proof.AuditPath)
				value, _, _, _ := tools.ParseAuditpath(auditpath)
				param := &common2.ToMerkleValue{}
				if err := param.Deserialization(common.NewZeroCopySource(value)); err != nil {
					log.Errorf("handleDepositEvents - failed to deserialize MakeTxParam (value: %x, err: %v)", value, err)
					continue
				}
				var isTarget bool
				if len(p.cfg.TargetContracts) > 0 {
					// todo assuming ToContractAddress is not bech32
					// handle error
					toContractStr, _ := bech32.ToBech32Address(util.EncodeHex(param.MakeTxParam.ToContractAddress))
					for _, v := range p.cfg.TargetContracts {
						toChainIdArr, ok := v[toContractStr]
						if ok {
							if len(toChainIdArr["inbound"]) == 0 {
								isTarget = true
								break
							}
							for _, id := range toChainIdArr["inbound"] {
								if id == param.FromChainID {
									isTarget = true
									break
								}
							}
							if isTarget {
								break
							}
						}
					}
					if !isTarget {
						continue
					}
				}
				cnt++
				sender := p.selectSender()
				log.Infof("sender %s is handling poly tx ( hash: %s, height: %d )",
					sender.acc, event.TxHash, height)
				// temporarily ignore the error for tx
				sender.commitDepositEventsWithHeader(hdr, param, hp, anchor, event.TxHash, auditpath)
				//if !sender.commitDepositEventsWithHeader(hdr, param, hp, anchor, event.TxHash, auditpath) {
				//	return false
				//}
			}
		}
	}

	return true
}

func (p *PolySyncManager) selectSender() *ZilSender {
	return &ZilSender{}
}

type ZilSender struct {
	acc string
}

func (sender *ZilSender) commitDepositEventsWithHeader(header *polytypes.Header, param *common2.ToMerkleValue, headerProof string, anchorHeader *polytypes.Header, polyTxHash string, rawAuditPath []byte) {
	// verifyHeaderAndExecuteTx
	var (
		sigs       []byte
		headerData []byte
	)
	if anchorHeader != nil && headerProof != "" {
		for _, sig := range anchorHeader.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			newsig, _ := signature.ConvertToEthCompatible(temp)
			sigs = append(sigs, newsig...)
		}
	} else {
		for _, sig := range header.SigData {
			temp := make([]byte, len(sig))
			copy(temp, sig)
			newsig, _ := signature.ConvertToEthCompatible(temp)
			sigs = append(sigs, newsig...)
		}
	}
}

type EpochStartHeight struct {
	CurEpochStartHeight string `json:"curEpochStartHeight"`
}

type EpochStartHeightRep struct {
	Id      string           `json:"id"`
	JsonRpc string           `json:"jsonrpc"`
	Result  EpochStartHeight `json:"result"`
}

func (p *PolySyncManager) findLatestHeight() uint32 {
	ccm, err := bech32.FromBech32Addr(p.cfg.ZilConfig.CrossChainManagerContract)
	if err != nil {
		log.Errorf("PolySyncManager FindLatestHeight -  failed to convert cross chain manager contract address: %s\n", err.Error())
		return 0
	}
	curEpochStartHeight, err1 := p.zilSdk.GetSmartContractSubState(ccm, "curEpochStartHeight", []interface{}{})
	if err1 != nil {
		log.Errorf("PolySyncManager FindLatestHeight -  faild to get current epoch start height: %s\n", err1.Error())
		return 0
	}

	var epochStartHeightRep EpochStartHeightRep
	err3 := json.Unmarshal([]byte(curEpochStartHeight), &epochStartHeightRep)
	if err3 != nil {
		log.Errorf("PolySyncManager FindLatestHeight -  faild to unmarshal current epoch start height: %s\n", err3.Error())
		return 0
	}

	height, err2 := strconv.ParseUint(epochStartHeightRep.Result.CurEpochStartHeight, 10, 32)
	if err2 != nil {
		log.Errorf("PolySyncManager FindLatestHeight -  faild to parse epoch start height: %s\n", err2.Error())
		return 0
	}
	return uint32(height)
}
