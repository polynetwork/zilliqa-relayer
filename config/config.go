package config

const (
	OntUsefulBlockNum = 1
)

type Config struct {
	ZilConfig       *ZILConfig
	PolyConfig      *POLYConfig
	Path            string
	TargetContracts []map[string]map[string][]uint64
}

type ZILConfig struct {
	ZilApiEndpoint            string
	ZilMonitorInterval        uint32
	ZilStartHeight            uint32
	SideChainId               uint64
	CrossChainManagerContract string
}

type POLYConfig struct {
	PolyWalletFile          string
	PolyWalletPassword      string
	PolyMonitorInterval     uint32
	EntranceContractAddress string
	RestUrl                 string
}
