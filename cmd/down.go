// Copyright © 2019 Steven Su suwei007@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stevensu1977/beej/pkg/config"
)

var tag *[]string

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "shutdown all ec2 instance",
	Long:  "shutdown all ec2 instance",
	Run: func(cmd *cobra.Command, args []string) {

		//use viper get common profile
		viper.BindPFlag("profile", cmd.Flags().Lookup("profile"))
		profile = aws.String(viper.GetString("profile"))

		viper.BindPFlag("region", cmd.Flags().Lookup("region"))
		region = aws.String(viper.GetString("region"))

		stopInstance()
	},
}

func init() {
	rootCmd.AddCommand(downCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// downCmd.PersistentFlags().String("foo", "", "A help for foo")
	key = downCmd.Flags().StringP("key", "k", "", "keypair")
	region = downCmd.Flags().StringP("region", "r", "cn-northwest-1", "region name")
	profile = downCmd.Flags().StringP("profile", "", "default", "profile name")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// downCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func stopInstance() {

	// Init aws session from profile , region argument and ~/.aws/config
	sess := config.LoadConfig(profile, region)

	// Create EC2 service client
	svc := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: []*string{aws.String(config.TagMaster), aws.String(config.TagSlave)},
			},
			{
				Name:   aws.String("instance-state-name"),
				Values: []*string{aws.String("running"), aws.String("pending")},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		panic(err)
	}

	var ids = []*string{}

	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Printf("[beej]  terminated instance [%s]  \n", *inst.InstanceId)
			ids = append(ids, inst.InstanceId)
		}
	}

	if len(ids) == 0 {
		fmt.Println("Not found beej env.")
		os.Exit(0)
	}
	input := &ec2.TerminateInstancesInput{
		InstanceIds: ids,
		DryRun:      aws.Bool(true),
	}

	_, err = svc.TerminateInstances(input)
	awsErr, ok := err.(awserr.Error)
	if ok && awsErr.Code() == "DryRunOperation" {
		input.DryRun = aws.Bool(false)
		_, err = svc.TerminateInstances(input)
		if err != nil {
			fmt.Println("Error", err)
		} else {
			fmt.Println("Down beej env success!")
		}
	} else {
		fmt.Println("Error", err)
	}

}
