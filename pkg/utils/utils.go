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
	fields := strings.Split(ipStr, ";")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if strings.HasPrefix(field, "#") {
			continue
		}
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
				continue
			case dValue > 0:
				for i := 0; i <= dValue; i++ {
					newIPBlock := append(startIPPrefix, strconv.Itoa(startIPLastNo+i))
					newIP := strings.Join(newIPBlock, ".")
					IP = append(IP, newIP)
				}
				continue
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
				IP = append(IP, splitted[0])
				continue
				//	return []string{splitted[0]}, nil
			}
			ip, ipnet, err := net.ParseCIDR(field)
			if err != nil {
				return nil, err
			}

			var tempIP []string
			for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
				tempIP = append(tempIP, ip.String())
			}
			// remove network address and broadcast address
			for i := 1; i < len(tempIP)-1; i++ {
				IP = append(IP, tempIP[i])
			}
			//return IP[1 : len(IP)-1], nil
			continue
		} else {
			ip := net.ParseIP(field)
			if ip == nil {
				return IP, fmt.Errorf("Illegal IP range,Plesae check")
			}
			IP = append(IP, ip.String())
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
func DoCommand(host string, commands string, timeout int) RunResult {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()
	runr := RunResult{}
	runr.Host = host
	passwords := []string{"abc", "passw0rd"}
	var client *ssh.Client
	var err error
	client, err = TryPasswords("root", passwords, host, 22, 30)
	if client != nil {
		defer client.Close()
	} else {
		runr.Status = "Unreachable"
		runr.RetrunCode = "1"
		runr.Result = "All passwords are wrong"
		return runr
	}
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
	ch := make(chan bool, 1)
	ch <- true
	close(ch)
	select {
	case <-ch:
		return runr
	case <-timer.C:
		runr.Status = "TimeOut"
		runr.RetrunCode = "1"
		runr.Result = fmt.Sprintf("Task not finished before %d seconds", timeout)
		return runr
	}
}

//TryPasswords ssh to a machine using a set of possible passwords concurrently.
func TryPasswords(user string, passwords []string, host string, port int, sshTimeout int) (*ssh.Client, error) {
	timer := time.NewTimer(time.Duration(sshTimeout) * time.Second)
	defer timer.Stop()
	ch := make(chan *ssh.Client)
	errTimeout := fmt.Errorf("Time out in %d seconds", sshTimeout)
	for _, password := range passwords {
		go func(password string) {
			c, err := autologin.Connect("root", password, host, 22)
			if err == nil {
				ch <- c
			} else {
			}
		}(password)
	}
	select {
	case client := <-ch:
		return client, nil
	case <-timer.C:
		return nil, errTimeout
	}
}

//ExecResult struct store command execute result
type ExecResult struct {
	Status     string
	RetrunCode string
	Result     string
}

//Execute run given commands  on  ssh clinet then return ExecResult
func Execute(client *ssh.Client, commands string, timeout int) ExecResult {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()
	runr := ExecResult{}
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
	ch := make(chan bool, 1)
	ch <- true
	close(ch)
	select {
	case <-ch:
		return runr
	case <-timer.C:
		runr.Status = "TimeOut"
		runr.RetrunCode = "1"
		runr.Result = fmt.Sprintf("Task not finished before %d seconds", timeout)
		return runr
	}
}
