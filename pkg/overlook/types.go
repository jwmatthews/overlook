package overlook

import (
	"github.com/aws/aws-sdk-go/service/ec2"
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
