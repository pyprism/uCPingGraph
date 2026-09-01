# uCPingGraph [![Server Tests](https://github.com/pyprism/uCPingGraph/actions/workflows/server-tests.yml/badge.svg)](https://github.com/pyprism/uCPingGraph/actions/workflows/server-tests.yml) [![codecov](https://codecov.io/gh/pyprism/uCPingGraph/graph/badge.svg?token=4PWDHUP8X0)](https://codecov.io/gh/pyprism/uCPingGraph)

uCPingGraph tracks network quality from microcontrollers and visualizes it in a web dashboard.

## Features

- ESP8266 and ESP32 support from one firmware codebase
- WiFi provisioning with WiFiManager portal
- Latency + packet-loss telemetry ingestion
- Dashboard with dual-series chart and summary cards
- Backend unit tests for ingestion and API behavior

## Screenshot

![Dashboard screenshot][screenshot]

[screenshot]: screenshot.png

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

4. Build and upload firmware:

```bash
cd client/esp
pio run -e nodemcuv2 -t upload
pio run -e esp32dev -t upload
```

See `server/README.md` and `client/esp/README.md` for details.
