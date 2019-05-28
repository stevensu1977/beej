package config

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//DefaultHome user home
const DefaultHome = "/home/ec2-user"

const TagMaster = "beepj-master"
const TagSlave = "beepj"

//T2Micro instance type
const T2Micro = "t2.micro"

//T2Small instance type
const T2Small = "t2.small"

//DefaultUser is ec2 instance , this specity user name , only use Amazon AMI as base image
const DefaultUser = "ec2-user"

//JmeterCluster is beej ec2 instance warpper
type JmeterCluster struct {
	Master Server
	Slave  []Server
}

//Server is ec2 instance warpper struct
type Server struct {
	PublicDnsName string
	PrivateIP     string
	IsMaster      bool
	RawInstance   *ec2.Instance
}

type InstanceConfig struct {
	ImageId        *string
	InstanceType   *string
	MinCount       *int64
	MaxCount       *int64
	KeyName        *string
	SecurityGroups []*string
}

//LoadConfig is  func  use  ~/.aws/config
func LoadConfig(profile, region *string) *session.Session {
	//use profile and region flag
	if profile != nil && *profile == "" {
		os.Setenv("AWS_PROFILE", *profile)
	}

	if *region != "" {
		os.Setenv("AWS_REGION", *region)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return sess
}

func IsMaster(tags []*ec2.Tag) bool {
	return containsTag(tags, TagMaster)
}

func IsSlave(tags []*ec2.Tag) bool {
	return containsTag(tags, TagSlave)
}

func containsTag(tags []*ec2.Tag, value string) bool {
	for _, v := range tags {
		if *v.Value == value {
			return true
		}
	}
	return false
}
