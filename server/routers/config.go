package routers

import (
	"fmt"
	"github.com/pyprism/uCPingGraph/utils"
)

func Init() {
	r := NewRouter()
	serverPort := utils.GetEnv("SERVER_PORT", "8880")
	fmt.Println("Server running on http://127.0.0.1:" + serverPort)
	err := r.Run(":" + serverPort)
	if err != nil {
		panic(err)
	}
}
