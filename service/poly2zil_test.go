package service

import (
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/magiconair/properties/assert"
	"github.com/polynetwork/zilliqa-relayer/config"
	"testing"
)

var p *PolySyncManager
var s *ZilSender

func init() {
	zilSdk := provider.NewProvider("https://polynetworkcc3dcb2-5-api.dev.z7a.xyz")
	p = &PolySyncManager{
		zilSdk: zilSdk,
		cfg: &config.Config{ZilConfig: &config.ZILConfig{
			CrossChainManagerContract: "zil1ur4vwcmcz3jqypksgq7qeju2sk5jrskzaadau5",
		}},
	}
	s = &ZilSender{
		cfg: &config.Config{ZilConfig: &config.ZILConfig{
			CrossChainManagerContract: "zil1ur4vwcmcz3jqypksgq7qeju2sk5jrskzaadau5",
		}},
		zilSdk: zilSdk,
		acc:    "zil1ur4vwcmcz3jqypksgq7qeju2sk5jrskzaadau5",
	}
}

func TestPolySyncManager_FindLatestHeight(t *testing.T) {
	fmt.Println(p.findLatestHeight())
}

func TestZilSender_CheckIfFromChainTxExist(t *testing.T) {
	exist := s.checkIfFromChainTxExist(3, "0x00ca93f8738111a063d8ab7221f47c70a4cade0ca4a2829df494cd4b5e231bd6")
	assert.Equal(t, exist, true)

	exist = s.checkIfFromChainTxExist(3, "0x00ca93f8738111a063d8ab7221f47c70a4cade0ca4a2829df494cd4b5e231bd7")
	assert.Equal(t, exist, false)

}
