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
	"gansible/pkg/utils"
	"reflect"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var src string
var dest string

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Upload file to remote host",
	Long:  `Upload file to remote host.`,
	Run: func(cmd *cobra.Command, args []string) {
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
			p, _ := ants.NewPoolWithFunc(forks, func(host interface{}) {
				noder := utils.NodeResult{}
				noder.Node = reflect.ValueOf(host).String()
				var client *ssh.Client
				client, err = utils.TryPasswords("root", passwords, reflect.ValueOf(host).String(), 22, 30)
				if err != nil {
					noder.Result.Status = "Unreachable"
					noder.Result.RetrunCode = "1"
					noder.Result.Out = err.Error()
					nrInfo := utils.NodeResultInfo(noder)
					result <- noder
					fmt.Println(nrInfo)
					fmt.Printf("\n")
					wg.Done()
				} else {
					defer client.Close()
					var sftpClient *sftp.Client
					sftpClient, err = sftp.NewClient(client)
					if err != nil {
						noder.Result.Status = "Unreachable"
						noder.Result.RetrunCode = "1"
						noder.Result.Out = err.Error()
						nrInfo := utils.NodeResultInfo(noder)
						result <- noder
						fmt.Println(nrInfo)
						fmt.Printf("\n")
						wg.Done()
					}
					noder.Result = utils.Upload(sftpClient, src, dest)
					nrInfo := utils.NodeResultInfo(noder)
					result <- noder
					fmt.Println(nrInfo)
					fmt.Printf("\n")
					wg.Done()
				}
			})
			defer p.Release()
			go func() {
				for i := 0; i <= len(ip); i++ {
					t := <-result
					sumr.NodeResult = append(sumr.NodeResult, t)
				}
			}()
			for _, host := range ip {
				wg.Add(1)
				p.Invoke(host)
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
	rootCmd.AddCommand(pushCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pushCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	pushCmd.Flags().StringVarP(&src, "src", "s", "", "Source file or directory")
	pushCmd.MarkFlagRequired("src")
	pushCmd.Flags().StringVarP(&dest, "dest", "d", "", "Destination file or directory")
	pushCmd.MarkFlagRequired("dest")
	pushCmd.Flags().StringVarP(&nodes, "nodes", "n", "", "eg: 10.0.0.1;10.0.0.2-5;10.0.0.6-10.0.0.8")
	pushCmd.Flags().StringVarP(&nodeFile, "nodefile", "f", "", "eg: /path/to/nodefile.txt  or ./nodefile.txt")
}
