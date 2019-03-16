package cmd

import (
	"fmt"

	"github.com/jwmatthews/overlook/pkg/overlook"
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
	fmt.Println("Running report")
	usageFileNames := overlook.GetBillingDataSortedFileNames()
	fmt.Println(usageFileNames)

	for _, f := range usageFileNames {
		fmt.Println("Processing: ", f)
		dailyEntry := overlook.ReadSnapshotInfo(f)
		r := overlook.GetReport(dailyEntry)
		overlook.PrintReport(r)
	}
}
