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
	"gansible/pkg/autologin"
	"gansible/pkg/utils"
	"log"
	"os"
	"path"

	"github.com/panjf2000/ants/v2"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var dir string

// scriptCmd represents the script command
var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scriptFile := args[0]
		ip, err := utils.ParseIPStr(hosts)
		if err != nil {
			fmt.Println(err)
		}
		if ip == nil {
			fmt.Println("No hosts specified!!!")
		} else {
			size = 100
			p, _ := ants.NewPool(size)
			defer p.Release()
			for _, host := range ip {
				wg.Add(1)
				_ = p.Submit(func() {
					passwords := []string{"abc", "passw0rd"}
					var client *ssh.Client
					var err error
					for _, password := range passwords {
						if client, err = autologin.Connect("root", password, host, 22); err == nil {
							break
						}
					}
					defer client.Close()
					var sftpClient *sftp.Client
					sftpClient, err = sftp.NewClient(client)
					if err != nil {
						log.Fatal(err)
					}
					srcFile, err := sftpClient.Open(scriptFile)
					if err != nil {
						log.Fatal(err)
					}
					defer srcFile.Close()
					tempDir := os.TempDir()
					var destFileName = path.Base(scriptFile)
					destFilePath := path.Join(tempDir, destFileName)
					destFile, err := os.Create(destFilePath)
					if err != nil {
						log.Fatal(err)
					}
					defer destFile.Close()

					if _, err = srcFile.WriteTo(destFile); err != nil {
						log.Fatal(err)
					}
					session, err := client.NewSession()
					if err != nil {
						log.Fatal(err)
					}
					defer session.Close()
					cmd := "sh " + destFilePath
					if dir != "" {
						cmd = "cd " + dir + ";" + cmd
					}
					out, _ := session.CombinedOutput(cmd)
					fmt.Println(string(out))
					wg.Done()
				})
			}
			wg.Wait()
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
	scriptCmd.Flags().StringVarP(&hosts, "hosts", "H", "", "eg: 10.0.0.1;10.1.1.1-3;10.2.2.2-10.2.2.5;10.3.3.1/31")
	scriptCmd.MarkFlagRequired("hosts")
	scriptCmd.Flags().StringVarP(&dir, "dir", "d", "", "run script at designated dir")
}
