package cmd

import (
	"github.com/jwmatthews/overlook/pkg/overlook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ReportCommand cobra command to invoke Report
var SpreadSheetCommand = &cobra.Command{
	Use:   "spreadsheet",
	Short: "Store usage data in Spreadsheet",
	Long:  `Store usage data in Spreadsheet`,
	Run: func(cmd *cobra.Command, args []string) {
		SpreadSheet()
	},
}

func SpreadSheet() {
	log.Infoln("Running spreadsheet")
	overlook.RunSpreadsheet()
	/*
		usageFileNames := overlook.GetBillingDataSortedFileNames()
		log.Infoln(usageFileNames)


		for _, f := range usageFileNames {
			log.Infoln("Processing: ", f)
			dailyEntry := overlook.ReadSnapshotInfo(f)
			r := overlook.GetReport(dailyEntry)
			overlook.PrintReport(r)
		}
	*/

}
