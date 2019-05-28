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
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stevensu1977/beej/pkg/config"
	"github.com/stevensu1977/beej/pkg/ec2"
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "scal up/down slave node ",
	Long:  `scal up/down slave node.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()

	},
}

// scaleCmd represents the scale command
var scaleUpCmd = &cobra.Command{
	Use:   "up",
	Short: "up slave node ",
	Long:  `up slave node.`,
	Run: func(cmd *cobra.Command, args []string) {
		//common flag setup
		setCommonFlag(cmd)

		viper.BindPFlag("count", cmd.Flags().Lookup("count"))
		count = aws.Int64(viper.GetInt64("count"))

		if *count == 0 {
			fmt.Println("[beej] scale up --count <num> !")
			return
		}

		cluster := ec2.ListInstance(profile, region)

		if cluster.Master.PublicDnsName == "" || cluster.Master.RawInstance == nil {
			fmt.Println("[beej] env not found, please check!")
			return
		}

		//create master instance
		instanceConfig := config.InstanceConfig{
			ImageId:        cluster.Master.RawInstance.ImageId,
			InstanceType:   cluster.Master.RawInstance.InstanceType,
			MinCount:       aws.Int64(*count),
			MaxCount:       aws.Int64(*count),
			KeyName:        cluster.Master.RawInstance.KeyName,
			SecurityGroups: []*string{cluster.Master.RawInstance.NetworkInterfaces[0].Groups[0].GroupName},
		}

		//create slave instance
		fmt.Printf("wait scale up %d slave node\n", *count)
		ec2.CreateInstance(profile, region, instanceConfig, config.TagSlave, count)

		data := [][]string{
			[]string{"master", cluster.Master.PublicDnsName, ""},
		}

		for _, v := range cluster.Slave {
			data = append(data, []string{
				"slave",
				"",
				v.PrivateIP,
			})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Type", "PublicDnsName", "PrivateIP"})

		for _, v := range data {
			table.Append(v)
		}
		table.Render() // Send output
	},
}

var scaleDownCmd = &cobra.Command{
	Use:   "down",
	Short: "down slave node ",
	Long:  `down slave node.`,
	Run: func(cmd *cobra.Command, args []string) {
		//common flag setup
		setCommonFlag(cmd)

		viper.BindPFlag("count", cmd.Flags().Lookup("count"))
		count = aws.Int64(viper.GetInt64("count"))

		if *count < 1 {
			fmt.Println("[beej] scale down --count <num> !")
			return
		}

		cluster := ec2.ListInstance(profile, region)

		if cluster.Master.PublicDnsName == "" {
			fmt.Println("[beej] env not found, please check!")
			return
		}

		if len(cluster.Slave) == 0 {
			fmt.Println("[beej] slave not found, please check!")
			return
		}

		slaves := int64(len(cluster.Slave))

		if *count > slaves {
			count = aws.Int64(slaves)
		}

		ids := []*string{}

		strconv.FormatInt(*count, 10)

		shutdown := int(*count)

		for i := 0; i < shutdown; i++ {
			ids = append(ids, cluster.Slave[i].RawInstance.InstanceId)
		}

		ec2.ShutdownInstance(profile, region, ids)

		cluster = ec2.ListInstance(profile, region)

		data := [][]string{
			[]string{"master", cluster.Master.PublicDnsName, ""},
		}

		for _, v := range cluster.Slave {
			data = append(data, []string{
				"slave",
				"",
				v.PrivateIP,
			})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Type", "PublicDnsName", "PrivateIP"})

		for _, v := range data {
			table.Append(v)
		}
		table.Render() // Send output
	},
}

func init() {

	scaleCmd.AddCommand(scaleUpCmd)
	scaleCmd.AddCommand(scaleDownCmd)

	rootCmd.AddCommand(scaleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scaleCmd.PersistentFlags().String("foo", "", "A help for foo")
	profile = scaleCmd.PersistentFlags().StringP("profile", "", "default", "profile name")
	region = scaleCmd.Flags().StringP("region", "r", "", "region name")

	scaleUpCmd.PersistentFlags().StringP("profile", "", "default", "profile name")
	scaleUpCmd.Flags().StringP("region", "r", "", "region name")
	scaleUpCmd.Flags().Int64P("count", "c", 0, "instance count")

	scaleDownCmd.PersistentFlags().StringP("profile", "", "default", "profile name")
	scaleDownCmd.Flags().StringP("region", "r", "", "region name")
	scaleDownCmd.Flags().Int64P("count", "c", 0, "instance count")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scaleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
