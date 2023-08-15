package prompts

import (
	"log"

	"github.com/pyprism/uCPingGraph/models"
)

type DevicePromptContent struct {
	ErrorMsg string
	Label    string
}

func devicePromptInput(content DevicePromptContent) string {
	networkModel := models.Network{}
	networks, err := networkModel.GetAllNetworkName()
	if err != nil {
		log.Println(err.Error())
		
	}
	index := -1
    var result string
    var err error

	for index < 0 {
		prompt := promptui.SelectWithAdd{
			Label: content.Label
			Items: networks,
		}

		index, result, err = prompt.Run()

        if index == -1 {
            items = append(items, result)
        }
	}

	if err != nil {
        fmt.Printf("Prompt failed %v\n", err)
        os.Exit(1)
    }
	return result
}

func AddNewDevice() {
	networkListPrompt := DevicePromptContent{
		Label: "Select network",
		ErrorMsg: "Please select network from the list",
	}

	networkName := devicePromptInput(networkListPrompt)
	log.Println(networkName)
}