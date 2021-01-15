package config

type Config struct {
	ZilConfig  *ZILConfig
	PolyConfig *POLYConfig
	Path       string
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
	EntranceContractAddress string
	RestUrl                 string
}
