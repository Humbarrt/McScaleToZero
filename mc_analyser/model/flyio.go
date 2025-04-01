package model

import "time"

type FlyAPI struct {
	FLY_API_BASE_URL string
	FLY_API_TOKEN    string
	FLY_APP_NAME     string
	FLY_MACHINE_ID   string
	FLY_MACHINE_IP   string
	RestAPI          RestAPI
}

type MachineInfo struct {
	Id        string    `json:"id"`
	Ip        string    `json:"ip"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

type RestAPI interface {
	GetNewestMachine() (*MachineInfo, error)
	IsServerOnline() (bool, error)
}
