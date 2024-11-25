/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/joelee2012/go-jenkins"
	"github.com/spf13/cobra"
)

var depth int

// jobCmd represents the job command
var jobCmd = &cobra.Command{
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
	getCmd.AddCommand(jobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	jobCmd.Flags().IntVarP(&opts.Depth, "depth", "d", 0, "list job depth")
}

func (o *JenkinsOpts) GetJob(args []string) error {
	if o.URL == "" {
		return fmt.Errorf("no jenkins url set")
	}
	client := jenkins.New(o.URL, o.User, o.Password)
	var job *jenkins.Job
	var err error
	if o.Path == "." {
		if len(args) == 0 {
			jobs, err := client.ListJobs(depth)
			cobra.CheckErr(err)
			for _, job := range jobs {
				fmt.Println(job)
			}
		} else {
			job, err = client.GetJob(args[0])
		}

	} else {
		if len(args) == 0 {
			job, err = client.GetJob(o.Path)
			cobra.CheckErr(err)
			jobs, err := job.List(depth)
			cobra.CheckErr(err)
			for _, job := range jobs {
				fmt.Println(job)
			}
		} else {
			job, err = client.GetJob(fmt.Sprintf("%s/%s", o.Path, args[0]))
		}
	}
	if err != nil {
		return err
	}
	fmt.Println(job)
	return nil

}
