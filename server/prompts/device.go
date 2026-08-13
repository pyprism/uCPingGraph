package prompts

import (
	"fmt"
	"log"
	"os"

	"github.com/manifoldco/promptui"

	"github.com/pyprism/uCPingGraph/models"
)

func devicePromptInputSelect(content commonPromptContent) string {
	networkModel := models.Network{}
	networks, err := networkModel.GetAllNetworkName()
	if err != nil {
		log.Println(err.Error())
	}

	prompt := promptui.SelectWithAdd{
		Label: content.Label,
		Items: networks,
	}

	index, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	if index == -1 {
		// The user typed a name that wasn't in the list; create the network
		// now so a later lookup by name succeeds.
		newNetwork := models.Network{}
		if _, err := newNetwork.CreateNetwork(result); err != nil {
			fmt.Printf("Failed to create network %q: %v\n", result, err)
			os.Exit(1)
		}
	}

	return result
}

func AddNewDevice() {
	networkListPrompt := commonPromptContent{
		Label:    "Select network",
		ErrorMsg: "Please select network from the list",
	}

	networkName := devicePromptInputSelect(networkListPrompt)
	network := models.Network{}
	networkId, err := network.GetNetworkIdByName(networkName)
	if err != nil {
		log.Printf("Failed to get network ID: %v\n", err)
		return
	}

	deviceNamePrompt := commonPromptContent{
		Label:    "Device name",
		ErrorMsg: "Device name cannot be empty and must be unique in the network",
	}

	deviceName := commonPromptInput(deviceNamePrompt)
	device := models.Device{}

	// Check if the device name is unique in the network
	isUnique := device.CheckDeviceNameIsUnique(int(networkId), deviceName)
	if !isUnique {
		log.Println("Device name is not unique in the network")
		return
	}

	// Create device
	_, token, errr := device.CreateDevice(int(networkId), deviceName)
	if errr != nil {
		log.Println(errr.Error())
		return
	}
	log.Println("Device token: " + token)
}
