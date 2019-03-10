package overlook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func readSnapshotInfo(filename string) BillingDailyEntry {
	dailyEntry := make(BillingDailyEntry)

	if Exists(filename) {
		inFile, err := os.OpenFile(filename, os.O_RDONLY, 0660)
		if err != nil {
			panic(err)
		}
		defer CheckClose(inFile)

		// Read any existing data if file is not empty
		byteValue, err := ioutil.ReadAll(inFile)
		if err != nil {
			panic(err)
		}
		if len(byteValue) > 0 {
			err = json.Unmarshal(byteValue, &dailyEntry)
			if err != nil {
				panic(err)
			}
		}
	}
	return dailyEntry
}

func writeSnapshotInfo(filename string, dailyEntry BillingDailyEntry) (err error) {
	//
	// Want to delete any existing data if present
	//
	if Exists(filename) {
		err = os.Truncate(filename, 0)
		if err != nil {
			panic(err)
		}
	}

	outFile, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		panic(err)
	}
	defer CheckClose(outFile)

	var bsJSON []byte
	bsJSON, err = json.Marshal(dailyEntry)
	if err != nil {
		panic(err)
	}
	_, err = outFile.WriteString(string(bsJSON))
	if err != nil {
		panic(err)
	}
	err = outFile.Sync()
	return err
}

// StoreBillingSnapshots will write billing snapshot data to S3
func StoreBillingSnapshots(regionInfo []RegionInfo, dirName string) {
	//
	// TODO: Add ability to change directory where snapshots are stored.
	//
	now := time.Now()
	hour := now.Second()
	ymd := now.Format("01-02-2006")
	var dailyEntry BillingDailyEntry

	billingDirName := filepath.Join(".", dirName)
	if _, err := os.Stat(billingDirName); os.IsNotExist(err) {
		err = os.MkdirAll(billingDirName, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	snapshotFilename := fmt.Sprintf("%s/%s.json", dirName, ymd)
	dailyEntry = readSnapshotInfo(snapshotFilename)

	var hourlyEntry BillingHourEntry
	var regionEntry BillingRegionEntry
	var ok bool

	hourlyEntry, ok = dailyEntry[ymd]
	if !ok {
		hourlyEntry = make(BillingHourEntry)
	}

	regionEntry, ok = hourlyEntry[hour]
	if !ok {
		regionEntry = make(BillingRegionEntry)
	}

	for _, r := range regionInfo {
		var instancesEntry BillingInstancesEntry
		instancesEntry, ok = regionEntry[r.RegionName]
		if !ok {
			instancesEntry = make(BillingInstancesEntry)
		}
		for _, bSnap := range r.BillingSnapshots {
			instancesEntry[bSnap.ID] = bSnap
		}
		regionEntry[r.RegionName] = instancesEntry
	}

	hourlyEntry[hour] = regionEntry
	dailyEntry[ymd] = hourlyEntry

	err := writeSnapshotInfo(snapshotFilename, dailyEntry)
	if err != nil {
		panic(err)
	}
}
