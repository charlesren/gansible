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
	"log"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open a remote shell session to remote node",
	Long:  `Open a remote shell session to remote node,so that you can execute command just like in localhost.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		node := args[0]
		fmt.Println("Node:", node)
		client, err := connect.Do(keyPath, keyPassword, user, password, node, port, sshTimeout, sshThreads, pwdFile)
		if err != nil {
			log.Fatal(err)
		} else {
			defer client.Close()
			session, err := client.NewSession()
			if err != nil {
				log.Fatal(err)
			}
			defer session.Close()
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr
			session.Stdin = os.Stdin
			modes := ssh.TerminalModes{
				ssh.ECHO:          1,     // enable echoing
				ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
				ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
			}
			fileDescriptor := int(os.Stdin.Fd())
			if terminal.IsTerminal(fileDescriptor) {
				originalState, err := terminal.MakeRaw(fileDescriptor)
				if err != nil {
					log.Fatalf("failed to get original state")
				}
				defer terminal.Restore(fileDescriptor, originalState)
				termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
				if err != nil {
					log.Fatalf("failed to get term width and  term height")
				}
				if err := session.RequestPty("xterm-256color", termHeight, termWidth, modes); err != nil {
					log.Fatalf("request for pseudo terminal failed: %s", err)
				}
			}
			if err := session.Shell(); err != nil {
				log.Fatalf("failed to start shell: %s", err)
			}
			session.Wait()
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
