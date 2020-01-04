package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

//SumResult struct store execute summary result of gansible command
type SumResult struct {
	StartTime   time.Time
	EndTime     time.Time
	CostTime    time.Duration
	Success     []interface{}
	Failed      []interface{}
	Unreachable []interface{}
	Skiped      []interface{}
	TotalHosts  []interface{}
}

//RunResult struct store cmd run result of ssh session
type RunResult struct {
	Host       string
	Status     string
	RetrunCode string
	Result     string
}

//RunInfo gengrate information of cmd result executed by ssh session
func RunInfo(runr RunResult) string {
	runInfo := fmt.Sprintf("%s | %s | rc=%s >>\n%s\n\n", runr.Host, runr.Status, runr.RetrunCode, runr.Result)
	return runInfo
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

//SumInfo gengrate summary of gansible result
func SumInfo(sumr SumResult, startTime time.Time) string {
	sumr.EndTime = time.Now()
	sumr.CostTime = sumr.EndTime.Sub(startTime)
	endTimeStr := sumr.EndTime.Format("2006-01-02 15:04:05")
	costTimeStr := sumr.CostTime.String()
	totalNum := len(sumr.Failed) + len(sumr.Success) + len(sumr.Unreachable) + len(sumr.Skiped)
	sumi := fmt.Sprintf("\nEnd Time: %s\nCost Time: %s\nTotal(%d) : Success=%d    Failed=%d    Unreachable=%d    Skipped=%d", endTimeStr, costTimeStr, totalNum, len(sumr.Success), len(sumr.Failed), len(sumr.Unreachable), len(sumr.Skiped))
	return sumi
}
