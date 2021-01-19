package service

import (
	"fmt"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	"github.com/polynetwork/zilliqa-relayer/config"
	"testing"
)

var p *PolySyncManager

func init() {
	zilSdk := provider.NewProvider("https://polynetworkcc3dcb2-5-api.dev.z7a.xyz")
	p = &PolySyncManager{
		zilSdk: zilSdk,
		cfg: &config.Config{ZilConfig: &config.ZILConfig{
			CrossChainManagerContract: "zil16vxy2u59sct5nupryxm3wfgteuhve9p0hp605f",
		}},
	}
}

func TestPolySyncManager_FindLatestHeight(t *testing.T) {
	fmt.Println(p.findLatestHeight())
}
