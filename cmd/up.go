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
	"os"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/stevensu1977/beej/pkg/config"
	"github.com/stevensu1977/beej/pkg/ec2"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "provisining ec2 instance for jmeter test",
	Long: `provisining ec2 instance for jmeter test 

	AMI id:
	      cn-northwest-1  ami-043f3420c7f54a469
	      cn-north-1      ami-0af811842a4af1f9a
	      ap-northeast-1  ami-002e780c728a7553f
	`,

	Run: func(cmd *cobra.Command, args []string) {

		//common flag setup
		setCommonFlag(cmd)

		viper.BindPFlag("key", cmd.Flags().Lookup("key"))
		key = aws.String(viper.GetString("key"))

		//check keypair
		if *key == "" {
			fmt.Println("keypair must be required!")
			os.Exit(0)
		}

		//check AMI id
		if *image == "" {
			fmt.Println("AMI id must be required!")
			os.Exit(0)
		}

		//check instance type
		if *instanceType == config.T2Micro {
			panic(fmt.Errorf("JMeter heap need 1g RAM, t2.micro can't be load"))
		}

		cluster := ec2.ListInstance(profile, region)

		if len(cluster.Slave) > 0 {
			fmt.Println("Already have beej env, please down first")
			os.Exit(0)
		}

		//create master instance
		instanceConfig := config.InstanceConfig{
			ImageId:        aws.String(*image),
			InstanceType:   aws.String(*instanceType),
			MinCount:       aws.Int64(*count),
			MaxCount:       aws.Int64(*count),
			KeyName:        aws.String(*key),
			SecurityGroups: []*string{security},
		}

		ec2.CreateInstance(profile, region, instanceConfig, config.TagMaster, aws.Int64(1))

		//create slave instance
		ec2.CreateInstance(profile, region, instanceConfig, config.TagSlave, count)
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	//common flag
	region = upCmd.Flags().StringP("region", "r", "", "region name")
	profile = upCmd.PersistentFlags().StringP("profile", "", "default", "profile name")

	//run flag
	count = upCmd.Flags().Int64P("count", "c", 1, "instance count")
	key = upCmd.Flags().StringP("key", "", "", "keypair")
	image = upCmd.Flags().StringP("image", "i", "", "image id")
	instanceType = upCmd.Flags().StringP("type", "t", config.T2Small, "instance type ")
	security = upCmd.Flags().StringP("sg", "", "default", "security group")

}
