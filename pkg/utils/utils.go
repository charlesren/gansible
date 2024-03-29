/*
Copyright © 2019 Chuancheng Ren <renccn@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
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
	sumi := fmt.Sprintf("\nEnd Time: %s\nCost Time: %s\nTotal(%d) : Success=%d    Failed=%d    Unreachable=%d    Skipped=%d\n", endTimeStr, costTimeStr, totalNum, successNum, failedNum, unreachableNum, skippedNum)
	return sumi
}

//ColorSumInfo print result summary to standout with color
func ColorPrintSumInfo(sumr ResultSum) {
	sumr.EndTime = time.Now()
	sumr.CostTime = sumr.EndTime.Sub(sumr.StartTime)
	endTimeStr := sumr.EndTime.Format("2006-01-02 15:04:05")
	costTimeStr := sumr.CostTime.String()
	//totalNum := len(sumr.Failed) + len(sumr.Success) + len(sumr.Unreachable) + len(sumr.Skiped)
	totalNum := len(sumr.NodeResult)
	successNum := 0
	failedNum := 0
	unreachableNum := 0
	timeoutNum := 0
	skippedNum := 0
	for _, r := range sumr.NodeResult {
		if r.Result.Status == StatusSuccess {
			successNum++
		} else if r.Result.Status == StatusFailed {
			failedNum++
		} else if r.Result.Status == StatusUnreachable {
			unreachableNum++
		} else if r.Result.Status == StatusTimeout {
			timeoutNum++
		} else if r.Result.Status == StatusSkiped {
			skippedNum++
		}
	}
	fmt.Printf("\nEnd Time: %s\nCost Time: %s\n", endTimeStr, costTimeStr)
	fmt.Printf("Total(%d) : ", totalNum)
	if successNum > 0 {
		fmt.Printf("%s%s    ", color.GreenString("Success="), color.GreenString(strconv.Itoa(successNum)))
	} else {
		fmt.Printf("%s%d    ", "Success=", successNum)
	}
	if failedNum > 0 {
		fmt.Printf("%s%s    ", color.RedString("Failed="), color.RedString(strconv.Itoa(failedNum)))
	} else {
		fmt.Printf("%s%d    ", "Failed=", failedNum)
	}
	if unreachableNum > 0 {
		fmt.Printf("%s%s    ", color.RedString("Unreachable="), color.RedString(strconv.Itoa(unreachableNum)))
	} else {
		fmt.Printf("%s%d    ", "Unreachable=", unreachableNum)
	}
	if timeoutNum > 0 {
		fmt.Printf("%s%s    ", color.RedString("Timeout="), color.RedString(strconv.Itoa(timeoutNum)))
	} else {
		fmt.Printf("%s%d    ", "Timeout=", timeoutNum)
	}
	if skippedNum > 0 {
		fmt.Printf("%s%s    ", color.CyanString("Skipped="), color.CyanString(strconv.Itoa(skippedNum)))
	} else {
		fmt.Printf("%s%d\n", "Skipped=", skippedNum)
	}
}

//ColorPrintNodeResult print node result to standout with color
func ColorPrintNodeResult(noder NodeResult, outputStyle string) {
	colorPrint := func(status string, nrInfo string) {
		switch status {
		case StatusSuccess:
			color.Green(nrInfo)
		case StatusFailed:
			color.Red(nrInfo)
		case StatusUnreachable:
			color.Red(nrInfo)
		case StatusSkiped:
			color.Cyan(nrInfo)
		case StatusTimeout:
			color.Yellow(nrInfo)
		default:
			color.Red("Unkonwn Exect Result Status")
		}
	}
	switch outputStyle {
	case "gansible":
		nrInfo := NodeResultInfo(noder)
		colorPrint(noder.Result.Status, nrInfo)
	case "json":
		nrInfo, err := json.Marshal(noder)
		if err != nil {
			fmt.Println("marshal node result error:", err)
			return
		}
		colorPrint(noder.Result.Status, string(nrInfo))
	case "yaml":
		nrInfo, err := yaml.Marshal(noder)
		if err != nil {
			fmt.Println("marshal node result error:", err)
			return
		}
		colorPrint(noder.Result.Status, string(nrInfo))
	default:
		color.Red("incorrect output format")
	}
}

//PrintNodeResult print node result to standout
func PrintNodeResult(noder NodeResult, outputStyle string) {
	switch outputStyle {
	case "gansible":
		nrInfo := NodeResultInfo(noder)
		fmt.Println(nrInfo)
	case "json":
		nrInfo, err := json.Marshal(noder)
		if err != nil {
			fmt.Println("marshal node result error:", err)
			return
		}
		fmt.Println(string(nrInfo))
	case "yaml":
		nrInfo, err := yaml.Marshal(noder)
		if err != nil {
			fmt.Println("marshal node result error:", err)
			return
		}
		fmt.Println(string(nrInfo))
	default:
		fmt.Println("incorrect output format")
	}
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

//Loging save gansible command result to log file
func Loging(sumr ResultSum, logFileName string, logFileFormat string, logDir string) {
	if logFileName == "" {
		logFileName = fmt.Sprintf("gansible_%s", sumr.StartTime.Format("2006-01-02_15:04:05"))
	}
	if logDir == "" {
		logDir = os.TempDir()
	}
	logfile := path.Join(logDir, logFileName+"."+logFileFormat)
	switch logFileFormat {
	case "log":
		var nrInfo string
		for _, noder := range sumr.NodeResult {
			nr := NodeResultInfo(noder)
			nrInfo = nrInfo + nr
		}
		suminfo := SumInfo(sumr)
		info := nrInfo + suminfo
		err := AppendToFile(logfile, info)
		if err != nil {
			fmt.Println("loging failed err:", err)
			return
		}
		fmt.Printf("save log to file: %s successfully!\n", logfile)
	case "json":
		jsonInfo, err := json.Marshal(sumr)
		if err != nil {
			fmt.Println("marshal node result error:", err)
			return
		}
		info := string(jsonInfo)
		err = AppendToFile(logfile, info)
		if err != nil {
			fmt.Println("loging failed err:", err)
			return
		}
		fmt.Printf("save log to file: %s successfully!\n", logfile)
	case "yaml":
		yamlInfo, err := yaml.Marshal(sumr)
		if err != nil {
			fmt.Println("marshal node result error:", err)
			return
		}
		info := string(yamlInfo)
		err = AppendToFile(logfile, info)
		if err != nil {
			fmt.Println("loging failed err:", err)
			return
		}
		fmt.Printf("save log to file: %s successfully!\n", logfile)
	case "csv":
		logFileObj, err := os.Create(logfile)
		if err != nil {
			fmt.Println("create log file error:", err)
			return
		}
		defer logFileObj.Close()
		title := []string{"Node", "Status", "ReturnCode", "Out"}
		w := csv.NewWriter(logFileObj)
		if err := w.Write(title); err != nil {
			fmt.Println("error writing title to csv:", err)
			return
		}
		for _, nr := range sumr.NodeResult {
			record := []string{nr.Node, nr.Result.Status, nr.Result.RetrunCode, nr.Result.Out}
			if err := w.Write(record); err != nil {
				fmt.Println("error writing record to csv:", err)
				return
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			fmt.Println("loging failed err:", err)
			return
		}
		fmt.Printf("save log to file: %s successfully!\n", logfile)
	default:
		fmt.Printf("incorrect file format!\n")
	}
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

//ParseIPFile read given file then store proper ips into []sting
func ParseIPFile(ipFile string) ([]string, error) {
	ip := []string{}
	var err error
	if ipFile != "" {
		ipFile, err = filepath.Abs(ipFile)
		if err != nil {
			log.Printf("get %s filepath err: %s ", ipFile, err)
			return nil, err
		}
		file, err := os.Open(ipFile)
		if err != nil {
			log.Printf("can not open file: %s, err: [%v]", ipFile, err)
			return nil, err
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			tempip, err := ParseIPStr(scanner.Text())
			if err != nil {
				log.Printf("parse ip from file error line: %s  err: [%v]", scanner.Text(), err)
				return nil, err
			}
			ip = append(ip, tempip...)
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Cannot scanner file: %s, err: [%v]", ipFile, err)
			return nil, err
		}
	}
	return ip, nil
}

//RemoveDuplicateString remove duplicate element of string slice.
func RemoveDuplicateString(sli []string) []string {
	result := make([]string, 0, len(sli))
	temp := make(map[string]struct{})
	for _, item := range sli {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

//ParseIP parse ip file and ip string use ParseIPFile and ParseStr func then store proper ips into []sting
func ParseIP(ipFile, ipStr string) ([]string, error) {
	ip := []string{}
	strip, err := ParseIPStr(ipStr)
	if err != nil {
		log.Printf("parse ip form string %s err: %s ", ipStr, err)
		return nil, err
	}
	ip = append(ip, strip...)
	fileip, err := ParseIPFile(ipFile)
	if err != nil {
		log.Printf("parse ip form file %s err: %s ", ipFile, err)
		return nil, err
	}
	ip = append(ip, fileip...)
	ip = RemoveDuplicateString(ip)
	return ip, nil
}

//MuxShell
func MuxShell(stdin io.Writer, stdout, stderr io.Reader) (chan<- string, <-chan string) {
	in := make(chan string, 5)
	out := make(chan string, 5)
	var wg sync.WaitGroup
	wg.Add(1) //for the shell itself
	go func() {
		for cmd := range in {
			wg.Add(1)
			stdin.Write([]byte(cmd + "\n"))
			wg.Wait()
		}
	}()

	go func() {
		var (
			buf [1024 * 1024]byte
			t   int
		)
		for {
			n, err := stdout.Read(buf[t:])
			if err != nil {
				fmt.Println(err.Error())
				close(in)
				close(out)
				return
			}
			t += n
			result := string(buf[:t])
			if strings.Contains(string(buf[t-n:t]), "More") {
				stdin.Write([]byte("\n"))
			}
			if strings.Contains(result, "username:") ||
				strings.Contains(result, "password:") ||
				strings.Contains(result, ">") {
				out <- string(buf[:t])
				t = 0
				wg.Done()
			}
		}
	}()
	return in, out
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
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,     // enable echoing
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}
		if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
			log.Fatal("request for pseudo terminal failed: ", err)
		}
		var outbuf bytes.Buffer
		session.Stdout = &outbuf
		var errbuf bytes.Buffer
		session.Stderr = &errbuf
		stdin, err := session.StdinPipe()
		if err != nil {
			panic(err)
		}

		if err := session.Shell(); err != nil {
			log.Fatal("failed to start shell: ", err)
			execr.Status = StatusFailed
			execr.RetrunCode = "1"
			execr.Out = err.Error()
		} else {
			commands = strings.TrimRight(commands, ";")
			cmdlist := strings.Split(commands, ";")
			for _, cmd := range cmdlist {
				cmd = cmd + "\n"
				//cmd = cmd + " 2>&1\n"
				stdin.Write([]byte(cmd))
			}
			stdin.Write([]byte("exit\n"))
			if err = session.Wait(); err != nil {
				if _, ok := err.(*ssh.ExitError); ok {
					//return err.ExitStatus(), nil
					fmt.Println("1")
				} else {
					//return -1, errors.New("failed to wait ssh command: " + err.Error())
					fmt.Println("2")
				}
			}
			// trim output
			trimOut := func(out io.Reader) string {
				scanner := bufio.NewScanner(out)
				ok := false
				po := ""
				for scanner.Scan() {
					if ok {
						po = fmt.Sprintf("%v%v\n", po, scanner.Text())
					}
					if scanner.Text() == "exit" {
						ok = true
					}
				}
				if scanner.Err() != nil {
					fmt.Printf("trim output error: %s\n", scanner.Err())
					return ""
				}
				return po
			}
			pureOut := trimOut(&outbuf)
			execr.Status = StatusSuccess
			execr.RetrunCode = "0"
			execr.Out = fmt.Sprintf("%s%s\n", pureOut, errbuf.String())
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
func UploadFile(sftpClient *sftp.Client, srcFilePath string, destDir string) ExecResult {
	execr := ExecResult{}
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	defer srcFile.Close()
	var destFileName = path.Base(srcFilePath)
	destFile, err := sftpClient.Create(path.Join(destDir, destFileName))
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	defer destFile.Close()
	c, err := ioutil.ReadAll(srcFile)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	destFile.Write(c)
	execr.Status = StatusSuccess
	execr.RetrunCode = "0"
	execr.Out = "upload successfully!"
	return execr
}

//UploadDir upload local dir to dest dir of remote host
func UploadDir(sftpClient *sftp.Client, srcDir string, destDir string) ExecResult {
	execr := ExecResult{}
	srcTargets, err := ioutil.ReadDir(srcDir)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	for _, srcTarget := range srcTargets {
		srcTargetPath := path.Join(srcDir, srcTarget.Name())
		destTargetPath := path.Join(destDir, srcTarget.Name())
		if srcTarget.IsDir() {
			err = sftpClient.MkdirAll(destTargetPath)
			if err != nil {
				execr.Status = StatusFailed
				execr.RetrunCode = "1"
				execr.Out = err.Error()
				return execr
			}
			execr = UploadDir(sftpClient, srcTargetPath, destTargetPath)
		} else {
			execr = UploadFile(sftpClient, path.Join(srcDir, srcTarget.Name()), destDir)
		}
	}
	return execr
}

//Upload func upload file or directory to dest dir
func Upload(sftpClient *sftp.Client, src string, dest string) ExecResult {
	execr := ExecResult{}
	srcInfo, err := os.Stat(src)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	destInfo, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			err := sftpClient.MkdirAll(dest)
			if err != nil {
				execr.Status = StatusFailed
				execr.RetrunCode = "1"
				execr.Out = err.Error()
				return execr
			}
		} else {
			execr.Status = StatusFailed
			execr.RetrunCode = "1"
			execr.Out = err.Error()
			return execr
		}
	} else {
		if !destInfo.IsDir() {
			execr.Status = StatusFailed
			execr.RetrunCode = "1"
			execr.Out = fmt.Sprintf("%s is not directory", dest)
			return execr
		}
	}
	if srcInfo.IsDir() {
		execr = UploadDir(sftpClient, src, dest)
	} else {
		execr = UploadFile(sftpClient, src, dest)
	}
	return execr
}

//DownloadFile download remote file to local dir
func DownloadFile(sftpClient *sftp.Client, srcFilePath string, destDir string) ExecResult {
	execr := ExecResult{}
	srcFile, err := sftpClient.Open(srcFilePath)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	defer srcFile.Close()
	var destFileName = path.Base(srcFilePath)
	destFile, err := os.Create(path.Join(destDir, destFileName))
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	defer destFile.Close()
	if _, err = srcFile.WriteTo(destFile); err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	execr.Status = StatusSuccess
	execr.RetrunCode = "0"
	execr.Out = "download successfully!"
	return execr
}

//DownloadDir Download remote dir to local dir
func DownloadDir(sftpClient *sftp.Client, srcDir string, destDir string) ExecResult {
	execr := ExecResult{}
	srcTargets, err := sftpClient.ReadDir(srcDir)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	for _, srcTarget := range srcTargets {
		srcTargetPath := path.Join(srcDir, srcTarget.Name())
		destTargetPath := path.Join(destDir, srcTarget.Name())
		if srcTarget.IsDir() {
			err = os.MkdirAll(destTargetPath, os.ModePerm)
			if err != nil {
				execr.Status = StatusFailed
				execr.RetrunCode = "1"
				execr.Out = err.Error()
				return execr
			}
			execr = DownloadDir(sftpClient, srcTargetPath, destTargetPath)
		} else {
			execr = DownloadFile(sftpClient, path.Join(srcDir, srcTarget.Name()), destDir)
		}
	}
	return execr
}

//Download func Download file or directory to dest dir
func Download(sftpClient *sftp.Client, src string, dest string) ExecResult {
	execr := ExecResult{}
	srcInfo, err := sftpClient.Stat(src)
	if err != nil {
		execr.Status = StatusFailed
		execr.RetrunCode = "1"
		execr.Out = err.Error()
		return execr
	}
	destInfo, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dest, os.ModePerm)
			if err != nil {
				execr.Status = StatusFailed
				execr.RetrunCode = "1"
				execr.Out = err.Error()
				return execr
			}
		} else {
			execr.Status = StatusFailed
			execr.RetrunCode = "1"
			execr.Out = err.Error()
			return execr
		}
	} else {
		if !destInfo.IsDir() {
			execr.Status = StatusFailed
			execr.RetrunCode = "1"
			execr.Out = fmt.Sprintf("%s is not directory", dest)
			return execr
		}
	}
	if srcInfo.IsDir() {
		execr = DownloadDir(sftpClient, src, dest)
	} else {
		execr = DownloadFile(sftpClient, src, dest)
	}
	return execr
}
