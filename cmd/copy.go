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
	"gansible/src/autologin"
	"log"
	"os"
	"path"

	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var src string
var dest string

// copyCmd represents the copy command
var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("copy called")
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
		var destFile = path.Base(src)
		dstFile, err := sftpClient.Create(path.Join(dest, destFile))
		if err != nil {
			fmt.Println("sftpClient.Create error : ", path.Join(dest, destFile))
			log.Fatal(err)

		}
		defer dstFile.Close()

		buf := make([]byte, 1024)
		for {
			n, _ := srcFile.Read(buf)
			if n == 0 {
				break
			}
			dstFile.Write(buf)
		}

		fmt.Println("copy file to remote server finished!")
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// copyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// copyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	copyCmd.Flags().StringVarP(&src, "src", "s", "", "Source file or directory")
	copyCmd.Flags().StringVarP(&dest, "dest", "d", "", "Destination file or directory")
}
