package cost

import "fmt"

var CostPerHour map[string]float64

func init() {
	//https://aws.amazon.com/ec2/pricing/on-demand/
	CostPerHour = make(map[string]float64)
	CostPerHour["m4.large"] = 0.10
	CostPerHour["t2.micro"] = 0.0116
	CostPerHour["m4.xlarge"] = 0.20
	CostPerHour["c4.4xlarge"] = 0.796
	CostPerHour["t2.xlarge"] = 0.1856
	CostPerHour["t2.2xlarge"] = 0.3712
}

// GetCostPerHour returns cost per hour based in instanceType
func GetCostPerHour(instanceType string) (float64, error) {
	cost := CostPerHour[instanceType]
	if cost == 0 {
		return 0.0, fmt.Errorf("Unknown instance type: %s", instanceType)
	}
	return cost, nil
}
