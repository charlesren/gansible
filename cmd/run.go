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
	"log"
	"strings"

	"gansible/pkg/autologin"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var commands string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		host := args[0]
		fmt.Printf("%s >>\n", host)
		passwords := []string{"abc", "passw0rd"}
		var client *ssh.Client
		var err error
		for _, password := range passwords {
			if client, err = autologin.Connect("root", password, host, 22); err == nil {
				break
			}
		}
		defer client.Close()
		session, err := client.NewSession()
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()
		//Exec cmd then quit
		commands = strings.TrimRight(commands, ";")
		command := strings.Split(commands, ";")
		cmdNew := strings.Join(command, "&&")
		out, err := session.CombinedOutput(cmdNew)
		if err != nil {
			log.Fatal("Remote Exec Field:", err)
		}
		fmt.Printf("%s", string(out))
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
	runCmd.Flags().StringVarP(&commands, "commands", "c", "", "separate multiple command with semicolons. (Example:  pwd;ls)")
	runCmd.MarkFlagRequired("commands")
}
