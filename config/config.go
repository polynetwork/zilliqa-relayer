package config

type Config struct {
	ZilConfig *ZILConfig
	Path      string
}

type ZILConfig struct {
	ZilApiEndpoint            string
	ZilMonitorInterval        uint32
	ZilStartHeight            uint32
	SideChainId               uint64
	CrossChainManagerContract string
}
