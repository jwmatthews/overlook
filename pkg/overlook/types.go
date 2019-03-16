package overlook

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sort"
)

// RegionInfo captures instance info across a region
type RegionInfo struct {
	Instances        []InstanceInfo
	RegionName       string
	Cost             float64
	TypeSummary      map[string]InstanceTypeSummary
	BillingSnapshots []BillingSnapshot
}

// InstanceInfo captures info we care most about
type InstanceInfo struct {
	Instance         *ec2.Instance
	HoursUp          float64
	Cost             float64
	State            string
	Tags             string
	InstanceType     string
	AvailabilityZone string
	Arn              string
	Region           string
}

// InstanceTypeSummary tracks aggregate info about a specific instance type
type InstanceTypeSummary struct {
	InstanceType      string
	NumberOfInstances int
	TotalHours        float64
	Cost              float64
}

// The layout for the billing snapshot is
// Each day is a new json file with structure of
// {"$DATE":
//  {"$HOUR":
//   {"$REGION:
//     { "$INSTANCE_ID_1":  {"$BillingSnapshot"}
//     { "$INSTANCE_ID_1":  {"$BillingSnapshot"}
//  }}}

// BillingDailyEntry, for a given day has all of the billing info organized by hour
type BillingDailyEntry map[string]BillingHourEntry

// BillingHourEntry, for a given hour, has all of the billing info organized by region
type BillingHourEntry map[int]BillingRegionEntry

// BillingRegionEntry, for a given region has all of the billing info organized by instance-id
type BillingRegionEntry map[string]BillingInstancesEntry

// BillingInstancesEntry, for a given instance-id has the billing information
type BillingInstancesEntry map[string]BillingSnapshot

// BillingSnapshot is used to capture time series data of usage
type BillingSnapshot struct {
	ID               string
	InstanceType     string
	Region           string
	AvailabilityZone string
	State            string
	Tags             string
	HoursUp          float64
	CostPerHour      float64
	CurrentCost      float64
	Arn              string
}

type ReportDaily struct {
	Regions map[string]ReportByRegion
	Cost    float64
	Date    string
}

func (r ReportDaily) String() string {
	s := fmt.Sprintf("%s, Cost:%.2f", r.Date, r.Cost)
	for region, reportByRegion := range r.Regions {
		s = s + fmt.Sprintf("\n\t%s, Cost: %.2f", region, reportByRegion.Cost)
		for instanceType, reportInstanceType := range reportByRegion.InstanceTypes {
			s = s + fmt.Sprintf("\n\t\t%s: Cost: %.2f, Hours:%d", instanceType, reportInstanceType.Cost, reportInstanceType.Hours)
		}
	}
	return s
}

func (r ReportDaily) FormatByCost() string {
	s := fmt.Sprintf("%s, Cost:%.2f", r.Date, r.Cost)
	var regionInfo = make([]ReportByRegion, 0)
	// Filter and remove regions with no activity
	for _, reportByRegion := range r.Regions {
		if reportByRegion.Cost > 0 {
			regionInfo = append(regionInfo, reportByRegion)
		}
	}
	// Sort by cost
	sort.Slice(regionInfo, func(i, j int) bool { return regionInfo[i].Cost > regionInfo[j].Cost })

	for _, r := range regionInfo {
		s = s + fmt.Sprintf("\n\t%s, Cost: %.2f", r.Region, r.Cost)
		for instanceType, reportInstanceType := range r.InstanceTypes {
			s = s + fmt.Sprintf("\n\t\t%s: Cost: %.2f, Hours:%d, NumberUniqueInstances:%d",
				instanceType, reportInstanceType.Cost, reportInstanceType.Hours, len(reportInstanceType.UniqueInstances))
		}
	}
	return s
}

type ReportByRegion struct {
	InstanceTypes map[string]ReportInstanceType
	Cost          float64
	Region        string
}

func (r ReportByRegion) String() string {
	var s string
	for region, reportByInstanceType := range r.InstanceTypes {
		s = s + fmt.Sprintf("\n\t%s: %s", region, reportByInstanceType)
	}
	return s
}

type ReportInstanceType struct {
	InstanceType    string
	Hours           int
	Cost            float64
	UniqueInstances map[string]bool
}

func (r ReportInstanceType) String() string {
	return fmt.Sprintf("%s: Cost:%.2f, Hours:%d", r.InstanceType, r.Cost, r.Hours)
}
