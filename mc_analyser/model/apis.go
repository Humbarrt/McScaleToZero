package model

import ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

type AWSAPI interface {
	StartInstance() error
	GetInstanceState() (ec2Types.InstanceStateName, error)
	IsInstanceOnline() (bool, error)
	GetInstanceIP() (string, error)
}
