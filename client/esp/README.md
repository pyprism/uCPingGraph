# Firmware (ESP8266 + ESP32)

This PlatformIO project supports both boards from one source file:

- `nodemcuv2` (ESP8266)
- `esp32dev` (ESP32)

## Required Setup

No runtime values need to be hardcoded in source.

WiFi credentials and telemetry settings are configured from WiFiManager portal:

- `uCPingGraph-Setup`

Portal fields:

- Server URL
- Device Token
- Ping Target (IP/hostname)
- Probe Count
- Send Interval (ms)

Values are persisted in LittleFS at `/ucpinggraph.json`.

## Runtime Factory Reset Button

- Pin: `GPIO14` (`D5` on NodeMCU ESP8266)
- Wiring: connect button between `GPIO14` and `GND`
- Action: press and hold for 5 seconds to:
  - clear WiFi credentials (WiFiManager reset)
  - delete saved `/ucpinggraph.json`
  - reboot into fresh setup state

## Build

```bash
platformio run -e nodemcuv2
platformio run -e esp32dev
```

## Payload Sent to Server

```json
{
  "latency_ms": 11.8,
  "sent_packets": 5,
  "received_packets": 5,
  "packet_loss_percent": 0,
  "target": "1.1.1.1",
  "platform": "esp32",
  "rssi": -61
}
```
