package utils

import (
	"fmt"
	"gansible/pkg/autologin"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
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
	runInfo := fmt.Sprintf("%s | %s | rc=%s >>\n%s", runr.Host, runr.Status, runr.RetrunCode, runr.Result)
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

//ParseIPStr parse  given string then store proper ips into []sting
func ParseIPStr(ipStr string) ([]string, error) {
	var IP []string
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return IP, nil
	}
	if strings.HasPrefix(ipStr, "#") {
		return IP, nil
	}
	ipStr = strings.TrimRight(ipStr, ";")
	fields := strings.Split(ipStr, ";")
	for _, field := range fields {
		if strings.Contains(field, "-") {
			f := strings.Split(field, "-")
			startIP := f[0]
			if net.ParseIP(startIP) == nil {
				return IP, fmt.Errorf("Illegal IP range,Plesae check")
			}
			startIPBlock := strings.Split(startIP, ".")
			startIPPrefix := startIPBlock[0 : len(startIPBlock)-1]
			startIPLastNo, _ := strconv.Atoi(startIPBlock[len(startIPBlock)-1])
			end := f[1]
			// dValue is d-dalue between startIPLastNo and endIPLastNo
			var dValue int
			if net.ParseIP(end) != nil {
				// string after - is IP
				endIPBlock := strings.Split(end, ".")
				endIPPrefix := endIPBlock[0 : len(endIPBlock)-1]
				endIPLastNo, _ := strconv.Atoi(endIPBlock[len(endIPBlock)-1])
				for i := 0; i < len(startIPPrefix); i++ {
					if startIPPrefix[i] != endIPPrefix[i] {
						return IP, fmt.Errorf("Illegal IP range,Plesae check")
					}
				}
				dValue = endIPLastNo - startIPLastNo
			} else {
				// string after - should be a number
				endIPLastNo, err := strconv.Atoi(end)
				if err != nil {
					return IP, fmt.Errorf("Illegal IP range,Plesae check")
				}
				dValue = endIPLastNo - startIPLastNo
			}
			switch {
			case dValue < 0:
				return IP, fmt.Errorf("Illegal IP range,Plesae check")
			case dValue == 0:
				IP = append(IP, startIP)
				return IP, nil
			case dValue > 0:
				for i := 0; i <= dValue; i++ {
					newIPBlock := append(startIPPrefix, strconv.Itoa(startIPLastNo+i))
					newIP := strings.Join(newIPBlock, ".")
					IP = append(IP, newIP)
				}
				return IP, nil
			}
		} else if strings.Contains(field, "/") {
			splitted := strings.Split(field, "/")
			if net.ParseIP(splitted[0]) == nil {
				return IP, fmt.Errorf("Illegal IP range,Plesae check")
			}
			//convert netmask style to cidr
			if strings.Contains(splitted[1], ".") {
				maskBlock := strings.Split(splitted[1], ".")
				ones, _ := net.IPv4Mask([]byte(maskBlock[0])[0], []byte(maskBlock[1])[0], []byte(maskBlock[2])[0], []byte(maskBlock[3])[0]).Size()
				splitted[1] = strconv.Itoa(ones)
				field = strings.Join(splitted, "/")
			}
			if splitted[1] == "32" {
				return []string{splitted[0]}, nil
			}
			ip, ipnet, err := net.ParseCIDR(field)
			if err != nil {
				return nil, err
			}

			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
				IP = append(IP, ip.String())
			}
			// remove network address and broadcast address
			return IP[1 : len(IP)-1], nil
		} else {
			ip := net.ParseIP(field)
			if ip == nil {
				return IP, fmt.Errorf("Illegal IP range,Plesae check")
			}
			IP = append(IP, ip.String())
			return IP, nil
		}
	}

	return IP, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

//DoCommand open a ssh session on a given host and run given commands  then return result
func DoCommand(host string, commands string) RunResult {
	runr := RunResult{}
	runr.Host = host
	passwords := []string{"abc", "passw0rd"}
	var client *ssh.Client
	var err error
	for _, password := range passwords {
		if client, err = autologin.Connect("root", password, runr.Host, 22); err == nil {
			break
		}
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	//Exec cmd then quit
	commands = strings.TrimRight(commands, ";")
	command := strings.Split(commands, ";")
	cmdNew := strings.Join(command, "&&")
	out, err := session.CombinedOutput(cmdNew)
	if err != nil {
		runr.Status = "Failed"
		runr.RetrunCode = "1"
	}
	runr.Status = "Success"
	runr.RetrunCode = "0"
	runr.Result = string(out)
	return runr
}
