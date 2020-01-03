package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

//Result struct store execute result of gansible command
type Result struct {
	StartTime    string
	SuccessHosts []interface{}
	FailedHosts  []interface{}
	EndTime      string
	CostTime     string
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
	endTime := time.Now()
	//result.StartTime = startTime.Format("2006-01-02 15:04:05")
	//result.EndTime = endTime.Format("2006-01-02 15:04:05")
	result.CostTime = endTime.Sub(startTime).String()
	summary := fmt.Sprintf("\nEnd Time: %s\nCost Time: %s\nTotal(%d) : Success=%d    Failed=%d", result.EndTime, result.CostTime, len(result.SuccessHosts)+len(result.FailedHosts), len(result.SuccessHosts), len(result.FailedHosts))
	return summary
}
