package ec2helper

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emetriq/gohelper/env"
	"github.com/thoas/go-funk"
)

// Client ...
type Client struct {
	Client                *ec2.EC2
	Session               *session.Session
	ec2Instances          []*string
	Ec2InstanceIps        []string
	Ec2InstancePrivateIps []string
	tags                  []*ec2.Tag
}

func convertToTagSlice(tags map[string]string) []*ec2.Tag {
	tagSlice := make([]*ec2.Tag, 0, len(tags))
	for key, value := range tags {
		tagSlice = append(tagSlice, &ec2.Tag{Key: aws.String(key), Value: aws.String(value)})
	}
	return tagSlice
}

// NewEC2Client ... ctor
// region is the region of the bucket you want to access
// leave it blank to use AWS_REGION environment variable
func NewEC2Client(region string, tags map[string]string) *Client {
	if region == "" {
		region = env.GetStrEnv("AWS_REGION", "eu-west-1")
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if env.GetStrEnv("AWS_SDK_LOAD_CONFIG", "") == "" {
		cred, err := GetCredentialsFromRole(sess)
		if err != nil {
			panic(err)
		}
		sess.Config.Credentials = cred
	}

	if err != nil {
		panic(err)
	}
	client := ec2.New(sess)

	ec2 := Client{
		Client:                client,
		Session:               sess,
		tags:                  convertToTagSlice(tags),
		ec2Instances:          make([]*string, 0),
		Ec2InstanceIps:        make([]string, 0),
		Ec2InstancePrivateIps: make([]string, 0),
	}
	return &ec2
}

// NewEC2ClientWithSession ... ctor
func NewEC2ClientWithSession(sess *session.Session, tags map[string]string) *Client {

	if sess == nil {
		return nil
	}

	client := ec2.New(sess)

	ec2 := Client{
		Client:  client,
		Session: sess,
		tags:    convertToTagSlice(tags),
	}
	return &ec2
}

//GetCredentialsFromRole get credentials from role when app is running in ec2
func GetCredentialsFromRole(sess *session.Session) (*credentials.Credentials, error) {
	roleProvider := &ec2rolecreds.EC2RoleProvider{
		Client: ec2metadata.New(sess),
	}
	creds := credentials.NewCredentials(roleProvider)
	if _, err := creds.Get(); err != nil {
		return nil, err
	}

	return creds, nil
}

//GetCredentialsFromRoleWithoutSession get credentials from role when app is running in ec2
func GetCredentialsFromRoleWithoutSession(region string) (*credentials.Credentials, error) {
	if region == "" {
		region = env.GetStrEnv("AWS_REGION", "eu-west-1")
	}

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})

	if err != nil {
		return nil, err
	}

	roleProvider := &ec2rolecreds.EC2RoleProvider{
		Client: ec2metadata.New(sess),
	}
	creds := credentials.NewCredentials(roleProvider)
	if _, err := creds.Get(); err != nil {
		return nil, err
	}

	return creds, nil
}

// CreateInstance creates an EC2 instance
func (c *Client) CreateInstance(instanceType string,
	iamProfile string,
	imageID string,
	subnetId string,
	securityGroupIDs []string,
	userData string,
	keyName string,
	minCount int64,
	maxCount int64) (*ec2.Reservation, error) {
	var tagSpec []*ec2.TagSpecification = nil
	if len(c.tags) > 0 {
		tagSpec = []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags:         c.tags,
			},
		}
	}
	return c.Client.RunInstances(&ec2.RunInstancesInput{
		UserData:          aws.String(userData),
		ImageId:           aws.String(imageID),
		InstanceType:      aws.String(instanceType),
		MinCount:          aws.Int64(minCount),
		MaxCount:          aws.Int64(maxCount),
		KeyName:           aws.String(keyName),
		SecurityGroupIds:  aws.StringSlice(securityGroupIDs),
		TagSpecifications: tagSpec,
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Name: aws.String(iamProfile),
		},
		SubnetId: aws.String(subnetId),
	})

}

// CreateInstanceAndWaitUntilReady creates a ec2 spot instance and waits until it is ready
func (c *Client) CreateInstanceAndWaitUntilReady(instanceName string,
	instanceType string,
	iamProfile string,
	imageID string,
	subnetId string,
	securityGroupIDs []string,
	spotPrice string,
	userData string,
	keyName string,
	minCount int64,
	maxCount int64) error {
	req, err := c.CreateInstance(instanceType, iamProfile, imageID, subnetId, securityGroupIDs, userData, keyName, minCount, maxCount)
	if err != nil {
		return err
	}

	for _, r := range req.Instances {
		c.ec2Instances = append(c.ec2Instances, r.InstanceId)
		c.Ec2InstanceIps = append(c.Ec2InstanceIps, *r.PublicIpAddress)
		c.Ec2InstancePrivateIps = append(c.Ec2InstanceIps, *r.PrivateIpAddress)
	}

	// create tags for the instance
	err = c.SetTags(instanceName, c.ec2Instances)

	if err != nil {
		c.TerminateAllEC2Instances()
		return err
	}
	err = c.Client.WaitUntilInstanceRunning(
		&ec2.DescribeInstancesInput{
			InstanceIds: c.ec2Instances,
		},
	)

	if err != nil {
		return err
	}

	if err != nil {
		c.TerminateAllEC2Instances()
		return err
	}

	return nil
}

// CreateSpotInstance creates a ec2 spot instance
func (c *Client) RequestSpotInstance(instanceType string, iamProfile string, imageID string, subnetId string, securityGroupIDs []string, spotPrice string, userData string, keyName string, instanceCount int64) (*ec2.RequestSpotInstancesOutput, error) {
	return c.Client.RequestSpotInstances(&ec2.RequestSpotInstancesInput{
		SpotPrice:     aws.String(spotPrice),
		InstanceCount: aws.Int64(instanceCount),
		LaunchSpecification: &ec2.RequestSpotLaunchSpecification{

			KeyName:      aws.String(keyName),
			UserData:     aws.String(userData),
			ImageId:      aws.String(imageID),
			InstanceType: aws.String(instanceType),
			IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
				Name: aws.String(iamProfile),
			},
			SubnetId:         aws.String(subnetId),
			SecurityGroupIds: aws.StringSlice(securityGroupIDs),
		},
	})
}

// SetTags set tags on ec2 instances with name
func (c *Client) SetTags(instanceName string, instanceIDs []*string) error {
	for id, instanceID := range instanceIDs {
		tags := []*ec2.Tag{{
			Key:   aws.String("Name"),
			Value: aws.String(instanceName + "-" + strconv.Itoa(id)),
		}}
		tags = append(tags, c.tags...)
		_, err := c.Client.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{instanceID},
			Tags:      tags,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateSpotInstanceAndWaitUntilReady creates a ec2 spot instance and waits until it is ready
func (c *Client) CreateSpotInstanceAndWaitUntilReady(instanceName string, instanceType string, iamProfile string, imageID string, subnetId string, securityGroupIDs []string, spotPrice string, userData string, keyName string, instanceCount int64) error {
	if instanceCount < 1 {
		return errors.New("instanceCount must be greater than 0")
	}
	req, err := c.RequestSpotInstance(instanceType, iamProfile, imageID, subnetId, securityGroupIDs, spotPrice, userData, keyName, instanceCount)
	if err != nil {
		return err
	}

	spotInstanceRequestIds := funk.Map(req.SpotInstanceRequests, func(x *ec2.SpotInstanceRequest) *string {
		return x.SpotInstanceRequestId
	}).([]*string)

	params := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: spotInstanceRequestIds,
	}

	err = c.Client.WaitUntilSpotInstanceRequestFulfilled(params)

	if err != nil {
		_, err = c.Client.CancelSpotInstanceRequests(&ec2.CancelSpotInstanceRequestsInput{
			SpotInstanceRequestIds: spotInstanceRequestIds,
		})
		return err
	}

	// Now we try to get the InstanceID of the instance we got
	requestDetails, err := c.Client.DescribeSpotInstanceRequests(params)
	if err != nil {
		return err
	}

	instanceIDs := make([]*string, 0, len(req.SpotInstanceRequests))
	// due to the waiter we can now safely assume all this data is available
	for _, r := range requestDetails.SpotInstanceRequests {
		c.ec2Instances = append(c.ec2Instances, r.InstanceId)
		instanceIDs = append(instanceIDs, r.InstanceId)
	}

	// create tags for the instance
	err = c.SetTags(instanceName, instanceIDs)
	if err != nil {
		c.TerminateAllEC2Instances()
		return err
	}

	err = c.Client.WaitUntilInstanceRunning(
		&ec2.DescribeInstancesInput{
			InstanceIds: instanceIDs,
		},
	)

	if err != nil {
		return err
	}

	des, err := c.Client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: c.ec2Instances,
	})

	if err != nil {
		c.TerminateAllEC2Instances()
		return err
	}

	c.Client.CreateTagsRequest(
		&ec2.CreateTagsInput{
			Resources: instanceIDs,
			Tags:      c.tags,
		})

	if err != nil {
		c.TerminateAllEC2Instances()
		return err
	}
	for _, r := range des.Reservations {
		for _, i := range r.Instances {
			if i.PublicIpAddress != nil {
				c.Ec2InstanceIps = append(c.Ec2InstanceIps, *i.PublicIpAddress)
			}
			if i.PrivateIpAddress != nil {
				c.Ec2InstancePrivateIps = append(c.Ec2InstancePrivateIps, *i.PrivateIpAddress)
			}
		}
	}

	return nil
}

// TerminateAllEC2Instances terminate all running EC2 instances created by this client
func (c *Client) TerminateAllEC2Instances() (*ec2.TerminateInstancesOutput, error) {
	if len(c.ec2Instances) == 0 {
		return nil, nil
	}

	req, err := c.Client.TerminateInstances(
		&ec2.TerminateInstancesInput{
			InstanceIds: c.ec2Instances,
		},
	)

	if err != nil {
		return nil, err
	}

	c.ec2Instances = make([]*string, 0)
	c.Ec2InstanceIps = make([]string, 0)
	c.Ec2InstancePrivateIps = make([]string, 0)
	return req, nil
}

// TerminateAllEC2InstancesAndWait terminate all running EC2 instances created by this client and wait until shutdown
func (c *Client) TerminateAllEC2InstancesAndWait() error {
	if len(c.ec2Instances) == 0 {
		return nil
	}
	_, err := c.Client.TerminateInstances(
		&ec2.TerminateInstancesInput{
			InstanceIds: c.ec2Instances,
		},
	)
	if err != nil {
		return err
	}

	err = c.Client.WaitUntilInstanceTerminated(
		&ec2.DescribeInstancesInput{
			InstanceIds: c.ec2Instances,
		},
	)

	return err
}

// GetInstancesInfos returns a list of instances infos matching the given filters
func (c *Client) GetInstancesInfos(params *ec2.DescribeInstancesInput, mapFunc interface{}) (interface{}, error) {
	resp, err := c.Client.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	if len(resp.Reservations) == 0 {
		return nil, nil
	}

	values := funk.Map(resp.Reservations[0].Instances, mapFunc)

	return values, nil
}

// GetAllRunningInstanceIDs returns all running instance IDs of region account
func (c *Client) GetAllRunningInstanceIDs() ([]*string, error) {
	res, err := c.GetInstancesInfos(&ec2.DescribeInstancesInput{}, func(x *ec2.Instance) *string {
		return x.InstanceId
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.([]*string), nil
}

// GetAllRunningInstanceIDsFilteredByName returns all running instances ids filtered by name
// name is the name of the instance as defined in the tag "Name"
// it is also possible to use wildcards in the name e.g. "*my-*" or "my-*"
func (c *Client) GetAllRunningInstanceIDsFilteredByName(name string) ([]*string, error) {
	res, err := c.GetInstancesInfos(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(name),
				},
			},
		},
	}, func(x *ec2.Instance) *string {
		return x.InstanceId
	})

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}
	return res.([]*string), nil
}

// GetAllRunningInstanceIDsFilteredByTagNameAndValue returns all running instances filtered by tag name and value
func (c *Client) GetAllRunningInstancesFilteredByTag(tag string, value string) ([]*ec2.Instance, error) {
	res, err := c.GetInstancesInfos(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:" + tag),
				Values: []*string{
					aws.String(value),
				},
			},
		},
	}, func(x *ec2.Instance) *ec2.Instance {
		return x
	})

	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return res.([]*ec2.Instance), nil
}

// GetAllRunningInstancesByName returns all running instances with the given name to the client
// name is the name of the instance as defined in the tag "Name"
// it is also possible to use wildcards in the name e.g. "*my-*" or "my-*"
func (c *Client) RestoreRunningInstancesByName(name string) error {
	res, err := c.GetAllRunningInstancesFilteredByTag("Name", name)
	if err != nil {
		return err
	}
	if res == nil {
		return nil
	}
	c.ec2Instances = make([]*string, 0, len(res))
	c.Ec2InstancePrivateIps = make([]string, 0, len(res))
	c.Ec2InstanceIps = make([]string, 0, len(res))
	for _, i := range res {
		c.ec2Instances = append(c.ec2Instances, i.InstanceId)
		if i.PublicIpAddress != nil {
			c.Ec2InstancePrivateIps = append(c.Ec2InstancePrivateIps, *i.PrivateIpAddress)
		}
		if i.PublicIpAddress != nil {
			c.Ec2InstanceIps = append(c.Ec2InstanceIps, *i.PublicIpAddress)
		}
	}
	return nil
}

// GetAllPrivateIPsOfStackByChildPrivateDNSName returns all private IPs of the stack of a given child private DNS name
func (c *Client) GetAllPrivateIPsOfStackByChildPrivateDNSName(childPrivateDNSName string) ([]*string, error) {
	stackID, err := c.GetStackIDByPrivateDNSName(childPrivateDNSName)
	if stackID == nil || err != nil {
		return nil, fmt.Errorf("stack with private DNS name %s not found", childPrivateDNSName)
	}
	return c.getAllIPsOfStackByStackID(stackID, func(reservation *ec2.Reservation) *string {
		if len(reservation.Instances) == 0 || reservation.Instances[0].PrivateIpAddress == nil || *reservation.Instances[0].PrivateDnsName == childPrivateDNSName {
			return nil
		}
		return reservation.Instances[0].PrivateIpAddress
	})
}

func (c *Client) getAllIPsOfStackByStackID(stackID *string, myMapFunc func(*ec2.Reservation) *string) ([]*string, error) {
	resultDescribeInstances, err := c.Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			// only gab instances with stack-id tag
			{
				Name:   aws.String("tag:aws:cloudformation:stack-id"),
				Values: []*string{stackID},
			},
		},
	})
	results := funk.Map(resultDescribeInstances.Reservations, myMapFunc).([]*string)
	results = funk.Filter(results, func(ip *string) bool {
		return ip != nil
	}).([]*string)
	return results, err
}

// GetAllPrivateIPsOfStackByStackID returns all private IPs of the stack of a given stack ID
func (c *Client) GetAllPrivateIPsOfStackByStackID(stackID *string) ([]*string, error) {
	return c.getAllIPsOfStackByStackID(stackID, func(reservation *ec2.Reservation) *string {
		if len(reservation.Instances) == 0 {
			return nil
		}
		return reservation.Instances[0].PrivateIpAddress
	})
}

// GetStackIDByPrivateDNSName returns the stack ID of a given private DNS name of a stack child
func (c *Client) GetStackIDByPrivateDNSName(dnsName string) (*string, error) {
	if dnsName == "" {
		return nil, errors.New("dnsName not set")
	}
	result, err := c.Client.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("private-dns-name"),
				Values: []*string{aws.String(dnsName)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Reservations) == 0 && len(result.Reservations[0].Instances) == 0 {
		return nil, errors.New("No instance found with private-dns-name " + dnsName)
	}
	stackIdTag := funk.Filter(result.Reservations[0].Instances[0].Tags, func(tag *ec2.Tag) bool {
		return *tag.Key == "aws:cloudformation:stack-id"
	}).([]*ec2.Tag)
	if stackIdTag != nil && len(stackIdTag) == 0 {
		return nil, errors.New("No stack-id tag found for instance with private-dns-name " + dnsName)
	}
	return stackIdTag[0].Value, nil
}
