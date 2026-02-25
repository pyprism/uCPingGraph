package prompts

import (
	"log"

	"github.com/pyprism/uCPingGraph/models"
)

func CreateNetwork() {
	networkContent := commonPromptContent{
		ErrorMsg: "Network name cannot be empty!",
		Label:    "Network name",
	}

	networkName := commonPromptInput(networkContent)
	networkDb := models.Network{}
	id, err := networkDb.CreateNetwork(networkName)
	if err != nil {
		log.Printf("Failed to create network: %v\n", err)
		return
	}
	log.Printf("Network created successfully with ID: %d\n", id)
}
