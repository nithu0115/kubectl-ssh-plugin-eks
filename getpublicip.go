package main

import (
	"fmt"
	"flag"
	"strings"
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

	subnetID, RouteTableResponse, err, PrivateDnsName, PublicDnsName := descrInstance(ec2svc, result)
	if err != nil {
		fmt.Println(err)
	}

	isPublic, err := isSubnetPublic(RouteTableResponse, subnetID)
	if err != nil {
		fmt.Println(err)
	}
	if !isPublic {
		fmt.Println("PrivateDnsName",PrivateDnsName)
	} else {
		fmt.Println("PublicDnsName",PublicDnsName)
	}
}

func descrInstance(ec2svc *ec2.EC2, result *ec2.DescribeInstancesOutput) (subnetID string, RouteTableResponse []*ec2.RouteTable ,err error, PrivateDnsName string, PublicDnsName string ) {
	for idx, _ := range result.Reservations {
		for _, isSubnetPublic := range result.Reservations[idx].Instances{
			subnetID := *isSubnetPublic.SubnetId
			subnetRequest := &ec2.DescribeSubnetsInput{
				SubnetIds: []*string{
					&subnetID,
				},
			}
			subnetResponse, err := ec2svc.DescribeSubnets(subnetRequest)
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
				//return
			}

			vpcId := subnetResponse.Subnets[0].VpcId
			//vpcId := subnetResponse.Subnets

			routeTableRequest := &ec2.DescribeRouteTablesInput{
				Filters: []*ec2.Filter{
					  {
						Name: aws.String("vpc-id"),
						Values: []*string{
						     aws.String(*vpcId),
					      },
				      },
				      /*{
						Name: aws.String("association.subnet-id"),
						Values: []*string{
						     aws.String(subnetID),
					      },
				      },*/
				  },
			   }
			routeTableresponse, err := ec2svc.DescribeRouteTables(routeTableRequest)
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
				//return
			}
			RouteTableResponse := routeTableresponse.RouteTables
			return subnetID, RouteTableResponse, err, *isSubnetPublic.PrivateDnsName, *isSubnetPublic.PublicIpAddress
		}
	}
	return
}

func isSubnetPublic(RouteTableResponse []*ec2.RouteTable, subnetID string) (bool, error) {
	var subnetTable *ec2.RouteTable
	for _, table := range RouteTableResponse {
		for _, assoc := range table.Associations {
			if aws.StringValue(assoc.SubnetId) == subnetID {
				subnetTable = table
				break
			}
		}
	}

	if subnetTable == nil {
		// If there is no explicit association, the subnet will be implicitly
		// associated with the VPC's main routing table.
		for _, table := range RouteTableResponse {
			for _, assoc := range table.Associations {
				if aws.BoolValue(assoc.Main) == true {
					//fmt.Println("Assuming implicit use of main routing table %s for %s",
					//	aws.StringValue(table.RouteTableId), subnetID)
					subnetTable = table
					break
				}
			}
		}
	}

	if subnetTable == nil {
		return false, fmt.Errorf("Could not locate routing table for subnet %s", subnetID)
	}

	for _, route := range subnetTable.Routes {
		// There is no direct way in the AWS API to determine if a subnet is public or private.
		// A public subnet is one which has an internet gateway route
		// we look for the gatewayId and make sure it has the prefix of igw to differentiate
		// from the default in-subnet route which is called "local"
		// or other virtual gateway (starting with vgv)
		// or vpc peering connections (starting with pcx).
		if strings.HasPrefix(aws.StringValue(route.GatewayId), "igw") {
			return true, nil
		}
	}

	return false, nil
}
