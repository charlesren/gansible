package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

//Result struct store execute result of gansible command
type Result struct {
	StartTime        time.Time
	SuccessHosts     []interface{}
	FailedHosts      []interface{}
	UnreachableHosts []interface{}
	SkipedHosts      []interface{}
	TotalHosts       []interface{}
	EndTime          time.Time
	CostTime         time.Duration
}

// AppendToFile will print any string of text to a file safely by
// checking for errors and syncing at the end.
func AppendToFile(file string, str string) error {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write([]byte(str)); err != nil {
		log.Fatal(err)
	}
	return f.Sync()
}

//EndInfo gengrate summary of gansible result
func EndInfo(result Result, startTime time.Time) string {
	//result.StartTime = startTime.Format("2006-01-02 15:04:05")
	result.EndTime = time.Now()
	result.CostTime = result.EndTime.Sub(startTime)
	endTimeStr := result.EndTime.Format("2006-01-02 15:04:05")
	costTimeStr := result.CostTime.String()
	totalHostsNum := len(result.FailedHosts) + len(result.SuccessHosts) + len(result.UnreachableHosts) + len(result.SkipedHosts)
	summary := fmt.Sprintf("\nEnd Time: %s\nCost Time: %s\nTotal(%d) : Success=%d    Failed=%d    Unreachable=%d    Skipped=%d", endTimeStr, costTimeStr, totalHostsNum, len(result.SuccessHosts), len(result.FailedHosts), len(result.UnreachableHosts), len(result.SkipedHosts))
	return summary
}
