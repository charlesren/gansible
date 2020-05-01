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
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	forks         int
	nodes         string
	nodeFile      string
	loging        bool
	logDir        string
	logFileName   string
	logFileFormat string
	outputStyle   string
	keyPath       string
	keyPassword   string
	user          string
	password      string
	node          string
	port          int
	sshTimeout    int
	pwdFile       string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gansible",
	Short: "Gansible is a lightweight cli tool designed for system administrator",
	Long:  `Gansible is a lightweight cli tool designed for system administrator.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gansible.yaml)")
	rootCmd.PersistentFlags().IntVar(&forks, "forks", 5, "number of concurrenrt tasks")
	rootCmd.PersistentFlags().BoolVarP(&loging, "loging", "", false, "save result log")
	rootCmd.PersistentFlags().StringVar(&logFileFormat, "log-file-format", "csv", "log file format: log/json/yaml/csv")
	rootCmd.PersistentFlags().StringVar(&logDir, "log-dir", "", "dir to save log file")
	rootCmd.PersistentFlags().StringVar(&logFileName, "log-file-name", "", "define name of log file")
	rootCmd.PersistentFlags().StringVarP(&outputStyle, "output", "o", "gansible", "gansible output style: gansible/json/yaml")
	rootCmd.PersistentFlags().StringVar(&keyPath, "keyPath", "", "ssh private key file path")
	rootCmd.PersistentFlags().StringVar(&keyPassword, "keyPassword", "", "password of ssh private key")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "user used to login remote server")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "password of remote server")
	//rootCmd.Flags().StringVarP(&nodes, "nodes", "n", "", "eg: 10.0.0.1;10.0.0.2-5;10.0.0.6-10.0.0.8")
	rootCmd.Flags().IntVar(&port, "port", 22, "port used to login remote server")
	rootCmd.PersistentFlags().IntVar(&sshTimeout, "ssh-timeout", 30, "login should be successful before timeout")
	rootCmd.PersistentFlags().StringVar(&pwdFile, "pwdfile", "", "password file (default is $HOME/.pwdfile)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gansible" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gansible")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
