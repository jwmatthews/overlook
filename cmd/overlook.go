package main

import (
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jwmatthews/overlook/pkg/overlook"
)

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
	var runningTotal float64
	var regionInfo = make([]overlook.RegionInfo, 0)
	for rInfo := range c {
		regionInfo = append(regionInfo, rInfo)
		runningTotal += rInfo.Cost
	}
	overlook.DisplayRegionInfo(regionInfo)
	overlook.StoreBillingSnapshots(regionInfo, "billing")
	return runningTotal, regionInfo
}

func processRegion(sess client.ConfigProvider, region string) overlook.RegionInfo {
	fmt.Println("Processing region: ", region)
	var rInfo overlook.RegionInfo
	rInfo.RegionName = region

	svc := ec2.New(sess, aws.NewConfig().WithRegion(rInfo.RegionName))
	instances, err := overlook.GetInstances(svc)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	rInfo.Instances = instances
	rInfo.Cost = overlook.CalculateCost(instances)
	rInfo.TypeSummary = overlook.CreateInstanceTypeSummary(instances)
	rInfo.BillingSnapshots = overlook.FormBillingSnapshots(instances)
	fmt.Println("Completed processing region: ", region)
	return rInfo
}

//
// Will walk through all regions and gather a report of
//	- Instances
//			By type:  Number of instances with uptime of each, total hours up
//	- Volumes: TODO
func main() {
	var region string
	flag.StringVar(&region, "r", "", "Specify a single region, by default will assume all regions")
	flag.Parse()

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

	fmt.Println("Working with ", len(regions), "regions: ", regions)

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
	fmt.Println("RunningTotal: ", formattedTotal)
}
