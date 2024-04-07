package prompts

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"log"
	"os"
)

type commonPromptContent struct {
	ErrorMsg string
	Label    string
}

func commonPromptInput(pc commonPromptContent) string {
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
		log.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", result)

	return result
}
