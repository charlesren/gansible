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
	"gansible/pkg/utils"
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a remote shell session to remote machine",
	Long:  `Open a remote shell session to remote machine,so that you can execute command just like in localhost.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		passwords := utils.GetPassword(pwdFile)
		host := args[0]
		fmt.Println("Host:", host)
		var client *ssh.Client
		client, _ = utils.TryPasswords("root", passwords, host, 22, sshTimeout)
		if client == nil {
			fmt.Println("All passwords are wrong.")
		} else {
			defer client.Close()
			session, err := client.NewSession()
			if err != nil {
				log.Fatal(err)
			}
			defer session.Close()
			/*
				//Exec cmd then quit
				cmd := "date"
				out, err := session.CombinedOutput(cmd)
				if err != nil {
					log.Fatal("Remote Exec Field:", err)
				}
				fmt.Println("Remote Exec Output:\n", string(out))
			*/
			//Start shell on remote host then interactive
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr
			session.Stdin = os.Stdin
			modes := ssh.TerminalModes{
				ssh.ECHO:          0,     // disable echoing
				ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
				ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
			}
			err = session.RequestPty("xterm-256color", 40, 80, modes)
			if err != nil {
				log.Fatal(err)
			}
			err = session.Shell()
			err = session.Wait()
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// shellCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// shellCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
