package service

import (
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/Zilliqa/gozilliqa-sdk/util"
	"github.com/ontio/ontology-crypto/signature"
	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	polytypes "github.com/polynetwork/poly/core/types"
	common2 "github.com/polynetwork/poly/native/service/cross_chain_manager/common"
	"github.com/polynetwork/zilliqa-relayer/config"
	log "github.com/sirupsen/logrus"
)

type ZilSender struct {
	cfg        *config.Config
	zilSdk     *provider.Provider
	address    string //non-bech32 address
	privateKey string

	polySdk *poly_go_sdk.PolySdk
}

func (sender *ZilSender) commitDepositEventsWithHeader(header *polytypes.Header, param *common2.ToMerkleValue, headerProof string, anchorHeader *polytypes.Header, polyTxHash string, rawAuditPath []byte) bool {
	// verifyHeaderAndExecuteTx
	var (
		sigs []byte
		//headerData []byte
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

	// todo ensure that TxHash is bytes of hash, not utf8 bytes
	exist := sender.checkIfFromChainTxExist(param.FromChainID, util.EncodeHex(param.TxHash))
	if exist {
		log.Infof("ZilSender commitDepositEventsWithHeader - already relayed to zil: (from_chain_id: %d, from_txhash: %x, param.TxHash: %x\n)", param.FromChainID, param.TxHash, param.MakeTxParam.TxHash)
		return true
	}

	// todo

	return true

}
