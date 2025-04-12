package aws

import (
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"
	"time"

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
	state      ec2Types.InstanceStateName
}

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

	_, err = aws.getInstanceStatus()
	if err != nil {
		logging.Error("Failed to get instance status", "error", err)
		return nil, err
	}

	return aws, nil
}

func (aws *awsAPI) getInstanceStatus() (InstanceStatus, error) {
	filter1Name := "tag:Project"
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   &filter1Name,
				Values: []string{aws.app.Config.GetAWSProjectName()},
			},
		},
	}

	instanceStatus := InstanceStatus{
		InstanceId: "",
		isOnline:   false,
		PublicIP:   nil,
		state:      "",
	}

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
	instanceStatus.state = instance.State.Name
	if instance.PublicIpAddress != nil {
		instanceStatus.PublicIP = instance.PublicIpAddress
	}

	logging.Debug("Instance status", "instanceId", instanceStatus.InstanceId, "isOnline", instanceStatus.isOnline, "publicIP", instanceStatus.PublicIP)
	return instanceStatus, nil
}

func (aws *awsAPI) getInstanceId() (string, error) {
	aws.mutex.Lock()
	instanceId := aws._instanceId
	aws.mutex.Unlock()
	if instanceId != nil {
		return *instanceId, nil
	}

	instanceStatus, err := aws.getInstanceStatus()
	if err != nil {
		return "", err
	}

	if instanceStatus.InstanceId == "" {
		return "", fmt.Errorf("instance ID is empty")
	}

	return instanceStatus.InstanceId, nil
}

func (aws *awsAPI) GetInstanceState() (ec2Types.InstanceStateName, error) {
	instanceStatus, err := aws.getInstanceStatus()
	if err != nil {
		logging.Error("Failed to get instance status", "error", err)
		return "", err
	}
	if instanceStatus.state == "" {
		return "", fmt.Errorf("instance state is empty")
	}
	return instanceStatus.state, nil
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
	instanceId, err := aws.getInstanceId()
	if err != nil {
		logging.Error("Failed to get instance ID", "error", err)
		return err
	}

InstanceCheck:
	for i := 0; i < 12; i++ {
		instanceStatus, err := aws.getInstanceStatus()
		if err != nil {
			logging.Error("Failed to get instance status", "error", err)
			goto waitAndContinue
		}
		switch instanceStatus.state {
		case ec2Types.InstanceStateNameRunning:
			logging.Info("Instance is already running")
			return nil
		case ec2Types.InstanceStateNamePending:
			goto waitAndContinue
		case ec2Types.InstanceStateNameStopping:
			goto waitAndContinue
		case ec2Types.InstanceStateNameStopped:
			break InstanceCheck
		case ec2Types.InstanceStateNameShuttingDown:
			goto waitAndContinue
		case ec2Types.InstanceStateNameTerminated:
			return fmt.Errorf("instance is terminated, cannot start it")
		default:
			logging.Error("Unknown instance state", "state", instanceStatus.state)
			continue
		}

	waitAndContinue:
		time.Sleep(5 * time.Second)
	}

	input := &ec2.StartInstancesInput{
		InstanceIds: []string{instanceId},
	}

	_, err = aws.client.StartInstances(context.Background(), input)
	if err != nil {
		logging.Error("Failed to start instance", "error", err)
		return err
	}

	logging.Info("Instance started successfully")
	return nil
}
