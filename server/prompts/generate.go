package prompts

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pyprism/uCPingGraph/utils"
)

func GenerateDummyData() {
	tokenContent := commonPromptContent{
		ErrorMsg: "Device token cannot be empty!",
		Label:    "Device token",
	}

	token := commonPromptInput(tokenContent)
	serverPort := utils.GetEnv("SERVER_PORT", "8080")
	url := "http://127.0.0.1:" + serverPort + "/api/stats"

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// call local API
	for {
		select {
		case <-stop:
			log.Println("stopping dummy data generator")
			return
		default:
		}

		latency := utils.RandomFloat()
		sentPackets := 5
		// 0..sentPackets received, so loss ranges from none to a full outage
		// instead of only ever 0% or 20%.
		receivedPackets := utils.RandomInt() % (sentPackets + 1)
		packetLoss := (float64(sentPackets-receivedPackets) / float64(sentPackets)) * 100

		jsonData := map[string]interface{}{
			"latency_ms":          latency,
			"sent_packets":        sentPackets,
			"received_packets":    receivedPackets,
			"packet_loss_percent": packetLoss,
			"target":              "1.1.1.1",
			"platform":            "dummy",
			"rssi":                -55,
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

		log.Println("dummy response Status:", resp.Status)
		_ = resp.Body.Close()

		sleepSec := utils.RandomInt()
		if sleepSec < 1 {
			sleepSec = 1
		}
		select {
		case <-stop:
			log.Println("stopping dummy data generator")
			return
		case <-time.After(time.Duration(sleepSec) * time.Second):
		}
	}

}
