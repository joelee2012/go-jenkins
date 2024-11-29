/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/joelee2012/go-jenkins"
	"github.com/spf13/cobra"
)

// getJobCmd represents the job command
var getJobCmd = &cobra.Command{
	Use:   "job",
	Short: "A brief description of your command",
	Args:  cobra.MaximumNArgs(1),
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return opts.GetJob(args)
	},
}

func init() {
	getCmd.AddCommand(getJobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	getJobCmd.Flags().IntVarP(&opts.Depth, "depth", "d", 0, "list job depth")
	getJobCmd.MarkFlagsMutuallyExclusive("depth", "output")
}

func (o *JenkinsOpts) GetJob(args []string) error {
	if o.URL == "" {
		return fmt.Errorf("no jenkins url set")
	}
	client := jenkins.New(o.URL, o.User, o.Password)
	var names []string
	if len(args) != 0 {
		for _, name := range args {
			if o.Folder == "" {
				names = append(names, name)
			} else {
				names = append(names, fmt.Sprintf("%s/%s", o.Folder, name))
			}
		}
		for _, name := range names {
			printJobs(client, name, o.Output, o.Depth)
		}
	} else {
		printJobs(client, o.Folder, o.Output, o.Depth)

	}
	return nil
}

func printJobs(client *jenkins.Jenkins, folder, output string, depth int) {
	job, err := client.GetJob(folder)
	cobra.CheckErr(err)
	showConfigOrName(job, output)
	jobs, err := job.List(depth)
	cobra.CheckErr(err)
	for _, job := range jobs {
		showConfigOrName(job, output)
	}
}

func showConfigOrName(job *jenkins.Job, output string) {
	if output == "xml" {
		fmt.Println(job.Configure())
	} else {
		fmt.Println(job)
	}
}
