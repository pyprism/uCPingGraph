package prompts

import (
	"bytes"
	"encoding/json"
	"github.com/pyprism/uCPingGraph/utils"
	"log"
	"net/http"
	"time"
)

func GenerateDummyData() {
	tokenContent := commonPromptContent{
		ErrorMsg: "Device token cannot be empty!",
		Label:    "Device token",
	}

	token := commonPromptInput(tokenContent)
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
