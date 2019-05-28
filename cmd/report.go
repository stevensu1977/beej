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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stevensu1977/beej/pkg/ec2"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		//common flag setup
		setCommonFlag(cmd)

		viper.BindPFlag("key", cmd.Flags().Lookup("key"))
		key = aws.String(viper.GetString("key"))

		cluster := ec2.ListInstance(profile, region)

		if len(cluster.Slave) == 0 {
			fmt.Println("Not found beej env, please check! ")
			os.Exit(0)
		}

		if *key == "" {
			fmt.Println("private key file must be required!")
			os.Exit(0)
		}
		fmt.Println(fmt.Sprintf("[beej] Report url: http://%s", cluster.Master.PublicDnsName))
		session, err := connect("ec2-user", *key, cluster.Master.PublicDnsName, 22)
		if err != nil {
			panic(err)
		}

		defer session.Close()
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr

		session.Run("cd dashboard \n ls -d */")

	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")
	profile = reportCmd.PersistentFlags().StringP("profile", "", "default", "profile name")
	region = reportCmd.Flags().StringP("region", "r", "", "region name")
	key = reportCmd.Flags().StringP("key", "k", "", "keypair")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
