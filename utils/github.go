package utils

import (
	"encoding/json"
	"net/http"
)

func GetPublicSSHKeys(username string) ([]string, error) {
	client := http.Client{}

	resp, err := client.Get("https://api.github.com/users/" + username + "/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	keys := []map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	var publicKeys []string
	for _, key := range keys {
		if key["key"] != nil {
			publicKeys = append(publicKeys, key["key"].(string))
		}
	}

	return publicKeys, nil
}
