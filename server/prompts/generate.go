package prompts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pyprism/uCPingGraph/utils"
	"log"
	"net/http"
	"os"
	"time"
)

type dummyDataPromptContent struct {
	ErrorMsg string
	Label    string
}

func dummyDataPromptInput(pc dummyDataPromptContent) string {
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

func GenerateDummyData() {
	tokenContent := dummyDataPromptContent{
		ErrorMsg: "Device token cannot be empty!",
		Label:    "Device token",
	}

	token := dummyDataPromptInput(tokenContent)
	serverPort := utils.GetEnv("SERVER_PORT", "8080")
	url := "http://127.0.0.1:" + serverPort + "/api/stats/"

	// call local API
	for {
		latency := utils.RandomFloat()
		jsonData := map[string]interface{}{
			"latency": latency,
		}

		jsonDataBytes, err := json.Marshal(jsonData)
		if err != nil {
			log.Println("json encoding error:", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonDataBytes))
		if err != nil {
			log.Println("http request error:", err)
			return
		}

		req.Header.Set("Authorization", token)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println("http failed: ", err)
			return
		}
		defer resp.Body.Close()

		log.Println("dummy response Status:", resp.Status)
		time.Sleep(time.Duration(utils.RandomInt()) * time.Second)
	}

}
