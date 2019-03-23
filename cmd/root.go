package cmd

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

const logFileName = "overlook.log"

var rootCmd = &cobra.Command{
	Use:   "overlook",
	Short: "Overlook samples EC2 usage and creates reports of usage and costs",
	Long: `Overlook provides a means of sampling EC2 usage for scenarios where users lack rights to see billing information from AWS.
A long running process will periodically inspect all regions and capture usage at that specific point in time.
This sampling of usage is then used to create reports on usage and estimate of costs over a given period.`,
	Run: func(commmand *cobra.Command, args []string) {
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		fmt.Println("Quitting...unable to create logging file: " + logFileName)
		os.Exit(1)
	}
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	log.SetReportCaller(true)
	log.SetOutput(logFile)
	//mw := io.MultiWriter(os.Stdout, logFile)
	//log.SetOutput(mw)

	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(WatchCommand)
	rootCmd.AddCommand(ReportCommand)
	rootCmd.AddCommand(EmailCommand)
	rootCmd.AddCommand(SpreadSheetCommand)

	log.Infoln("Starting")
}

func initConfig() {
	//
}
