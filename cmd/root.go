package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "overlook",
	Short: "Overlook samples EC2 usage and creates reports of usage and costs",
	Long: `Overlook provides a means of sampling EC2 usage for scenarios where users lack rights to see billing information from AWS.
A long running process will periodically inspect all regions and capture usage at that specific point in time.
This sampling of usage is then used to create reports on usage and estimate of costs over a given period.`,
	Run: func(commmand *cobra.Command, args []string) {
	},
}

var cmdWatch = &cobra.Command{
	Use:   "watch [# times] [string to echo]",
	Short: "Watches ec2 usage",
	Long:  `Watches ec2 usage, sampling at a given interval and recording usage info.`,
	Run: func(cmd *cobra.Command, args []string) {
		Watch()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(cmdWatch)
}

func initConfig() {
	//
}
