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

// BillingHourEntry higher level wrapper for snapshot data
type BillingHourEntry map[int]map[string]map[string]BillingSnapshot

// BillingRegionEntry gather snapshots by region
type BillingRegionEntry struct {
	Region    string
	Snapshots []BillingSnapshot
}

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
