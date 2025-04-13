package aws

import (
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"sync"
	"time"

	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"golang.org/x/net/context"
)

type awsAPI struct {
	app *model.App
	ec2 *ec2.Client
	ce  *costexplorer.Client

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

	ce := costexplorer.NewFromConfig(cfg)
	ec2 := ec2.NewFromConfig(cfg)
	aws := &awsAPI{
		app: app,
		ec2: ec2,
		ce:  ce,
	}

	_, err = aws.getInstanceStatus()
	if err != nil {
		logging.Error("Failed to get instance status", "error", err)
		return nil, err
	}

	return aws, nil
}

func (aws *awsAPI) getInstanceStatus() (InstanceStatus, error) {
	filterName := "tag:Project"
	input := &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   &filterName,
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

	resp, err := aws.ec2.DescribeInstances(context.Background(), input)
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

	_, err = aws.ec2.StartInstances(context.Background(), input)
	if err != nil {
		logging.Error("Failed to start instance", "error", err)
		return err
	}

	logging.Info("Instance started successfully")
	return nil
}

func (aws *awsAPI) GetCostAllocationTag(tagKey string) (*ceTypes.CostAllocationTag, error) {
	input := &costexplorer.ListCostAllocationTagsInput{
		TagKeys: []string{tagKey},
	}

	resp, err := aws.ce.ListCostAllocationTags(context.Background(), input)
	if err != nil {
		logging.Error("Failed to list cost allocation tags", "error", err)
		return nil, err
	}

	for _, tag := range resp.CostAllocationTags {
		if tag.TagKey == &tagKey {
			return &tag, nil
		}
	}
	return nil, fmt.Errorf("tag not found")
}

func (aws *awsAPI) ActivateCostAllocationTag(tagKey string) error {
	input := &costexplorer.UpdateCostAllocationTagsStatusInput{
		CostAllocationTagsStatus: []ceTypes.CostAllocationTagStatusEntry{
			{
				TagKey: &tagKey,
				Status: ceTypes.CostAllocationTagStatusActive,
			},
		},
	}

	resp, err := aws.ce.UpdateCostAllocationTagsStatus(context.Background(), input)
	if err != nil {
		logging.Error("Failed to create cost allocation tag", "tag", tagKey, "error", err)
		return err
	}

	if len(resp.Errors) > 0 {
		logging.Error("Failed to create cost allocation tag")
		for _, err := range resp.Errors {
			logging.Error("Error creating cost allocation tag", "tag", err.TagKey, "error", err.Message, "code", err.Code)
		}
		return fmt.Errorf("failed to create cost allocation tag: %s", tagKey)
	}

	logging.Info("Cost allocation tag created successfully")
	return nil
}

func (aws *awsAPI) BackfillCostAllocationTag(tagKey string, backfillFrom time.Time) error {
	backfillFrom = time.Date(backfillFrom.Year(), backfillFrom.Month(), 1, 0, 0, 0, 0, time.UTC)
	if backfillFrom.After(time.Now()) {
		return fmt.Errorf("backfillFrom date cannot be in the future")
	}

	backfillString := backfillFrom.Format("2006-01-02T15:04:05Z")
	logging.Info("Starting cost allocation tag backfill",
		"tag", tagKey,
		"backfillFrom", backfillString,
	)

	input := &costexplorer.StartCostAllocationTagBackfillInput{
		BackfillFrom: &backfillString,
	}

	resp, err := aws.ce.StartCostAllocationTagBackfill(context.Background(), input)
	if err != nil {
		logging.Error("Failed to backfill cost allocation tag", "tag", tagKey, "error", err)
		return err
	}
	logging.Info("Cost allocation tag backfill started successfully",
		"tag", tagKey,
		"status", resp.BackfillRequest.BackfillStatus,
		"backfillFrom", backfillFrom,
		"lastUpdate", resp.BackfillRequest.LastUpdatedAt,
		"completedAt", resp.BackfillRequest.CompletedAt,
		"requestedAt", resp.BackfillRequest.RequestedAt,
	)
	return nil
}
