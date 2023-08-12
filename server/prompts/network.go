package prompts

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pyprism/uCPingGraph/models"
	"log"
	"os"
)

type NetworkPromptContent struct {
	ErrorMsg string
	Label    string
}

func networkPromptInput(pc NetworkPromptContent) string {
	validate := func(input string) error {
		if len(input) <= 0 {
			return errors.New(pc.ErrorMsg)
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
		Label:     pc.Label,
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

func CreateNetwork() {
	networkContent := NetworkPromptContent{
		ErrorMsg: "Network name cannot be empty!",
		Label:    "Network name",
	}

	networkName := networkPromptInput(networkContent)
	log.Println(networkName)
	networkDb := models.Network{}
	networkDb.CreateNetwork(networkName)

}
