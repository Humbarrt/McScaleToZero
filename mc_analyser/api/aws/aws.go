package aws

import (
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"

	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"golang.org/x/net/context"
)

type awsAPI struct {
	app    *model.App
	client *ec2.Client

	// protected
	mutex       sync.Mutex
	_instanceId *string
}

type InstanceStatus struct {
	InstanceId string
	isOnline   bool
	PublicIP   *string
}

/*
StartInstance() error
StopInstance() error
IsInstanceOnline() (bool, error)
GetInstanceIP() (string, error)
*/

func NewAWSAPI(app *model.App) (*awsAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithAppID("mc_analyser"), config.WithRetryMaxAttempts(2))
	if err != nil {
		logging.Error("Unable to load SDK config", "error", err)
		return nil, err
	}

	client := ec2.NewFromConfig(cfg)
	aws := &awsAPI{
		app:    app,
		client: client,
	}

	// Check if the instance exists
	_, err = aws.getInstanceStatus()
	if err != nil {
		logging.Error("Failed to get instance status", "error", err)
		return nil, err
	}

	return aws, nil
}

func (aws *awsAPI) getInstanceStatus() (InstanceStatus, error) {
	filter1Name := "tag:Name"
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   &filter1Name,
				Values: []string{aws.app.Config.AWS_INSTANCE_NAME},
			},
		},
	}

	instanceStatus := InstanceStatus{
		InstanceId: "",
		isOnline:   false,
		PublicIP:   nil,
	}

	// Call the DescribeInstances API
	resp, err := aws.client.DescribeInstances(context.Background(), input)
	if err != nil {
		logging.Error("Faild to describe the instance", "error", err)
		return instanceStatus, err
	}

	if len(resp.Reservations) == 0 {
		return instanceStatus, fmt.Errorf("no reservations found")
	}

	if len(resp.Reservations[0].Instances) == 0 {
		return instanceStatus, fmt.Errorf("no instances found")
	}

	if len(resp.Reservations[0].Instances) > 1 {
		return instanceStatus, fmt.Errorf("more than one matching instance found")
	}

	instance := resp.Reservations[0].Instances[0]
	if instance.InstanceId == nil {
		return instanceStatus, fmt.Errorf("instance ID is nil")
	}
	if instance.State == nil {
		return instanceStatus, fmt.Errorf("instance state is nil")
	}

	aws.mutex.Lock()
	aws._instanceId = instance.InstanceId
	aws.mutex.Unlock()

	instanceStatus.InstanceId = *instance.InstanceId
	instanceStatus.isOnline = instance.State.Name == ec2Types.InstanceStateNameRunning
	if instance.PublicIpAddress != nil {
		instanceStatus.PublicIP = instance.PublicIpAddress
	}

	logging.Debug("Instance status", "instanceId", instanceStatus.InstanceId, "isOnline", instanceStatus.isOnline, "publicIP", instanceStatus.PublicIP)
	return instanceStatus, nil
}

func (aws *awsAPI) IsInstanceOnline() (bool, error) {
	instanceStatus, err := aws.getInstanceStatus()
	if err != nil {
		return false, err
	}

	return instanceStatus.isOnline, nil
}

func (aws *awsAPI) GetInstanceIP() (string, error) {
	instanceStatus, err := aws.getInstanceStatus()
	if err != nil {
		return "", err
	}

	if instanceStatus.PublicIP == nil {
		return "", fmt.Errorf("instance does not have a public IP address")
	}

	return *instanceStatus.PublicIP, nil
}

func (aws *awsAPI) StartInstance() error {
	aws.mutex.Lock()
	if aws._instanceId == nil {
		return fmt.Errorf("instance ID is nil")
	}
	aws.mutex.Unlock()

	input := &ec2.StartInstancesInput{
		InstanceIds: []string{*aws._instanceId},
	}

	_, err := aws.client.StartInstances(context.Background(), input)
	if err != nil {
		logging.Error("Failed to start instance", "error", err)
		return err
	}

	logging.Info("Instance started successfully")
	return nil
}

func (aws *awsAPI) StopInstance() error {
	aws.mutex.Lock()
	if aws._instanceId == nil {
		return fmt.Errorf("instance ID is nil")
	}
	aws.mutex.Unlock()

	input := &ec2.StopInstancesInput{
		InstanceIds: []string{*aws._instanceId},
	}

	_, err := aws.client.StopInstances(context.Background(), input)
	if err != nil {
		logging.Error("Failed to stop instance", "error", err)
		return err
	}

	logging.Info("Instance stopped successfully")
	return nil
}
