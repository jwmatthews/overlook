package overlook

import (
	"github.com/aws/aws-sdk-go/service/ec2"
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
