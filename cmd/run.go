package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
)


func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().String("api", "https://api.zilliqa.com", "zilliqa api endpoint")
	if err := viper.BindPFlag("api", runCmd.Flags().Lookup("api")); err != nil {
		log.Fatal("Unable to bind flag:", err)
	}}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run zilliqa relayer",
	Long:  `Run zilliqa relayer`,
	Run: func(cmd *cobra.Command, args []string) {
		msg := viper.GetString("api")
		fmt.Println(msg)
	},
}
