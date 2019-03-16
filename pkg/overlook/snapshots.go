package overlook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"
)

// GetBillingDataLocation returns where the billing directory will exist
func GetBillingDataLocation() string {
	billingDirName := filepath.Join(".", "billing")
	s, _ := filepath.Abs(billingDirName)
	return s
}

// GetBillingDataSortedFileNames returns a slice of billing data file names sorted by date
func GetBillingDataSortedFileNames() []string {
	var billingDataPath = GetBillingDataLocation()
	if _, err := os.Stat(billingDataPath); os.IsNotExist(err) {
		fmt.Println("Billing directory of", billingDataPath, " doesn't exist")
		panic(err)
	}
	// Get all files
	files, err := ioutil.ReadDir(billingDataPath)
	if err != nil {
		log.Fatal(err)
	}
	fileNames := make([]string, 0)
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	sort.Slice(fileNames, func(i, j int) bool { return fileNames[i] > fileNames[j] })

	absFileNames := make([]string, 0)
	for _, f := range fileNames {
		absFileNames = append(absFileNames, path.Join(billingDataPath, f))
	}
	return absFileNames
}

// ReadSnapshotInfo returns a BillingDailyEntry for given filename
func ReadSnapshotInfo(filename string) BillingDailyEntry {
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

// StoreBillingSnapshots will write billing snapshot data to billingDirPath
func StoreBillingSnapshots(regionInfo []RegionInfo, billingDirPath string) {
	//
	// TODO: Add ability to write to S3
	//
	now := time.Now()
	hour := now.Hour()
	ymd := now.Format("01-02-2006")
	var dailyEntry BillingDailyEntry

	if _, err := os.Stat(billingDirPath); os.IsNotExist(err) {
		err = os.MkdirAll(billingDirPath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	snapshotFilename := fmt.Sprintf("%s/%s.json", billingDirPath, ymd)
	dailyEntry = ReadSnapshotInfo(snapshotFilename)

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
