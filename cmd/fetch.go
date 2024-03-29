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
	"path"
	"reflect"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Download file from remote nodes",
	Long: `Download file from remote nodes.
Fetch will create folder named by ip for each host.
The most typical example:  gansible	fetch  -n  "10.0.0.1;10.0.0.3-5;10.0.0.7-10.0.0.9"  -s /remote/file/or/dir -d /local/dir`,
	Args: cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var sumr utils.ResultSum
		sumr.StartTime = time.Now()
		ip, err := utils.ParseIP(nodeFile, nodes)
		if err != nil {
			fmt.Println(err)
		}
		if len(ip) == 0 {
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
				client, err = connect.DoSilent(keyPath, keyPassword, user, password, reflect.ValueOf(node).String(), port, sshTimeout, sshThreads, pwdFile)
				if err != nil {
					noder.Result.Status = utils.StatusUnreachable
					noder.Result.RetrunCode = "1"
					noder.Result.Out = err.Error()
					result <- noder
					utils.ColorPrintNodeResult(noder, outputStyle)
					wg.Done()
				} else {
					defer client.Close()
					var sftpClient *sftp.Client
					sftpClient, err = sftp.NewClient(client)
					if err != nil {
						noder.Result.Status = utils.StatusUnreachable
						noder.Result.RetrunCode = "1"
						noder.Result.Out = err.Error()
						result <- noder
						utils.ColorPrintNodeResult(noder, outputStyle)
					}
					noder.Result = utils.Download(sftpClient, src, path.Join(dest, reflect.ValueOf(node).String()))
					result <- noder
					utils.ColorPrintNodeResult(noder, outputStyle)
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
		utils.ColorPrintSumInfo(sumr)
		if loging {
			utils.Loging(sumr, logFileName, logFileFormat, logDir)
		}
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	fetchCmd.Flags().StringVarP(&src, "src", "s", "", "Source file or directory")
	fetchCmd.MarkFlagRequired("src")
	fetchCmd.Flags().StringVarP(&dest, "dest", "d", "", "Destination file or directory")
	fetchCmd.MarkFlagRequired("dest")
	fetchCmd.Flags().StringVarP(&nodes, "nodes", "n", "", "eg: 10.0.0.1;10.0.0.3-5;10.0.0.7-10.0.0.8")
	fetchCmd.Flags().StringVarP(&nodeFile, "nodefile", "f", "", "eg: /path/to/nodefile.txt  or ./nodefile.txt")
}
