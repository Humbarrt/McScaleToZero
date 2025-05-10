package aws

import (
	"fmt"
	"mc_manager/logging"
	"mc_manager/model"
	"strings"
	"sync"
	"time"

	cwlTypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/shopspring/decimal"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"golang.org/x/net/context"
)

type awsAPI struct {
	app    *model.App
	ec2    *ec2.Client
	ce     *costexplorer.Client
	ct     *cloudtrail.Client
	cwl    *cloudwatchlogs.Client
	tagKey string

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

func NewAWSAPI(app *model.App, tagKey string) (*awsAPI, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithAppID("mc_manager"), config.WithRetryMaxAttempts(2))
	if err != nil {
		logging.Error("Unable to load SDK config", "error", err)
		return nil, err
	}

	ec2 := ec2.NewFromConfig(cfg)
	ce := costexplorer.NewFromConfig(cfg)
	ct := cloudtrail.NewFromConfig(cfg)
	cwl := cloudwatchlogs.NewFromConfig(cfg)

	if ec2 == nil || ce == nil || ct == nil || cwl == nil {
		logging.Error("Failed to create AWS clients", "ec2", ec2, "ce", ce, "ct", ct, "cwl", cwl)
		return nil, fmt.Errorf("failed to create AWS clients")
	}

	aws := &awsAPI{
		app:    app,
		ec2:    ec2,
		ce:     ce,
		ct:     ct,
		cwl:    cwl,
		tagKey: tagKey,
	}

	_, err = aws.getInstanceStatus()
	if err != nil {
		logging.Error("Failed to get instance status", "error", err)
		return nil, err
	}

	return aws, nil
}

func (aws *awsAPI) getInstanceStatus() (InstanceStatus, error) {
	filterName := fmt.Sprintf("%s:%s", "tag", aws.tagKey)
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

func (aws *awsAPI) GetCostAllocationTag() (*ceTypes.CostAllocationTag, error) {
	input := &costexplorer.ListCostAllocationTagsInput{
		TagKeys: []string{aws.tagKey},
	}

	resp, err := aws.ce.ListCostAllocationTags(context.Background(), input)
	if err != nil {
		logging.Error("Failed to list cost allocation tags", "error", err)
		return nil, err
	}

	for _, tag := range resp.CostAllocationTags {
		if *tag.TagKey == aws.tagKey {
			return &tag, nil
		}
	}
	return nil, fmt.Errorf("tag not found")
}

func (aws *awsAPI) IsCostAllocationTagActive() (bool, error) {
	tag, err := aws.GetCostAllocationTag()
	if err != nil || tag == nil {
		logging.Error("Failed to get cost allocation tag", "error", err, "tag", tag)
		return false, err
	}
	return tag.Status == ceTypes.CostAllocationTagStatusActive, nil
}

func (aws *awsAPI) ActivateCostAllocationTag() error {
	input := &costexplorer.UpdateCostAllocationTagsStatusInput{
		CostAllocationTagsStatus: []ceTypes.CostAllocationTagStatusEntry{
			{
				TagKey: &aws.tagKey,
				Status: ceTypes.CostAllocationTagStatusActive,
			},
		},
	}

	resp, err := aws.ce.UpdateCostAllocationTagsStatus(context.Background(), input)
	if err != nil {
		logging.Error("Failed to create cost allocation tag", "tag", aws.tagKey, "error", err)
		return err
	}

	if len(resp.Errors) > 0 {
		logging.Error("Failed to create cost allocation tag")
		for _, err := range resp.Errors {
			logging.Error("Error creating cost allocation tag", "tag", err.TagKey, "error", err.Message, "code", err.Code)
		}
		return fmt.Errorf("failed to create cost allocation tag: %s", aws.tagKey)
	}

	logging.Info("Cost allocation tag created successfully")
	return nil
}

func (aws *awsAPI) BackfillCostAllocationTag(backfillFrom time.Time) error {
	backfillFrom = time.Date(backfillFrom.Year(), backfillFrom.Month(), 1, 0, 0, 0, 0, time.UTC)
	if backfillFrom.After(time.Now().UTC()) {
		return fmt.Errorf("backfillFrom date cannot be in the future")
	}

	backfillString := backfillFrom.Format("2006-01-02T15:04:05Z")
	logging.Info("Starting cost allocation tag backfill",
		"tag", aws.tagKey,
		"backfillFrom", backfillString,
	)

	input := &costexplorer.StartCostAllocationTagBackfillInput{
		BackfillFrom: &backfillString,
	}

	resp, err := aws.ce.StartCostAllocationTagBackfill(context.Background(), input)
	if err != nil {
		logging.Error("Failed to backfill cost allocation tag", "tag", aws.tagKey, "error", err)
		return err
	}
	logging.Info("Cost allocation tag backfill started successfully",
		"tag", aws.tagKey,
		"status", resp.BackfillRequest.BackfillStatus,
		"backfillFrom", backfillFrom,
		"lastUpdate", resp.BackfillRequest.LastUpdatedAt,
		"completedAt", resp.BackfillRequest.CompletedAt,
		"requestedAt", resp.BackfillRequest.RequestedAt,
	)
	return nil
}

func (aws *awsAPI) IsBackfillInProgress() (bool, error) {
	maxResults := int32(1)
	input := &costexplorer.ListCostAllocationTagBackfillHistoryInput{
		MaxResults: &maxResults,
	}
	resp, err := aws.ce.ListCostAllocationTagBackfillHistory(context.Background(), input)
	if err != nil {
		logging.Error("Failed to list cost allocation tag backfill history", "error", err)
		return false, err
	}
	if len(resp.BackfillRequests) == 0 {
		logging.Info("No cost allocation tag backfill history found")
		return false, nil
	}

	logging.Info("Last allocation tag backfill",
		"status", resp.BackfillRequests[0].BackfillStatus,
		"backfillFrom", *resp.BackfillRequests[0].BackfillFrom,
		"lastUpdate", *resp.BackfillRequests[0].LastUpdatedAt,
		"completedAt", resp.BackfillRequests[0].CompletedAt,
		"requestedAt", resp.BackfillRequests[0].RequestedAt,
	)
	return resp.BackfillRequests[0].BackfillStatus == ceTypes.CostAllocationTagBackfillStatusProcessing, nil
}

func (aws *awsAPI) GetCostForTimeframe(timeframe model.Timeframe) ([]*model.MonthsCost, error) {
	startDateString := timeframe.Start.Format("2006-01-02")
	endDateString := timeframe.End.Format("2006-01-02")

	logging.Info("Getting cost for timeframe",
		"startDate", startDateString,
		"endDate", endDateString,
	)

	fixedLabels := []string{
		"VolumeUsage",
		"APIRequest",
	}

	keyName := "UnblendedCost"
	groupBy := string(ceTypes.DimensionUsageType)
	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &ceTypes.DateInterval{
			Start: &startDateString,
			End:   &endDateString,
		},
		Metrics:     []string{"UnblendedCost"},
		Granularity: ceTypes.GranularityMonthly,
		GroupBy: []ceTypes.GroupDefinition{
			{
				Key:  &groupBy,
				Type: ceTypes.GroupDefinitionTypeDimension,
			},
		},
		Filter: &ceTypes.Expression{
			Tags: &ceTypes.TagValues{
				Key:    &aws.tagKey,
				Values: []string{aws.app.Config.GetAWSProjectName()},
			},
		},
	}

	resp, err := aws.ce.GetCostAndUsage(context.Background(), input)
	if err != nil {
		logging.Error("Failed to get cost and usage", "error", err)
		return nil, err
	}
	monthlyCosts := make([]*model.MonthsCost, 0)
	for _, result := range resp.ResultsByTime {
		if result.TimePeriod == nil || result.TimePeriod.Start == nil || result.TimePeriod.End == nil {
			logging.Error("Time period is nil")
			continue
		}
		startDate, err := time.Parse("2006-01-02", *result.TimePeriod.Start)
		if err != nil {
			logging.Error("Failed to parse start date", "error", err)
			continue
		}
		endDate, err := time.Parse("2006-01-02", *result.TimePeriod.End)
		if err != nil {
			logging.Error("Failed to parse end date", "error", err)
			continue
		}
		if startDate.Month() != endDate.Month() {
			logging.Error("Start and end date are not in the same month", "startDate", startDate, "endDate", endDate)
			continue
		}

		billingMonth := model.NewYearMonthFromTime(startDate)

		timeframe := model.Timeframe{
			Start: startDate,
			End:   endDate,
		}

		totalAmount := decimal.Zero
		fixedAmount := decimal.Zero
		variableAmount := decimal.Zero
		currency := "USD"
		items := make([]model.CostItem, 0)
		for _, group := range result.Groups {
			if len(group.Keys) != 1 {
				logging.Error("Group metrics are not 1")
				continue
			}
			if len(group.Metrics) == 0 {
				logging.Error("Group metrics are nil or empty")
				continue
			}
			label := group.Keys[0]

			if group.Metrics[keyName].Unit == nil {
				logging.Error("Group metrics unit is nil", "label", label)
				continue
			}
			if *group.Metrics[keyName].Unit != "USD" {
				logging.Error("Group metrics unit is not in", "base_currency", currency, "label", label, "unit", *group.Metrics[keyName].Unit)
				continue
			}

			if group.Metrics[keyName].Amount == nil {
				logging.Error("Group metrics amount is nil", "label", label)
				continue
			}

			amountString := *group.Metrics[keyName].Amount
			amount, err := decimal.NewFromString(amountString)
			if err != nil {
				logging.Error("Failed to parse amount", "error", err)
				continue
			}

			isFixed := false
			for _, fixedLabel := range fixedLabels {
				if strings.Contains(label, fixedLabel) {
					isFixed = true
					break
				}
			}

			item := model.CostItem{
				Label:   label,
				Amount:  amount,
				IsFixed: isFixed,
			}

			items = append(items, item)
			totalAmount = totalAmount.Add(amount)
			if isFixed {
				fixedAmount = fixedAmount.Add(amount)
			} else {
				variableAmount = variableAmount.Add(amount)
			}
		}

		//calculate taxes
		taxRate := aws.app.Config.GetTaxRate()
		if !taxRate.Equal(decimal.Zero) {
			fixedTax := fixedAmount.Mul(taxRate)
			fixedTaxItem := model.CostItem{
				Label:   "Fixed Tax",
				Amount:  fixedTax,
				IsFixed: true,
			}

			variableTax := variableAmount.Mul(taxRate)
			variableTaxItem := model.CostItem{
				Label:   "Variable Tax",
				Amount:  variableTax,
				IsFixed: false,
			}

			items = append(items, fixedTaxItem, variableTaxItem)
			fixedAmount = fixedAmount.Add(fixedTax)
			variableAmount = variableAmount.Add(variableTax)
			totalAmount = totalAmount.Add(fixedTax).Add(variableTax)
		}

		cost := model.MonthsCost{
			Timeframe:    timeframe,
			BillingMonth: billingMonth,
			Total:        totalAmount,
			Fixed:        fixedAmount,
			Variable:     variableAmount,
			Currency:     currency,
			Items:        items,
		}
		monthlyCosts = append(monthlyCosts, &cost)
		logging.Debug("Monthly cost", "total", totalAmount, "fixed", fixedAmount, "variable", variableAmount, "currency", currency)

	}

	return monthlyCosts, nil
}

func (aws *awsAPI) GetLogsSince(startTime time.Time, limit int) ([]*model.Log, error) {
	logging.Info("Getting logs since", "startTime", startTime, "limit", limit)
	logGroupName := "/mc-server/logs"
	//query log insights
	query := "fields @timestamp, @message, @logStream, @log | sort @timestamp asc"
	startTimestamp := startTime.Unix()
	endTimestamp := time.Now().Unix()
	limitInt := int32(limit)
	input := &cloudwatchlogs.StartQueryInput{
		LogGroupName: &logGroupName,
		StartTime:    &startTimestamp,
		EndTime:      &endTimestamp,
		QueryString:  &query,
		Limit:        &limitInt,
	}
	resp, err := aws.cwl.StartQuery(context.Background(), input)
	if err != nil {
		logging.Error("Failed to start query", "error", err)
		return nil, err
	}

	logging.Debug("Started query", "queryId", *resp.QueryId)
	var queryResults *cloudwatchlogs.GetQueryResultsOutput
	for {
		queryResults, err = aws.cwl.GetQueryResults(context.Background(), &cloudwatchlogs.GetQueryResultsInput{
			QueryId: resp.QueryId,
		})
		if err != nil {
			logging.Error("Failed to get query results", "error", err)
			return nil, err
		}
		if queryResults.Status == cwlTypes.QueryStatusComplete {
			break
		}

		if queryResults.Status == cwlTypes.QueryStatusRunning || queryResults.Status == cwlTypes.QueryStatusScheduled {
			logging.Debug("Query is still running", "queryId", *resp.QueryId, "status", queryResults.Status)
			time.Sleep(5 * time.Second)
			continue
		}

		return nil, fmt.Errorf("cloudwatch query failed with status: %s", queryResults.Status)
	}

	if queryResults.Results == nil {
		logging.Error("Query results are nil")
		return nil, fmt.Errorf("query results are nil")
	}

	events := make([]*model.Log, 0)
	for _, fields := range queryResults.Results {
		var timestampString *string
		var message *string
		for _, field := range fields {
			if field.Field == nil {
				continue
			}
			if *field.Field == "@timestamp" {
				timestampString = field.Value
			} else if *field.Field == "@message" {
				message = field.Value
			}
		}

		if timestampString == nil || message == nil {
			logging.Error("Timestamp or message is nil")
			continue
		}

		timestampTime, err := time.Parse("2006-01-02 15:04:05.999999999", *timestampString)
		if err != nil {
			logging.Error("Failed to parse timestamp", "error", err)
			continue
		}

		events = append(events, &model.Log{
			Timestamp: timestampTime,
			Message:   *message,
		})
	}

	return events, nil
}
