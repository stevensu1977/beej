package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/stevensu1977/beej/pkg/config"
)

//ListInstance get all region string
func ListInstance(profile, region *string, tags ...string) config.JmeterCluster {

	sess := config.LoadConfig(profile, region)

	// Create EC2 service client
	svc := ec2.New(sess)
	_tagFilter := []*string{aws.String(config.TagMaster), aws.String(config.TagSlave)}
	if len(tags) > 0 {
		_tagFilter = []*string{}
		for idx := range tags {
			_tagFilter = append(_tagFilter, aws.String(tags[idx]))
		}
	}

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:Name"),
				Values: _tagFilter,
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
	var cluster = config.JmeterCluster{
		Master: config.Server{},
		Slave:  []config.Server{},
	}

	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return cluster
	}

	for idx := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if config.IsMaster(inst.Tags) {
				cluster.Master = config.Server{
					PublicDnsName: *inst.PublicDnsName,
					PrivateIP:     *inst.PrivateIpAddress,
					IsMaster:      true,
					RawInstance:   inst,
				}

			} else if config.IsSlave(inst.Tags) {
				cluster.Slave = append(cluster.Slave, config.Server{
					PrivateIP:   *inst.PrivateIpAddress,
					IsMaster:    false,
					RawInstance: inst,
				})

			} else {
				fmt.Println("Other Tag")
			}

			ids = append(ids, inst.InstanceId)
		}
	}

	return cluster
}

//CreateInstance create ec2 instance
func CreateInstance(profile, region *string, instance config.InstanceConfig, tag string, count *int64) {

	// Init aws session from profile , region argument and ~/.aws/config
	sess := config.LoadConfig(profile, region)

	// Create EC2 service client
	svc := ec2.New(sess)

	fmt.Printf("[beej] %s  waitig creating \n", tag)
	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:        instance.ImageId,
		InstanceType:   instance.InstanceType,
		MinCount:       instance.MinCount,
		MaxCount:       instance.MaxCount,
		KeyName:        instance.KeyName,
		SecurityGroups: instance.SecurityGroups,
	})

	if err != nil {
		fmt.Println("Could not create instance ", err)
		return
	}

	//create tag for every instance
	for _, instance := range runResult.Instances {
		fmt.Printf("[beej] created instance [%s] success , private ip : %s \n", *instance.InstanceId, *instance.PrivateIpAddress)
		_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{instance.InstanceId},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(tag),
				},
			},
		})
		if errtag != nil {
			log.Println("Could not create tags for instance", *instance.InstanceId, errtag)
			return
		}

	}

}

//ShutdownInstance is terminate
func ShutdownInstance(profile, region *string, ids []*string) {

	// Init aws session from profile , region argument and ~/.aws/config
	sess := config.LoadConfig(profile, region)

	// Create EC2 service client
	svc := ec2.New(sess)

	fmt.Printf("[beej] %v  waitig terminate \n", ids)
	// Specify the details of the instance that you want to create.
	result, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: ids,
	})

	if err != nil {
		fmt.Println("Could not stop instance ", err)
		return
	}

	if len(result.TerminatingInstances) > 0 {
		fmt.Println("Wait terminate instances: ")
		for idx := range result.TerminatingInstances {
			fmt.Printf("=> %s  \n", *result.TerminatingInstances[idx].InstanceId)
		}
	}

}
