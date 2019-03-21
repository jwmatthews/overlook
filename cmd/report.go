package cmd

import (
	"github.com/jwmatthews/overlook/pkg/overlook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ReportCommand cobra command to invoke Report
var ReportCommand = &cobra.Command{
	Use:   "report",
	Short: "Report parses usage data",
	Long:  `Report parses usage data`,
	Run: func(cmd *cobra.Command, args []string) {
		Report()
	},
}

func Report() {
	log.Infoln("Running report")
	usageFileNames := overlook.GetBillingDataSortedFileNames()
	log.Infoln(usageFileNames)

	for _, f := range usageFileNames {
		log.Infoln("Processing: ", f)
		dailyEntry := overlook.ReadSnapshotInfo(f)
		r := overlook.GetReport(dailyEntry)
		overlook.PrintReport(r)
	}
}
