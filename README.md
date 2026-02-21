# uCPingGraph

uCPingGraph tracks network quality from microcontrollers and visualizes it in a web dashboard.

## Project Layout

- `client/esp8266`: PlatformIO firmware for both ESP8266 and ESP32 (using WiFiManager)
- `server`: Go backend, SQLite storage, API, and dashboard

## Features

- ESP8266 and ESP32 support from one firmware codebase
- WiFi provisioning with WiFiManager portal
- Latency + packet-loss telemetry ingestion
- Dashboard with dual-series chart and summary cards
- Backend unit tests for ingestion and API behavior

## Quick Start

1. Start server:

```bash
cd server
go run . server
```

2. Create network and device token:

```bash
go run . network add
go run . device add
```

3. Power the board and configure WiFi + telemetry values in WiFiManager portal:

- SSID: `uCPingGraph-Setup`

4. Build firmware:

```bash
cd client/esp
platformio run -e nodemcuv2
platformio run -e esp32dev
```

See `server/README.md` and `client/esp/README.md` for details.
