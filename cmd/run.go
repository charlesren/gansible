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
	"sync"

	"github.com/spf13/cobra"
)

var commands string
var hosts string
var size int
var ch = make(chan string, size)
var wg sync.WaitGroup

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ip, err := utils.ParseIPStr(hosts)
		if err != nil {
			fmt.Println(err)
		}
		if ip == nil {
			fmt.Println("No hosts specified!!!")
		} else {
			for _, host := range ip {
				go func(commands string) {
					host := <-ch
					runr := utils.DoCommand(host, commands)
					runinfo := utils.RunInfo(runr)
					fmt.Println(runinfo)
					wg.Done()
				}(commands)
				ch <- host
				wg.Add(1)
			}
			wg.Wait()
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
	runCmd.Flags().StringVarP(&hosts, "hosts", "H", "", "eg: 10.0.0.1;10.1.1.1-3;10.2.2.2-10.2.2.5;10.3.3.1/31")
	runCmd.MarkFlagRequired("commands")
	runCmd.MarkFlagRequired("hosts")
}
