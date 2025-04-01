package rest

import (
	"encoding/json"
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"net/http"
	"sort"
)

type restAPI struct {
	flyApi *model.FlyAPI
}

func NewRestAPI(flyApi *model.FlyAPI) *restAPI {
	return &restAPI{
		flyApi: flyApi,
	}
}

func (r *restAPI) request(method string, route string) (*http.Response, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/v1/apps/%s/machines%s", r.flyApi.FLY_API_BASE_URL, r.flyApi.FLY_APP_NAME, route)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+r.flyApi.FLY_API_TOKEN)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (r *restAPI) GetNewestMachine() (*model.MachineInfo, error) {
	resp, err := r.request("GET", "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get machine ID, status: %d", resp.StatusCode)
	}

	var machines []model.MachineInfo
	if err := json.NewDecoder(resp.Body).Decode(&machines); err != nil {
		return nil, err
	}

	if len(machines) == 0 {
		logging.Fatal("No machines found")
	}

	sort.Slice(machines, func(i, j int) bool {
		return (machines)[i].CreatedAt.After((machines)[j].CreatedAt)
	})
	return &machines[0], nil
}

func (r *restAPI) IsServerOnline() (bool, error) {
	resp, err := r.request("GET", "/"+r.flyApi.FLY_MACHINE_ID)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to check server status, status: %d", resp.StatusCode)
	}

	var machine model.MachineInfo
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return false, err
	}
	if machine.Id != r.flyApi.FLY_MACHINE_ID {
		return false, fmt.Errorf("machine ID mismatch: expected %s, got %s", r.flyApi.FLY_MACHINE_ID, machine.Id)
	}

	if machine.State == "started" {
		return true, nil
	} else {
		return false, nil
	}
}
