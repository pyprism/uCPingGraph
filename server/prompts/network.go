package prompts

import (
	"github.com/pyprism/uCPingGraph/models"
)

func CreateNetwork() {
	networkContent := commonPromptContent{
		ErrorMsg: "Network name cannot be empty!",
		Label:    "Network name",
	}

	networkName := commonPromptInput(networkContent)
	networkDb := models.Network{}
	networkDb.CreateNetwork(networkName)

}
