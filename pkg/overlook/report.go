package overlook

import (
	"errors"
	log "github.com/sirupsen/logrus"
)

// NewReportDaily returns a new ReportDaily
func NewReportDaily() ReportDaily {
	var report = ReportDaily{}
	report.Regions = make(map[string]ReportByRegion)
	return report
}

func NewReportByRegion() ReportByRegion {
	var report ReportByRegion
	report.InstanceTypes = make(map[string]ReportInstanceType)
	return report
}

func PrintReport(r ReportDaily) {
	log.Infoln(r.FormatByCost())
}

// GetReport returns a summary of usage and costs for as given BillingDailyEntry
func GetReport(dailyEntry BillingDailyEntry) ReportDaily {
	//
	// We will walk through a report of usage which is focused on days usage of ec2..
	// The usage report is organized by hour, showing what instances we've seen in that hour
	// We want to consolidate the info by region and instanceType
	//
	for date, dayEntry := range dailyEntry {
		// For each day we create a new report, we structured the JSON to only contain 1 day in an entry
		var report = NewReportDaily()
		report.Date = date
		for _, hourEntry := range dayEntry {
			for region, regionEntry := range hourEntry {
				var ok bool
				var reportByRegion ReportByRegion
				reportByRegion, ok = report.Regions[region]
				if !ok {
					reportByRegion = NewReportByRegion()
					reportByRegion.Region = region
				}

				for _, instanceEntry := range regionEntry {
					instType := instanceEntry.InstanceType
					reportInst, ok := reportByRegion.InstanceTypes[instType]
					if !ok {
						reportInst = ReportInstanceType{}
						reportInst.InstanceType = instType
						reportInst.Hours = 0
						reportInst.UniqueInstances = make(map[string]bool)
					}
					instanceCost, err := GetCostPerHour(reportInst.InstanceType)
					if err != nil {
						panic(err)
					}
					reportInst.Hours = reportInst.Hours + 1
					reportInst.Cost = float64(reportInst.Hours) * instanceCost
					reportInst.UniqueInstances[instanceEntry.ID] = true
					reportByRegion.InstanceTypes[instType] = reportInst
				}
				report.Regions[region] = reportByRegion
			}
		}
		// Calculate cost per region
		for region, reportByRegion := range report.Regions {
			reportByRegion.Cost = 0
			for _, reportInstType := range reportByRegion.InstanceTypes {
				reportByRegion.Cost = reportByRegion.Cost + reportInstType.Cost
			}
			report.Regions[region] = reportByRegion
		}
		// Calculate total cost
		report.Cost = 0
		for _, reportByRegion := range report.Regions {
			for _, reportInstType := range reportByRegion.InstanceTypes {
				report.Cost = report.Cost + reportInstType.Cost
			}
		}
		return report
	}
	// This should never happen
	panic(errors.New("We should never reach here, problem parsing date in billing data"))
	return ReportDaily{}
}

// PrintCalculateReport returns a summary of usage and costs for as given BillingDailyEntry
func PrintCalculateReport(dailyEntry BillingDailyEntry) {
	for date, dayEntry := range dailyEntry {
		log.Infoln(date)
		for hour, hourEntry := range dayEntry {
			log.Infoln("\t", hour)
			for region, regionEntry := range hourEntry {
				log.Infoln("\t\t", region)
				for instanceID, instanceEntry := range regionEntry {
					log.Infoln("\t\t\t", instanceID, ":", "up ", instanceEntry.HoursUp)
				}
			}
		}
	}
}
