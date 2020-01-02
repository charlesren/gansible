//Package cmd ...
/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
	"log"
	"os"
	"path"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("fetch called")
		host := args[0]
		fmt.Println("Host:", host)
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
		srcFile, err := sftpClient.Open(src)
		if err != nil {
			log.Fatal(err)
		}
		defer srcFile.Close()

		var destFileName = path.Base(src)
		destFile, err := os.Create(path.Join(dest, destFileName))
		if err != nil {
			log.Fatal(err)
		}
		defer destFile.Close()

		if _, err = srcFile.WriteTo(destFile); err != nil {
			log.Fatal(err)
		}

		fmt.Println("copy file from remote server finished!")
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
}
