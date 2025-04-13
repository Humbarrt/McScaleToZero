package model

import (
	"time"

	ceTypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type AWSAPI interface {
	StartInstance() error
	GetInstanceState() (ec2Types.InstanceStateName, error)
	IsInstanceOnline() (bool, error)
	GetInstanceIP() (string, error)
	GetCostAllocationTag(TagKey string) (*ceTypes.CostAllocationTag, error)
	ActivateCostAllocationTag(TagKey string) error
	BackfillCostAllocationTag(TagKey string, backfillFrom time.Time) error
}
