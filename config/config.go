package config

type Config struct {
	Path string

	ZilMonitorInterval uint32
	ZilApiEndpoint     string
	ZilStartHeight     uint32
}
