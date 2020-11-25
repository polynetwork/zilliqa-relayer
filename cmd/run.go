package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/polynetwork/zilliqa-relayer/config"
	"github.com/polynetwork/zilliqa-relayer/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
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
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err.Error())
	}

}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run zilliqa relayer",
	Long:  `Run zilliqa relayer`,
	Run: func(cmd *cobra.Command, args []string) {

		cfg := &config.Config{
			ZilApiEndpoint:  viper.GetString("api"),
			ZilStartHeight:  viper.GetUint32("zil_start_height"),
			ZilScanInterval: viper.GetUint64("zil_scan_interval"),
		}

		syncService := service.NewSyncService(cfg)
		syncService.Run()
	},
}
