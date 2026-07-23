package connections

import (
	"blueprint/models"
	"strings"
)

func FindConnectionById(config *models.ConnectionsFile, id string) *models.Connection {
	for i := range config.Connections {
		if config.Connections[i].Id == id {
			return &config.Connections[i]
		}
	}
	return nil
	//return nil, fmt.Errorf("Connection with Id %q not found", id)
}

func FindConnectionByName(Name string) (string, *models.Connection) {
	config, _ := LoadConnections()

	for i := range config.Connections {
		if strings.EqualFold(config.Connections[i].Name, Name) { // Case insensitive match
			return config.Connections[i].Id, &config.Connections[i]
		}
	}
	return "", nil
}
