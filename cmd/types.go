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
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stevensu1977/beej/pkg/config"
)

var region *string
var image *string
var instanceType *string
var key *string
var profile *string
var security *string
var count *int64
var testPlan *string

func buildSlaveList(servers []config.Server) string {
	list := []string{}
	for _, v := range servers {
		if !v.IsMaster {
			list = append(list, v.PrivateIP+":2099")
		}

	}
	return strings.Join(list, ",")
}

func getTS() string {
	return time.Now().Format("20060102150405")
}

func getJTLName(name string) string {

	return strings.Split(filepath.Base(name), ".")[0] + "-" + getTS() + ".jtl"
}

func setCommonFlag(cmd *cobra.Command) {

	viper.BindPFlag("profile", cmd.Flags().Lookup("profile"))
	profile = aws.String(viper.GetString("profile"))

	viper.BindPFlag("region", cmd.Flags().Lookup("region"))
	region = aws.String(viper.GetString("region"))
}
