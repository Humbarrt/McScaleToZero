package model

type AWSAPI interface {
	StartInstance() error
	StopInstance() error
	IsInstanceOnline() (bool, error)
	GetInstanceIP() (string, error)
}
