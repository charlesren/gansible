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

//Package cmd is the central point of the gansible application.
//Gansible commands are defined in this package.
package cmd

import (
	"fmt"
	"gansible/pkg/connect"
	"gansible/pkg/utils"
	"reflect"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var commands string
var wg sync.WaitGroup
var timeout int

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run commands on multiple nodes in parallel",
	Long: `Run commands on multiple nodes in parallel,return result when finished.Default number of concurrenrt tasks is 5.
Default timeout of each task is 300 seconds.`,
	Args: cobra.MaximumNArgs(1),
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
			p, _ := ants.NewPoolWithFunc(forks, func(node interface{}) {
				noder := utils.NodeResult{}
				noder.Node = reflect.ValueOf(node).String()
				var client *ssh.Client
				client, err = connect.DoSilent(keyPath, keyPassword, user, password, reflect.ValueOf(node).String(), port, sshTimeout, pwdFile)
				if err != nil {
					noder.Result.Status = "Unreachable"
					noder.Result.RetrunCode = "1"
					noder.Result.Out = err.Error()
					nrInfo := utils.NodeResultInfo(noder)
					result <- noder
					fmt.Println(nrInfo)
					fmt.Printf("\n")
				} else {
					defer client.Close()
					noder.Result = utils.Execute(client, commands, timeout)
					result <- noder
					utils.PrintNodeResult(noder, outputStyle)
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
		fmt.Printf("%s", sumrinfo)
		if loging {
			utils.Loging(sumr, logFileName, logFileFormat, logDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	runCmd.Flags().StringVarP(&commands, "commands", "c", "", "separate multiple command with semicolons(eg: pwd;ls)")
	runCmd.Flags().StringVarP(&nodes, "nodes", "n", "", "eg: 10.0.0.1;10.0.0.2-5;10.0.0.6-10.0.0.8")
	runCmd.Flags().StringVarP(&nodeFile, "nodefile", "f", "", "eg: /path/to/nodefile.txt  or ./nodefile.txt")
	runCmd.Flags().IntVarP(&timeout, "timeout", "", 300, "task should finished before timeout")
	runCmd.MarkFlagRequired("commands")
}
