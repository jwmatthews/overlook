package overlook

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

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
	cost, err := GetCostPerHour(inst.InstanceType)
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

// FormBillingSnapshots is used to help us form data for estimating costs over time
func FormBillingSnapshots(instances []InstanceInfo) []BillingSnapshot {
	var billSnaps = make([]BillingSnapshot, 0)
	for _, inst := range instances {
		b := BillingSnapshot{}
		cost, err := GetCostPerHour(inst.InstanceType)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		b.ID = *inst.Instance.InstanceId
		b.CostPerHour = cost
		b.CurrentCost = CalculateCostPer(inst)
		b.InstanceType = inst.InstanceType
		b.HoursUp = inst.HoursUp
		b.Tags = inst.Tags
		b.State = inst.State
		b.AvailabilityZone = inst.AvailabilityZone
		b.Region = inst.Region
		b.Arn = inst.Arn
		billSnaps = append(billSnaps, b)
	}
	return billSnaps
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
			info.AvailabilityZone = *inst.Placement.AvailabilityZone
			info.Region = *svc.Config.Region
			info.State = *inst.State.Name
			info.Tags = tags
			info.InstanceType = *inst.InstanceType
			if inst.IamInstanceProfile != nil {
				if inst.IamInstanceProfile.Arn != nil {
					info.Arn = *inst.IamInstanceProfile.Arn
				}
			}
			instances = append(instances, info)
		}
	}
	//
	// Sort by HoursUp
	//
	sort.Slice(instances, func(i, j int) bool { return instances[i].HoursUp > instances[j].HoursUp })
	return instances, nil
}
