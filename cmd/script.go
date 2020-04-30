//Package cmd ...
/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"fmt"
	"gansible/pkg/connect"
	"gansible/pkg/utils"
	"path"
	"reflect"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var dir string
var scriptArgs string

// scriptCmd represents the script command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Run local script on remote nodes",
	Long:  `Run local script on remote nodes.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scriptFile := args[0]
		var sumr utils.ResultSum
		sumr.StartTime = time.Now()
		ip, err := utils.ParseIP(nodeFile, nodes)
		if err != nil {
			fmt.Println(err)
		}
		if ip == nil {
			fmt.Println("No node specified!")
		} else {
			if forks < 1 {
				forks = 1
			} else if forks > 10000 {
				fmt.Println("Max forks is 10000")
				return
			}
			result := make(chan utils.NodeResult, len(ip))
			p, _ := ants.NewPoolWithFunc(forks, func(node interface{}) {
				noder := utils.NodeResult{}
				noder.Node = reflect.ValueOf(node).String()
				var client *ssh.Client
				client, err = connect.DoSilent(keyPath, keyPassword, user, password, reflect.ValueOf(node).String(), port, sshTimeout, pwdFile)
				if err != nil {
					noder.Result.Status = "Unreachable"
					noder.Result.RetrunCode = "1"
					noder.Result.Out = err.Error()
					result <- noder
					utils.PrintNodeResult(noder, outputStyle)
				} else {
					defer client.Close()
					var sftpClient *sftp.Client
					sftpClient, err = sftp.NewClient(client)
					if err != nil {
						noder.Result.Status = "Unreachable"
						noder.Result.RetrunCode = "1"
						noder.Result.Out = err.Error()
						result <- noder
						utils.PrintNodeResult(noder, outputStyle)
					}
					noder.Result = utils.Upload(sftpClient, scriptFile, "/tmp")
					//tempDir := os.TempDir()
					var destFileName = path.Base(scriptFile)
					destFilePath := path.Join("/tmp", destFileName)
					cmd := "sh " + destFilePath
					if scriptArgs != "" {
						cmd = cmd + " " + scriptArgs
					}
					if dir != "" {
						cmd = "cd " + dir + ";" + cmd
					}
					noder.Result = utils.Execute(client, cmd, timeout)
					nrInfo := utils.NodeResultInfo(noder)
					result <- noder
					fmt.Println(nrInfo)
				}
			})
			defer p.Release()
			go func() {
				for i := 0; i <= len(ip); i++ {
					t := <-result
					sumr.NodeResult = append(sumr.NodeResult, t)
					wg.Done()
				}
			}()
			for _, node := range ip {
				wg.Add(1)
				p.Invoke(node)
			}
			wg.Wait()
		}
		sumrinfo := utils.SumInfo(sumr)
		fmt.Println(sumrinfo)
		if loging {
			utils.Loging(sumr, logFileName, logFileFormat, logDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(scriptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scriptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scriptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	scriptCmd.Flags().StringVarP(&dir, "dir", "d", "", "run script at designated dir")
	scriptCmd.Flags().StringVarP(&scriptArgs, "args", "a", "", "args for script")
	scriptCmd.Flags().StringVarP(&nodes, "nodes", "n", "", "eg: 10.0.0.1;10.0.0.2-5;10.0.0.6-10.0.0.8")
	scriptCmd.Flags().StringVarP(&nodeFile, "nodefile", "f", "", "eg: /path/to/nodefile.txt  or ./nodefile.txt")
}
