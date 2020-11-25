package config

type Config struct {
	Path string

	ZilScanInterval uint64
	ZilApiEndpoint  string
	ZilStartHeight  uint32
}
