package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "zilliqa-relayer",
	Short: "To run zilliqa relayer for poly network",
	Long:  `To run zilliqa relayer for poly network`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
