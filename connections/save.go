package connections

import (
	"blueprint/appinfo"
	"blueprint/models"
	"encoding/json"
	"fmt"
	"os"
)

// Write to JSON file
func SaveConnections(config *models.ConnectionsFile) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(appinfo.ConnectionsFile, data, 0644)
}

func UpdateConnection(config *models.ConnectionsFile, id string, updated models.Connection) error {
	for i := range config.Connections {
		if config.Connections[i].Id == id {
			updated.Id = id
			config.Connections[i] = updated
			return SaveConnections(config)
		}
	}

	return fmt.Errorf("connection with id %q not found", id)
}
