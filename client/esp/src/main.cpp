#include <Arduino.h>
#include <ArduinoJson.h>
#include <WiFiManager.h>

#if defined(ESP8266)
#include <ESP8266HTTPClient.h>
#include <ESP8266Ping.h>
#include <ESP8266WiFi.h>
#include <LittleFS.h>
#include <WiFiClientSecureBearSSL.h>
#elif defined(ESP32)
#include <HTTPClient.h>
#include <ESP32Ping.h>
#include <LittleFS.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#else
#error "Unsupported board. Build for ESP8266 or ESP32."
#endif

const char *kWiFiPortalSSID = "uCPingGraph-Setup";
const char *kConfigPath = "/ucpinggraph.json";
const uint8_t kResetButtonPin = 14; // D5 on ESP8266, GPIO14 on ESP32
const uint8_t kBootButtonPin = 0;   // BOOT button on many ESP32/ESP8266 boards
const unsigned long kResetHoldMs = 5000;

struct DeviceConfig {
  String serverBaseURL = "";
  String deviceToken = "";
  String pingTarget = "1.1.1.1";
  uint8_t probeCount = 5;
  unsigned long telemetryIntervalMs = 5000;
};

#if defined(ESP8266)
const char *kPlatformName = "esp8266";
#else
const char *kPlatformName = "esp32";
#endif

DeviceConfig gConfig;
bool gShouldSaveConfig = false;

struct ProbeResult {
  int sentPackets = 0;
  int receivedPackets = 0;
  float packetLossPercent = 0.0f;
  float averageLatencyMs = 0.0f;
};

String buildStatsURL() {
  String base = gConfig.serverBaseURL;
  base.trim();
  while (base.endsWith("/")) {
    base.remove(base.length() - 1);
  }
  return base + "/api/stats";
}

bool resolveTarget(IPAddress &targetIP) {
  if (targetIP.fromString(gConfig.pingTarget)) {
    return true;
  }
  return WiFi.hostByName(gConfig.pingTarget.c_str(), targetIP) == 1;
}

ProbeResult runProbe(const IPAddress &targetIP) {
  ProbeResult result;
  result.sentPackets = gConfig.probeCount;

  float latencySum = 0.0f;
  for (uint8_t i = 0; i < result.sentPackets; i++) {
    if (Ping.ping(targetIP, 1)) {
      result.receivedPackets++;
      latencySum += Ping.averageTime();
    }
    delay(80);
  }

  if (result.receivedPackets > 0) {
    result.averageLatencyMs = latencySum / result.receivedPackets;
  }

  result.packetLossPercent =
      (float(result.sentPackets - result.receivedPackets) / result.sentPackets) *
      100.0f;

  return result;
}

bool postStats(const ProbeResult &probe) {
  const String url = buildStatsURL();

  JsonDocument payload;
  payload["latency_ms"] = probe.averageLatencyMs;
  payload["sent_packets"] = probe.sentPackets;
  payload["received_packets"] = probe.receivedPackets;
  payload["packet_loss_percent"] = probe.packetLossPercent;
  payload["target"] = gConfig.pingTarget;
  payload["platform"] = kPlatformName;
  payload["rssi"] = WiFi.RSSI();

  String body;
  serializeJson(payload, body);

#if defined(ESP8266)
  BearSSL::WiFiClientSecure client;
  client.setInsecure();
#else
  WiFiClientSecure client;
  client.setInsecure();
#endif

  HTTPClient https;
  if (!https.begin(client, url)) {
    Serial.println("HTTPS begin failed");
    return false;
  }

  https.addHeader("Content-Type", "application/json");
  https.addHeader("Accept", "application/json");
  https.addHeader("Authorization", gConfig.deviceToken);

  const int statusCode = https.POST(body);
  const bool ok = statusCode >= 200 && statusCode < 300;

  if (!ok) {
    Serial.printf("POST failed. Status=%d, body=%s\n", statusCode,
                  https.getString().c_str());
  } else {
    Serial.printf("POST ok. Status=%d\n", statusCode);
  }

  https.end();
  return ok;
}

void saveConfigCallback() { gShouldSaveConfig = true; }

bool beginStorage() {
  if (!LittleFS.begin()) {
    Serial.println("LittleFS init failed, formatting...");
    if (!LittleFS.format()) {
      Serial.println("LittleFS format failed");
      return false;
    }
    if (!LittleFS.begin()) {
      Serial.println("LittleFS init failed after format");
      return false;
    }
  }
  return true;
}

bool loadConfig() {
  if (!LittleFS.exists(kConfigPath)) {
    return false;
  }

  File file = LittleFS.open(kConfigPath, "r");
  if (!file) {
    return false;
  }

  JsonDocument doc;
  DeserializationError err = deserializeJson(doc, file);
  file.close();
  if (err) {
    return false;
  }

  if (doc["server_base_url"].is<const char *>()) {
    gConfig.serverBaseURL = String(doc["server_base_url"].as<const char *>());
  }
  if (doc["device_token"].is<const char *>()) {
    gConfig.deviceToken = String(doc["device_token"].as<const char *>());
  }
  if (doc["ping_target"].is<const char *>()) {
    gConfig.pingTarget = String(doc["ping_target"].as<const char *>());
  }
  if (doc["probe_count"].is<int>()) {
    int probeCount = doc["probe_count"].as<int>();
    if (probeCount >= 1 && probeCount <= 20) {
      gConfig.probeCount = static_cast<uint8_t>(probeCount);
    }
  }
  if (doc["interval_ms"].is<unsigned long>()) {
    unsigned long intervalMs = doc["interval_ms"].as<unsigned long>();
    if (intervalMs >= 1000 && intervalMs <= 3600000) {
      gConfig.telemetryIntervalMs = intervalMs;
    }
  }

  gConfig.serverBaseURL.trim();
  gConfig.deviceToken.trim();
  gConfig.pingTarget.trim();
  return true;
}

bool saveConfig() {
  File file = LittleFS.open(kConfigPath, "w");
  if (!file) {
    Serial.println("Unable to open config for write");
    return false;
  }

  JsonDocument doc;
  doc["server_base_url"] = gConfig.serverBaseURL;
  doc["device_token"] = gConfig.deviceToken;
  doc["ping_target"] = gConfig.pingTarget;
  doc["probe_count"] = gConfig.probeCount;
  doc["interval_ms"] = gConfig.telemetryIntervalMs;

  bool ok = serializeJson(doc, file) > 0;
  file.flush();
  file.close();
  if (ok) {
    Serial.printf("Saved config to %s\n", kConfigPath);
  }
  return ok;
}

bool hasRequiredConfig() {
  return !gConfig.serverBaseURL.isEmpty() && !gConfig.deviceToken.isEmpty() &&
         !gConfig.pingTarget.isEmpty();
}

void factoryResetAndReboot() {
  Serial.println("Factory reset requested: clearing WiFi and config");

  WiFiManager wifiManager;
  wifiManager.resetSettings();

  if (LittleFS.exists(kConfigPath)) {
    if (LittleFS.remove(kConfigPath)) {
      Serial.println("Deleted saved config file");
    } else {
      Serial.println("Failed to delete saved config file");
    }
  }

#if defined(ESP8266)
  WiFi.disconnect(true);
#else
  WiFi.disconnect(true, true);
#endif
  delay(500);
  ESP.restart();
}

void handleResetButton() {
  static bool pressed = false;
  static unsigned long pressedAt = 0;
  const bool isPressed =
      (digitalRead(kResetButtonPin) == LOW) || (digitalRead(kBootButtonPin) == LOW);

  if (isPressed && !pressed) {
    pressed = true;
    pressedAt = millis();
    Serial.println("Reset button pressed; hold to factory reset");
  } else if (!isPressed && pressed) {
    pressed = false;
    Serial.println("Reset button released");
  }

  if (pressed && (millis() - pressedAt >= kResetHoldMs)) {
    factoryResetAndReboot();
  }
}

void checkBootResetRequest() {
  if (digitalRead(kResetButtonPin) == LOW || digitalRead(kBootButtonPin) == LOW) {
    Serial.println("Reset button held at boot. Keep holding for factory reset...");
    unsigned long start = millis();
    while (millis() - start < kResetHoldMs) {
      if (digitalRead(kResetButtonPin) != LOW && digitalRead(kBootButtonPin) != LOW) {
        Serial.println("Boot reset canceled");
        return;
      }
      delay(20);
    }
    factoryResetAndReboot();
  }
}

void applyConfigFromPortal(WiFiManagerParameter &serverURLParam,
                           WiFiManagerParameter &deviceTokenParam,
                           WiFiManagerParameter &pingTargetParam,
                           WiFiManagerParameter &probeCountParam,
                           WiFiManagerParameter &intervalParam) {
  gConfig.serverBaseURL = String(serverURLParam.getValue());
  gConfig.deviceToken = String(deviceTokenParam.getValue());
  gConfig.pingTarget = String(pingTargetParam.getValue());
  gConfig.serverBaseURL.trim();
  gConfig.deviceToken.trim();
  gConfig.pingTarget.trim();

  int probeCount = atoi(probeCountParam.getValue());
  if (probeCount >= 1 && probeCount <= 20) {
    gConfig.probeCount = static_cast<uint8_t>(probeCount);
  }

  unsigned long intervalMs = strtoul(intervalParam.getValue(), nullptr, 10);
  if (intervalMs >= 1000 && intervalMs <= 3600000) {
    gConfig.telemetryIntervalMs = intervalMs;
  }
}

void connectWiFi() {
  WiFi.mode(WIFI_STA);

  WiFiManager wifiManager;
  wifiManager.setConfigPortalTimeout(300);
  wifiManager.setConnectTimeout(15);
  wifiManager.setSaveConfigCallback(saveConfigCallback);

  char serverURL[256];
  char deviceToken[96];
  char pingTarget[64];
  char probeCount[8];
  char intervalMs[12];
  snprintf(serverURL, sizeof(serverURL), "%s", gConfig.serverBaseURL.c_str());
  snprintf(deviceToken, sizeof(deviceToken), "%s", gConfig.deviceToken.c_str());
  snprintf(pingTarget, sizeof(pingTarget), "%s", gConfig.pingTarget.c_str());
  snprintf(probeCount, sizeof(probeCount), "%u", gConfig.probeCount);
  snprintf(intervalMs, sizeof(intervalMs), "%lu", gConfig.telemetryIntervalMs);

  WiFiManagerParameter serverURLParam("server", "Server URL", serverURL,
                                      sizeof(serverURL));
  WiFiManagerParameter deviceTokenParam("token", "Device Token", deviceToken,
                                        sizeof(deviceToken));
  WiFiManagerParameter pingTargetParam("target", "Ping Target (IP/host)",
                                       pingTarget, sizeof(pingTarget));
  WiFiManagerParameter probeCountParam("probes", "Probe Count (1-20)",
                                       probeCount, sizeof(probeCount));
  WiFiManagerParameter intervalParam("interval", "Send Interval ms",
                                     intervalMs, sizeof(intervalMs));

  wifiManager.addParameter(&serverURLParam);
  wifiManager.addParameter(&deviceTokenParam);
  wifiManager.addParameter(&pingTargetParam);
  wifiManager.addParameter(&probeCountParam);
  wifiManager.addParameter(&intervalParam);

  bool connected = false;
  if (hasRequiredConfig()) {
    connected = wifiManager.autoConnect(kWiFiPortalSSID);
  } else {
    Serial.println("Missing config; opening WiFiManager portal");
    connected = wifiManager.startConfigPortal(kWiFiPortalSSID);
    gShouldSaveConfig = true;
  }

  if (!connected) {
    Serial.println("WiFiManager failed. Rebooting...");
    delay(1000);
    ESP.restart();
  }

  applyConfigFromPortal(serverURLParam, deviceTokenParam, pingTargetParam,
                        probeCountParam, intervalParam);
  if (gShouldSaveConfig || !LittleFS.exists(kConfigPath)) {
    if (saveConfig()) {
      Serial.println("Config saved");
    } else {
      Serial.println("Config save failed");
    }
    gShouldSaveConfig = false;
  }

  Serial.print("Connected. IP: ");
  Serial.println(WiFi.localIP());
}

void setup() {
  Serial.begin(115200);
  delay(100);
  pinMode(kResetButtonPin, INPUT_PULLUP);
  pinMode(kBootButtonPin, INPUT_PULLUP);

  Serial.println();
  Serial.println("uCPingGraph client booting...");

  beginStorage();
  checkBootResetRequest();
  if (loadConfig()) {
    Serial.println("Config loaded from LittleFS");
  } else {
    Serial.println("No saved config found");
  }
  connectWiFi();
}

void loop() {
  handleResetButton();

  static unsigned long lastSentAt = 0;
  const unsigned long now = millis();
  if (now - lastSentAt < gConfig.telemetryIntervalMs) {
    delay(40);
    return;
  }
  lastSentAt = now;

  if (WiFi.status() != WL_CONNECTED) {
    Serial.println("WiFi disconnected, reconnecting...");
    connectWiFi();
    return;
  }

  IPAddress targetIP;
  if (!resolveTarget(targetIP)) {
    Serial.println("Unable to resolve ping target");
    return;
  }

  const ProbeResult probe = runProbe(targetIP);
  Serial.printf("Probe target=%s sent=%d received=%d loss=%.2f latency=%.2fms\n",
                gConfig.pingTarget.c_str(), probe.sentPackets,
                probe.receivedPackets,
                probe.packetLossPercent, probe.averageLatencyMs);

  postStats(probe);
}
