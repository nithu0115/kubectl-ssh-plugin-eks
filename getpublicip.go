package main

import (
	"fmt"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func main() {
	awsRegionPtr := flag.String("region", "us-east-1", "a string")
	instanceIDPtr := flag.String("instanceid", "i-123332", "a string")
	flag.Parse()
	sess := session.Must(session.NewSession())
	//fmt.Println("Region:", *awsRegionPtr)
	ec2svc := ec2.New(sess, &aws.Config{Region: aws.String(*awsRegionPtr)})
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(*instanceIDPtr),
		},
	}
	result, err := ec2svc.DescribeInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	//fmt.Println(result)
	for idx, _ := range result.Reservations {
		for _, publicip := range result.Reservations[idx].Instances{
			fmt.Println (*publicip.PublicIpAddress)
		}
	}
}
