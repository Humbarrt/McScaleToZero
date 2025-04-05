package mojang

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetPlayerUUID(playerName string) (string, error) {
	url := "https://api.mojang.com/users/profiles/minecraft/" + playerName
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get player UUID: %s", resp.Status)
	}

	var result struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	// convert into 8-4-4-4-12 format

	if len(result.ID) != 32 {
		return "", fmt.Errorf("invalid UUID format: %s", result.ID)

	}

	result.ID = fmt.Sprintf("%s-%s-%s-%s-%s", result.ID[:8], result.ID[8:12], result.ID[12:16], result.ID[16:20], result.ID[20:])

	return result.ID, nil
}
