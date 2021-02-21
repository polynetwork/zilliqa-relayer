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
	ZilApiEndpoint                 string
	ZilChainId                     int
	ZilMessageVersion              int
	ZilMonitorInterval             uint32
	ZilHeadersPerBatch             uint32
	ZilStartHeight                 uint32
	ZilForceHeight                 uint64
	SideChainId                    uint64
	CrossChainManagerContract      string
	CrossChainManagerProxyContract string
	KeyStorePath                   string
	KeyStorePwdSet                 map[string]interface{}
}

type POLYConfig struct {
	PolyWalletFile          string
	PolyWalletPassword      string
	PolyStartHeight         uint32
	PolyMonitorInterval     uint32
	EntranceContractAddress string
	RestUrl                 string
}
