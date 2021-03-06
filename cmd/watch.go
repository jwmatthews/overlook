package cmd

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jwmatthews/overlook/pkg/overlook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

// WatchCommand cobra command to invoke Watch
var WatchCommand = &cobra.Command{
	Use:   "watch",
	Short: "Watches ec2 usage",
	Long:  `Watches ec2 usage, sampling at a given interval and recording usage info.`,
	Run: func(cmd *cobra.Command, args []string) {
		Watch()
	},
}

var region string

func init() {
	WatchCommand.Flags().StringVarP(&region, "region", "r", "", "Specify a single region, by default will assume all regions")
}

// GetRegions returns a slice of all region strings
func GetRegions(sess client.ConfigProvider) []string {
	svc := ec2.New(sess)
	resultRegions, err := svc.DescribeRegions(nil)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	var regions = make([]string, 0)
	for _, r := range resultRegions.Regions {
		regions = append(regions, *r.RegionName)
	}
	return regions
}

func aggregateAllInfo(c <-chan overlook.RegionInfo) (float64, []overlook.RegionInfo) {
	var billingDir = overlook.GetBillingDataLocation()
	var runningTotal float64
	var regionInfo = make([]overlook.RegionInfo, 0)
	for rInfo := range c {
		regionInfo = append(regionInfo, rInfo)
		runningTotal += rInfo.Cost
	}
	overlook.DisplayRegionInfo(regionInfo)
	overlook.StoreBillingSnapshots(regionInfo, billingDir)
	return runningTotal, regionInfo
}

func processRegion(sess client.ConfigProvider, region string) overlook.RegionInfo {
	fmt.Println("Processing region: ", region)
	var rInfo overlook.RegionInfo
	rInfo.RegionName = region

	svc := ec2.New(sess, aws.NewConfig().WithRegion(rInfo.RegionName))
	instances, err := overlook.GetInstances(svc)
	if err != nil {
		log.Fatalln("Error", err)
		os.Exit(1)
	}
	rInfo.Instances = instances
	rInfo.Cost, err = overlook.CalculateCost(instances)
	if err != nil {
		log.Errorln("Unable to calculate costs for all instances")
		log.Errorln(err)
	}
	rInfo.TypeSummary = overlook.CreateInstanceTypeSummary(instances)
	rInfo.BillingSnapshots = overlook.FormBillingSnapshots(instances)
	log.Infoln("Completed processing region: ", region)
	return rInfo
}

//
// Will walk through all regions and gather a report of
//	- Instances
//			By type:  Number of instances with uptime of each, total hours up
//	- Volumes: TODO
func Watch() {
	log.Infoln("Watch invoked")
	// Load session from shared config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	regions := make([]string, 0)
	if region != "" {
		regions = append(regions, region)
	} else {
		regions = GetRegions(sess)
	}

	log.Infoln("Working with ", len(regions), "regions: ", regions)

	var runningTotal float64
	var consumerGroup sync.WaitGroup
	var producerGroup sync.WaitGroup
	var regionInfoChannel = make(chan overlook.RegionInfo, 3)

	// Consumer:  Will aggregate all the info
	consumerGroup.Add(1)
	go func() {
		defer consumerGroup.Done()
		runningTotal, _ = aggregateAllInfo(regionInfoChannel)
	}()

	// Producer: Create a goroutine per region to produce info
	for _, r := range regions {
		producerGroup.Add(1)
		go func(sess *session.Session, reg string, c chan<- overlook.RegionInfo) {
			defer producerGroup.Done()
			rInfo := processRegion(sess, reg)
			c <- rInfo
		}(sess, r, regionInfoChannel)
	}
	producerGroup.Wait()
	close(regionInfoChannel)
	//
	// Now we wait for consumer to complete
	//
	consumerGroup.Wait()

	formattedTotal := fmt.Sprintf("%.2f", runningTotal)
	log.Infoln("RunningTotal: ", formattedTotal)
}
