package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jwmatthews/overlook/pkg/cost"
)

// RegionInfo captures instance info across a region
type RegionInfo struct {
	Instances   []InstanceInfo
	RegionName  string
	Cost        float64
	TypeSummary map[string]InstanceTypeSummary
}

// InstanceInfo captures info we care most about
type InstanceInfo struct {
	Instance     *ec2.Instance
	HoursUp      float64
	Cost         float64
	Region       string
	State        string
	Tags         string
	InstanceType string
}

// InstanceTypeSummary tracks aggregate info about a specific instance type
type InstanceTypeSummary struct {
	InstanceType      string
	NumberOfInstances int
	TotalHours        float64
	Cost              float64
}

func hoursSince(fromTime time.Time) float64 {
	return time.Since(fromTime).Hours()
}

// DisplayRegionInfo prints info to stdout
func DisplayRegionInfo(regionInfo []RegionInfo) {
	for _, r := range regionInfo {
		if r.TypeSummary != nil && len(r.TypeSummary) > 0 {
			fmt.Println(r.RegionName)
		}
		for _, sum := range r.TypeSummary {
			fmt.Println("\t", sum.InstanceType)
			fmt.Println("\t\t Number of Instances:", sum.NumberOfInstances)
			fmt.Printf("\t\t TotalHours: %.2f\n", sum.TotalHours)
			fmt.Printf("\t\t Cost of Current Running: %.2f\n", sum.Cost)
		}
	}
}

// CalculateCost calculates cost of current instances
func CalculateCost(instances []InstanceInfo) float64 {
	runningTotal := 0.0
	for _, inst := range instances {
		rawEstimatedCost := CalculateCostPer(inst)
		inst.Cost = rawEstimatedCost
		runningTotal += rawEstimatedCost
	}
	return runningTotal
}

// CalculateCostPer cost of a single instance
func CalculateCostPer(inst InstanceInfo) float64 {
	cost, err := cost.GetCostPerHour(inst.InstanceType)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rawEstimatedCost := cost * inst.HoursUp
	return rawEstimatedCost
}

// CreateInstanceTypeSummary creates summary info on instance types
func CreateInstanceTypeSummary(instances []InstanceInfo) map[string]InstanceTypeSummary {
	var summary = make(map[string]InstanceTypeSummary)
	for _, inst := range instances {
		instSumm := summary[inst.InstanceType]
		instSumm.InstanceType = inst.InstanceType
		instSumm.NumberOfInstances++
		instSumm.TotalHours += inst.HoursUp
		instSumm.Cost += CalculateCostPer(inst)
		summary[inst.InstanceType] = instSumm
	}
	return summary
}

//
// GetInstances a list of instances in the region ordered by launchTime
//
func GetInstances(svc *ec2.EC2) ([]InstanceInfo, error) {
	instances := make([]InstanceInfo, 0)
	resultInstances, err := svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}

	for _, r := range resultInstances.Reservations {
		for _, inst := range r.Instances {
			if *inst.State.Name != "running" {
				//fmt.Println("Skipping since state is: ", *inst.State.Name)
				continue
			}
			var info InstanceInfo
			var tags string
			for _, t := range inst.Tags {
				tags += fmt.Sprintf("%s:%s ", *t.Key, *t.Value)
			}
			info.Instance = inst
			info.HoursUp = hoursSince(*inst.LaunchTime)
			info.Region = *inst.Placement.AvailabilityZone
			info.State = *inst.State.Name
			info.Tags = tags
			info.InstanceType = *inst.InstanceType
			instances = append(instances, info)

			//if inst.IamInstanceProfile != nil {
			//	if inst.IamInstanceProfile.Arn != nil {
			//		fmt.Println("\t\tIamInstanceProfile.Arn:", *inst.IamInstanceProfile.Arn)
			//	} else {
			//		fmt.Println("\t\tIamInstanceProfile.Arn missing")
			//	}
			//} else {
			//	fmt.Println("\t\tIamInstanceProfile: missing")
			//}

		}
	}
	//
	// Sort by HoursUp
	//
	sort.Slice(instances, func(i, j int) bool { return instances[i].HoursUp > instances[j].HoursUp })
	return instances, nil
}

//
// Will walk through all regions and gather a report of
//	- Instances
//			By type:  Number of instances with uptime of each, total hours up
//	- Volumes

func main() {
	//var region string
	//flag.StringVar(&region, "r", "", "Specify a single region, by default will assume all regions")
	//flag.Parse()

	// Load session from shared config
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Find all regions
	svc := ec2.New(sess)
	resultRegions, err := svc.DescribeRegions(nil)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println(len(resultRegions.Regions), "regions found")
	regionInfo := make([]RegionInfo, 0)
	runningTotal := 0.0
	for _, r := range resultRegions.Regions {
		var rInfo RegionInfo
		rInfo.RegionName = *r.RegionName

		svc := ec2.New(sess, aws.NewConfig().WithRegion(rInfo.RegionName))
		instances, err := GetInstances(svc)
		if err != nil {
			fmt.Println("Error", err)
			return
		}
		rInfo.Instances = instances
		rInfo.Cost = CalculateCost(instances)
		rInfo.TypeSummary = CreateInstanceTypeSummary(instances)

		regionInfo = append(regionInfo, rInfo)
		runningTotal += rInfo.Cost
	}

	DisplayRegionInfo(regionInfo)

	formattedTotal := fmt.Sprintf("%.2f", runningTotal)
	fmt.Println("RunningTotal: ", formattedTotal)

}
