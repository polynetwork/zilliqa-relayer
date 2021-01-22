package cmd

import (
	"encoding/json"
	"github.com/Zilliqa/gozilliqa-sdk/provider"
	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/db"
	"github.com/polynetwork/zilliqa-relayer/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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

func setUpPoly(poly *poly_go_sdk.PolySdk, RpcAddr string) error {
	poly.NewRpcClient().SetAddress(RpcAddr)
	hdr, err := poly.GetHeaderByHeight(0)
	if err != nil {
		return err
	}
	poly.SetChainId(hdr.ChainID)
	return nil
}

func checkIfExist(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil && !os.IsExist(err) {
		return false
	}
	return true
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run zilliqa relayer",
	Long:  `Run zilliqa relayer`,
	Run: func(cmd *cobra.Command, args []string) {
		zilConfigMap := viper.GetStringMap("zil_config")
		zilConfig := &config.ZILConfig{
			ZilApiEndpoint:                 zilConfigMap["zil_api"].(string),
			ZilChainId:                     zilConfigMap["zil_chain_id"].(int),
			ZilMessageVersion:              zilConfigMap["zil_message_version"].(int),
			ZilStartHeight:                 uint32(zilConfigMap["zil_start_height"].(int)),
			ZilMonitorInterval:             uint32(zilConfigMap["zil_monitor_interval"].(int)),
			SideChainId:                    uint64(zilConfigMap["side_chain_id"].(int)),
			CrossChainManagerContract:      zilConfigMap["corss_chain_manager_address"].(string),
			CrossChainManagerProxyContract: zilConfigMap["cross_chain_manager_proxy_address"].(string),
			KeyStorePath:                   zilConfigMap["key_store_path"].(string),
			KeyStorePwdSet:                 zilConfigMap["key_store_pwd_set"].(map[string]interface{}),
		}

		polyConfigMap := viper.GetStringMap("poly_config")
		polyConfig := &config.POLYConfig{
			PolyWalletFile:          polyConfigMap["poly_wallet_file"].(string),
			PolyWalletPassword:      polyConfigMap["poly_wallet_pwd"].(string),
			PolyMonitorInterval:     uint32(polyConfigMap["poly_monitor_interval"].(int)),
			EntranceContractAddress: polyConfigMap["entrance_contract_address"].(string),
			RestUrl:                 polyConfigMap["rest_url"].(string),
		}

		cfg := &config.Config{
			ZilConfig:  zilConfig,
			PolyConfig: polyConfig,
		}

		cfgStr, _ := json.Marshal(cfg)
		log.Infof("config file: %s\n", cfgStr)

		zilSdk := provider.NewProvider(cfg.ZilConfig.ZilApiEndpoint)
		polySdk := poly_go_sdk.NewPolySdk()
		err1 := setUpPoly(polySdk, cfg.PolyConfig.RestUrl)
		if err1 != nil {
			log.Errorf("init poly sdk error: %s\n", err1.Error())
			return
		}

		if !checkIfExist(cfg.Path) {
			os.Mkdir(cfg.Path, os.ModePerm)
		}
		boltDB, err2 := db.NewBoltDB(cfg.Path)
		if err2 != nil {
			log.Errorf("cannot init bolt db: %s\n", err2.Error())
			return
		}

		zilliqaManager, err := service.NewZilliqaSyncManager(cfg, zilSdk, polySdk, boltDB)
		if err != nil {
			log.Errorf("init zilliqamanger error: %s\n", err.Error())
			return
		}
		polyManager, err1 := service.NewPolySyncManager(cfg, zilSdk, polySdk, boltDB, cfg.ZilConfig.CrossChainManagerContract, cfg.ZilConfig.CrossChainManagerProxyContract)
		if err1 != nil {
			log.Errorf("init polymanager error: %s\n", err1.Error())
			return
		}

		zilliqaManager.Run()
		polyManager.Run()

		service.WaitToExit()

	},
}
