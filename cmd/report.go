package cmd

import (
	"fmt"
	"github.com/jwmatthews/overlook/pkg/overlook"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
	"path"
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
	// Find where billing info is
	var billingDataPath = overlook.GetBillingDataLocation()
	if _, err := os.Stat(billingDataPath); os.IsNotExist(err) {
		fmt.Println("Billing directory of", billingDataPath, " doesn't exist")
		panic(err)
	}
	fmt.Println("Reading billing info stored in", billingDataPath)
	// Get all files
	files, err := ioutil.ReadDir(billingDataPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fName := path.Join(billingDataPath, f.Name())
		fmt.Println("Processing: ", fName)
		dailyEntry := overlook.ReadSnapshotInfo(fName)
		overlook.CalculateReport(dailyEntry)
	}

	// Parse each day

}
