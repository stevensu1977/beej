// Copyright © 2019 NAME Steven Su suwei007@gmail.com
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
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/stevensu1977/beej/pkg/ec2"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list beej ec2 instance",
	Long:  "list beej ec2 instance",
	Run: func(cmd *cobra.Command, args []string) {

		//common flag setup
		setCommonFlag(cmd)

		cluster := ec2.ListInstance(profile, region)

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
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")
	profile = listCmd.PersistentFlags().StringP("profile", "", "default", "profile name")
	region = listCmd.Flags().StringP("region", "r", "", "region name")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
