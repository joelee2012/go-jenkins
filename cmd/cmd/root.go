/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// var client *jenkins.Jenkins

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jenksctl",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

type JenkinsOpts struct {
	URL      string
	User     string
	Password string
	Path     string
	Depth    int
}

var opts = &JenkinsOpts{}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jenkins.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().StringVar(&opts.URL, "url", "", "the jenkins url")
	rootCmd.PersistentFlags().StringVarP(&opts.User, "user", "u", "", "jenkins user")
	rootCmd.PersistentFlags().StringVarP(&opts.Password, "password", "p", "", "jenkins password")
	rootCmd.PersistentFlags().StringVarP(&opts.Path, "path", "P", "", "jobpath")
	viper.BindPFlags(rootCmd.Flags())

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cmd" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".jenkins.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.IsSet("context") {
			ctx := viper.GetString("context")
			if v := viper.Sub("servers." + ctx); v != nil {
				v.BindPFlags(rootCmd.Flags())
				opts.URL = v.GetString("url")
				opts.User = v.GetString("user")
				opts.Password = v.GetString("password")
			}
		}
	}
}
