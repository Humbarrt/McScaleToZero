package mojang

import (
	"encoding/json"
	"fmt"
	"mc_analyser/model"
	"net/http"
)

func GetPlayerUUID(playerName string) (*model.UUID, error) {
	url := "https://api.mojang.com/users/profiles/minecraft/" + playerName
	resp, err := http.Get(url)
	if err != nil {
		return &model.UUID{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &model.UUID{}, fmt.Errorf("failed to get player UUID: %s", resp.Status)
	}

	var result struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return &model.UUID{}, err
	}

	return model.UUIDFromString(result.ID)
}
