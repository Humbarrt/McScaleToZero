package mojang

import (
	"encoding/json"
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"net/http"
)

func requestMojangAPI(route string, method string) (*http.Response, error) {
	baseUrl := "https://api.mojang.com/"
	client := &http.Client{}
	req, err := http.NewRequest(method, baseUrl+route, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		logging.Warn("Error making request to Mojang API, trying again:", "error", err)
		resp.Body.Close()
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("mojang API request failed: %s", resp.Status)
	}

	return resp, nil
}

func getRequestMojangAPI(route string, responseVal any) error {
	resp, err := requestMojangAPI(route, "GET")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(responseVal)
	if err != nil {
		return err
	}

	return nil
}

func GetPlayerUUID(playerName string) (*model.UUID, error) {
	route := "users/profiles/minecraft/" + playerName
	var result struct {
		ID string `json:"id"`
	}

	err := getRequestMojangAPI(route, &result)
	if err != nil {
		return nil, err
	}
	return model.UUIDFromString(result.ID)
}
