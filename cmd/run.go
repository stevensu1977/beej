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
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stevensu1977/beej/pkg/config"
	"github.com/stevensu1977/beej/pkg/ec2"
	"golang.org/x/crypto/ssh"
)

var manual *bool

func connect(user string, keyfile string, host string, port int) (*ssh.Session, error) {

	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	var (
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
	)

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	return session, nil
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run jmeter test plan",
	Long:  "run jmeter test plan",
	Run: func(cmd *cobra.Command, args []string) {

		//common flag setup
		setCommonFlag(cmd)

		viper.BindPFlag("key", cmd.Flags().Lookup("key"))
		key = aws.String(viper.GetString("key"))

		//check testPlan
		if *testPlan == "" {
			fmt.Println("Test plan file must be required!")
			os.Exit(0)
		}
		//check private key
		if *key == "" {
			fmt.Println("private key file must be required!")
			os.Exit(0)
		}

		runTestCase()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	key = runCmd.Flags().StringP("key", "", "", "private key full path")
	testPlan = runCmd.Flags().StringP("testplan", "", "", "jmeter test plan file etc. TestPlan01.jmx")
	region = runCmd.Flags().StringP("region", "r", "", "region name")
	profile = runCmd.Flags().StringP("profile", "", "default", "profile name")
	manual = runCmd.Flags().BoolP("manual", "", false, "only create test script , not execute")

}

func runTestCase() {

	//load all intance include master and slave
	cluster := ec2.ListInstance(profile, region)
	if len(cluster.Slave) == 0 {
		fmt.Println("beej env not exits, please check")
		os.Exit(0)
	}

	jtlName := getJTLName(*testPlan)
	fmt.Println(jtlName)

	fmt.Printf("Step1 upload %s test plan to beej master %s \n", *testPlan, cluster.Master.PublicDnsName)
	fmt.Println(*key)
	sftpClient, err := sftpconnect("ec2-user", *key, cluster.Master.PublicDnsName, 22)
	if err != nil {
		panic(err)
	}
	fmt.Println("already connect")
	err = scpCopy(sftpClient, *testPlan, config.DefaultHome)
	if err != nil {
		panic(err)
	}
	fmt.Println("already connect")
	file, err := os.Create(strings.Split(jtlName, ".")[0] + ".sh")
	if err != nil {
		panic(err)
	}
	file.WriteString("#!/bin/sh \n")
	file.WriteString(fmt.Sprintf("/usr/local/apache-jmeter-5.1.1/bin/jmeter -n -t ~/%s -R %s  -l %s \n  echo '\nCovert jtl to report\n' \nif [ -d 'dashboard/%s' ]; then \n rm -rf dashboard/%s \nfi \n /usr/local/apache-jmeter-5.1.1/bin/jmeter -g %s -o dashboard/%s", *testPlan, buildSlaveList(cluster.Slave), jtlName, strings.Split(jtlName, ".")[0], strings.Split(jtlName, ".")[0], jtlName, strings.Split(jtlName, ".")[0]))
	defer file.Close()

	err = scpCopy(sftpClient, strings.Split(jtlName, ".")[0]+".sh", config.DefaultHome)
	if err != nil {
		panic(err)
	}

	defer sftpClient.Close()

	fmt.Printf("\n[beej]  you can ssh %s ,  execute %s \n", cluster.Master.PublicDnsName, strings.Split(jtlName, ".")[0]+".sh")

	if *manual == false {
		fmt.Println("")
		fmt.Println("#############Begin run test plan#############")
		fmt.Println(cluster.Master.PublicDnsName)
		fmt.Println(buildSlaveList(cluster.Slave))

		session, err := connect("ec2-user", *key, cluster.Master.PublicDnsName, 22)
		if err != nil {
			panic(err)
		}

		defer session.Close()
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr

		fmt.Println("")
		fmt.Println("")
		session.Run(fmt.Sprintf("/usr/local/apache-jmeter-5.1.1/bin/jmeter -n -t ~/%s -R %s  -l %s \n  echo '\nCovert jtl to report\n' \n /usr/local/apache-jmeter-5.1.1/bin/jmeter -g %s -o dashboard/%s", *testPlan, buildSlaveList(cluster.Slave), jtlName, jtlName, strings.Split(jtlName, ".")[0]))
		fmt.Println("")
		fmt.Println("#############Report#############")
		fmt.Println(fmt.Sprintf("http://%s/%s", cluster.Master.PublicDnsName, strings.Split(jtlName, ".")[0]))

	}

	os.Remove(strings.Split(jtlName, ".")[0] + ".sh")
}

func sftpconnect(user, keyfile, host string, port int) (*sftp.Client, error) {
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	var (
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		sftpClient   *sftp.Client
	)

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(client); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

func scpCopy(sftpClient *sftp.Client, localFilePath, remoteDir string) error {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		log.Println("scpCopy:", err)
		return err
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		log.Println("scpCopy:", err)
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[0:n])
	}
	return nil
}
