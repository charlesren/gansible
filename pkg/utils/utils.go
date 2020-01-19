package utils

import (
	"fmt"
	"gansible/pkg/autologin"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

const (
	//StatusSuccess ...
	StatusSuccess string = "Success"
	//StatusFailed ...
	StatusFailed string = "Failed"
	//StatusUnreachable ...
	StatusUnreachable string = "Unreachable"
	//StatusSkiped ...
	StatusSkiped string = "Skipped"
	//StatusTimeout ...
	StatusTimeout string = "Timeout"
)

//ResultSum struct store execute summary result of gansible command
type ResultSum struct {
	StartTime  time.Time
	EndTime    time.Time
	CostTime   time.Duration
	NodeResult []NodeResult
}

//NodeResult struct store task result
type NodeResult struct {
	Node   string
	Result ExecResult
}

//ExecResult struct store command execute result
type ExecResult struct {
	Status     string
	RetrunCode string
	Out        string
}

//NodeResultInfo gengrate information from NodeResult
func NodeResultInfo(nodeResult NodeResult) string {
	nrInfo := fmt.Sprintf("%s | %s | rc=%s >>\n%s", nodeResult.Node, nodeResult.Result.Status, nodeResult.Result.RetrunCode, nodeResult.Result.Out)
	return nrInfo
}

//SumInfo gengrate summary of gansible result
func SumInfo(sumr ResultSum) string {
	sumr.EndTime = time.Now()
	sumr.CostTime = sumr.EndTime.Sub(sumr.StartTime)
	endTimeStr := sumr.EndTime.Format("2006-01-02 15:04:05")
	costTimeStr := sumr.CostTime.String()
	//totalNum := len(sumr.Failed) + len(sumr.Success) + len(sumr.Unreachable) + len(sumr.Skiped)
	totalNum := len(sumr.NodeResult)
	successNum := 0
	failedNum := 0
	unreachableNum := 0
	skippedNum := 0
	for _, r := range sumr.NodeResult {
		if r.Result.Status == "Success" {
			successNum++
		} else if r.Result.Status == "Failed" {
			failedNum++
		} else if r.Result.Status == "Unreachable" {
			unreachableNum++
		} else if r.Result.Status == "Skipped" {
			skippedNum++
		}
	}
	sumi := fmt.Sprintf("\nEnd Time: %s\nCost Time: %s\nTotal(%d) : Success=%d    Failed=%d    Unreachable=%d    Skipped=%d", endTimeStr, costTimeStr, totalNum, successNum, failedNum, unreachableNum, skippedNum)
	return sumi
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

//TryPasswords ssh to a machine using a set of possible passwords concurrently.
func TryPasswords(user string, passwords []string, host string, port int, sshTimeout int) (*ssh.Client, error) {
	timer := time.NewTimer(time.Duration(sshTimeout) * time.Second)
	defer timer.Stop()
	ch := make(chan *ssh.Client)
	count := 0
	var mutex sync.Mutex
	finish := make(chan bool)
	errTimeout := fmt.Errorf("Time out in %d seconds", sshTimeout)
	errAllPassWrong := fmt.Errorf("All passwords are wrong")
	for _, password := range passwords {
		go func(password string) {
			c, err := autologin.Connect("root", password, host, 22)
			if err == nil {
				ch <- c
			} else {
				mutex.Lock()
				count = count + 1
				if count == len(passwords) {
					finish <- true
				}
				mutex.Unlock()
			}
		}(password)
	}
	select {
	case client := <-ch:
		return client, nil
	case <-finish:
		return nil, errAllPassWrong
	case <-timer.C:
		return nil, errTimeout
	}
}

//Execute run given commands  on  ssh clinet then return ExecResult
func Execute(client *ssh.Client, commands string, timeout int) ExecResult {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)
	defer timer.Stop()
	execr := ExecResult{}
	session, err := client.NewSession()
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
	} else {
		defer session.Close()
		//Exec cmd then quit
		commands = strings.TrimRight(commands, ";")
		command := strings.Split(commands, ";")
		cmdNew := strings.Join(command, "&&")
		out, err := session.CombinedOutput(cmdNew)
		if err != nil {
			execr.Status = StatusFailed
			execr.RetrunCode = "1"
			execr.Out = string(out)
		} else {
			execr.Status = StatusSuccess
			execr.RetrunCode = "0"
			execr.Out = string(out)
		}
	}
	//send ExecResult
	ch := make(chan bool, 1)
	ch <- true
	close(ch)
	select {
	case <-ch:
		return execr
	case <-timer.C:
		execr.Status = StatusTimeout
		execr.RetrunCode = "1"
		execr.Out = fmt.Sprintf("Task not finished before %d seconds", timeout)
		return execr
	}
}

//UploadFile upload local file to dest dir of remote host
func UploadFile(sftpClient *sftp.Client, srcFilePath string, destDir string) {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		fmt.Println("os.Open error : ", srcFilePath)
		log.Fatal(err)

	}
	defer srcFile.Close()
	var destFileName = path.Base(srcFilePath)
	destFile, err := sftpClient.Create(path.Join(destDir, destFileName))
	if err != nil {
		fmt.Println("sftpClient.Create error : ", path.Join(destDir, destFileName))
		log.Fatal(err)

	}
	defer destFile.Close()
	c, err := ioutil.ReadAll(srcFile)
	if err != nil {
		fmt.Println("ReadAll error : ", srcFilePath)
		log.Fatal(err)
	}
	destFile.Write(c)
	fmt.Println("copy file to remote server finished!")
}

//UploadDir upload local dir to dest dir of remote host
func UploadDir(sftpClient *sftp.Client, srcDir string, destDir string) {
	srcTargets, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal("read dir list fail ", err)
	}
	for _, srcTarget := range srcTargets {
		srcTargetPath := path.Join(srcDir, srcTarget.Name())
		destTargetPath := path.Join(destDir, srcTarget.Name())
		if srcTarget.IsDir() {
			err = sftpClient.MkdirAll(destTargetPath)
			if err != nil {
				fmt.Println("create directory failed!")
				log.Fatal(err)
			}
			UploadDir(sftpClient, srcTargetPath, destTargetPath)
		} else {
			UploadFile(sftpClient, path.Join(srcDir, srcTarget.Name()), destDir)
		}
	}
	fmt.Println("copy directory to remote server finished!")
}

//Upload func upload file or directory to dest dir
func Upload(sftpClient *sftp.Client, src string, dest string) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}
	destInfo, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			err := sftpClient.MkdirAll(dest)
			if err != nil {
				fmt.Println("create directory failed!")
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	} else {
		if !destInfo.IsDir() {
			log.Fatal(fmt.Errorf("%s is not directory", dest))
		}
	}
	if srcInfo.IsDir() {
		UploadDir(sftpClient, src, dest)
	} else {
		UploadFile(sftpClient, src, dest)
	}
}
