package mojang

import (
	"encoding/json"
	"fmt"
	"mc_analyser/logging"
	"mc_analyser/model"
	"net/http"
	"time"
)

func request(route string, method string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, route, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		logging.Warn("Error making Mojang API request, trying again:", "error", err)
		resp.Body.Close()
		time.Sleep(1 * time.Second)
		resp, err = client.Do(req)
		if err != nil {
			logging.Error("Second Mojang API request also failed", "error", err)
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("mojang API request returned status: %s", resp.Status)
	}

	return resp, nil
}

func getRequest(route string, responseVal any) error {
	resp, err := request(route, "GET")
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
	route := "https://api.mojang.com/users/profiles/minecraft/" + playerName
	var result struct {
		ID string `json:"id"`
	}

	err := getRequest(route, &result)
	if err != nil {
		return nil, err
	}
	return model.NewUUIDFromString(result.ID)
}

func GetPlayerSession(serverId string, playerName string, ipv4 string) (*model.UUID, error) {
	route := fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s&ip=%s", playerName, serverId, ipv4)
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	err := getRequest(route, &result)
	if err != nil {
		logging.Error("Error getting player session", "error", err)
		return nil, err
	}

	if result.Name != playerName {
		logging.Error("Player name does not match", "expected", playerName, "got", result.Name)
		return nil, fmt.Errorf("player name does not match")
	}

	logging.Info("Player session", "id", result.ID, "name", result.Name)
	return model.NewUUIDFromString(result.ID)
}
