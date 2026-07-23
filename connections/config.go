package connections

import (
	"blueprint/appinfo"
	"blueprint/models"

	"encoding/json"
	"os"
)

// Load connections from config.json file
func LoadConnections() (*models.ConnectionsFile, error) {

	var config models.ConnectionsFile

	if _, err := os.Stat(appinfo.ConnectionsFile); os.IsNotExist(err) {
		SaveConnections(&config)
	}

	data, err := os.ReadFile(appinfo.ConnectionsFile)
	if err != nil {
		return nil, err
	}

	// d

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
