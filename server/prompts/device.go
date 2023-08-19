package prompts

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"log"
	"os"

	"github.com/pyprism/uCPingGraph/models"
)

type DevicePromptContent struct {
	ErrorMsg string
	Label    string
}

func devicePromptInputSelect(content DevicePromptContent) string {
	networkModel := models.Network{}
	networks, err := networkModel.GetAllNetworkName()
	if err != nil {
		log.Println(err.Error())

	}
	index := -1
	var result string

	for index < 0 {
		prompt := promptui.SelectWithAdd{
			Label: content.Label,
			Items: networks,
		}

		index, result, err = prompt.Run()

		if index == -1 {
			networks = append(networks, result)
		}
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return result
}

func devicePromptInput(dp DevicePromptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(dp.ErrorMsg)
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     dp.Label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", result)

	return result
}

func AddNewDevice() {
	networkListPrompt := DevicePromptContent{
		Label:    "Select network",
		ErrorMsg: "Please select network from the list",
	}

	networkName := devicePromptInputSelect(networkListPrompt)
	network := models.Network{}
	networkId, err := network.GetNetworkIdByName(networkName)
	if err != nil {
		log.Println(err.Error())
	}

	deviceNamePrompt := DevicePromptContent{
		Label:    "Device name",
		ErrorMsg: "Device name cannot be empty and must be unique in the network",
	}

	deviceName := devicePromptInput(deviceNamePrompt)
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
