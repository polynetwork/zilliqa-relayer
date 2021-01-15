package cmd

import (
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")

	runCmd.Flags().String("api", "https://api.zilliqa.com", "zilliqa api endpoint")
	if err := viper.BindPFlag("api", runCmd.Flags().Lookup("api")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}
	runCmd.Flags().String("zil_start_height", "1", "the from block number will be syncing to poly network")
	if err := viper.BindPFlag("api", runCmd.Flags().Lookup("zil_start_height")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}

	runCmd.Flags().String("zil_scan_interval", "2", "the interval scanning zilliqa block")
	if err := viper.BindPFlag("zil_scan_interval", runCmd.Flags().Lookup("zil_scan_interval")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}

	RootCmd.AddCommand(runCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	} else {
		log.Error(err.Error())
	}

}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run zilliqa relayer",
	Long:  `Run zilliqa relayer`,
	Run: func(cmd *cobra.Command, args []string) {
		zilConfigMap := viper.GetStringMap("zil_config")
		api := zilConfigMap["zil_api"].(string)
		zilConfig := &config.ZILConfig{ZilApiEndpoint: api,
			ZilStartHeight:            uint32(zilConfigMap["zil_start_height"].(int)),
			ZilMonitorInterval:        uint32(zilConfigMap["zil_monitor_interval"].(int)),
			SideChainId:               uint64(zilConfigMap["side_chain_id"].(int)),
			CrossChainManagerContract: zilConfigMap["corss_chain_manager_address"].(string),
		}

		cfg := &config.Config{
			ZilConfig: zilConfig,
		}

		log.Infof("config file: %+v\n", cfg)

		syncService := service.NewSyncService(cfg)
		syncService.Run()
	},
}
