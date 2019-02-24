package overlook

import "fmt"

var costPerHour map[string]float64

func init() {
	//https://aws.amazon.com/ec2/pricing/on-demand/
	costPerHour = make(map[string]float64)
	costPerHour["m4.large"] = 0.10
	costPerHour["t2.micro"] = 0.0116
	costPerHour["m4.xlarge"] = 0.20
	costPerHour["c4.4xlarge"] = 0.796
	costPerHour["t2.xlarge"] = 0.1856
	costPerHour["t2.2xlarge"] = 0.3712
}

// GetCostPerHour returns cost per hour based in instanceType
func GetCostPerHour(instanceType string) (float64, error) {
	cost := costPerHour[instanceType]
	if cost == 0 {
		return 0.0, fmt.Errorf("unknown instance type: %s", instanceType)
	}
	return cost, nil
}
