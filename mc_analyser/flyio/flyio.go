package flyio

import (
	"mc_analyser/flyio/graphql"
	"mc_analyser/flyio/rest"
	"mc_analyser/logging"
	"mc_analyser/model"
)

func NewAPI(FLY_API_BASE_URL string, FLY_API_TOKEN string, FLY_APP_NAME string) *model.FlyAPI {
	f := &model.FlyAPI{
		FLY_API_BASE_URL: FLY_API_BASE_URL,
		FLY_API_TOKEN:    FLY_API_TOKEN,
		FLY_APP_NAME:     FLY_APP_NAME,
	}
	f.RestAPI = rest.NewRestAPI(f)

	machine, err := f.RestAPI.GetNewestMachine()
	if err != nil {
		logging.Fatal("Error getting newest machine: ", "error", err)
	}

	f.FLY_MACHINE_ID = machine.Id
	logging.Debug("Machine ID: ", "machine_id", f.FLY_MACHINE_ID)
	logging.Debug("Machine IP: ", "machine_ip", f.FLY_MACHINE_IP)

	graphql.GetBilling(f.FLY_API_TOKEN)
	graphql.TestSSH(f)
	return f
}
