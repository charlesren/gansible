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
	"log"
	"os"
	"path"

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
		fmt.Println("push called")
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
		srcFile, err := os.Open(src)
		if err != nil {
			fmt.Println("os.Open error : ", src)
			log.Fatal(err)

		}
		defer srcFile.Close()
		var destFileName = path.Base(src)
		destFile, err := sftpClient.Create(path.Join(dest, destFileName))
		if err != nil {
			fmt.Println("sftpClient.Create error : ", path.Join(dest, destFileName))
			log.Fatal(err)

		}
		defer destFile.Close()

		buf := make([]byte, 1024)
		for {
			n, _ := srcFile.Read(buf)
			if n == 0 {
				break
			}
			destFile.Write(buf)
		}

		fmt.Println("copy file to remote server finished!")
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
}
