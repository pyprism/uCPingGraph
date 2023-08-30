#include <Arduino.h>
#include <ArduinoJSON.h>
#include <ESP8266HTTPClient.h>
#include <ESP8266WiFi.h>
#include <ESPping.h>
#include <WiFiClientSecureBearSSL.h>

// update these with values suitable for your network and server
const char *ssid = "";
const char *password = "";
const String domain = "";
const String deviceToken = "";

const IPAddress remote_ip(1, 1, 1, 1);

void setup() {
  // put your setup code here, to run once:
  Serial.begin(9600);
  Serial.setDebugOutput(true);
  // pinMode(LED_BUILTIN, OUTPUT);
  delay(10);
  Serial.println();
  Serial.println("Connecting to WiFi");

  WiFi.mode(WIFI_STA);
  WiFi.begin(ssid, password);

  while (WiFi.status() != WL_CONNECTED) {
    delay(100);
    Serial.print(".");
  }

  Serial.println();
  Serial.print("WiFi connected with ip ");
  Serial.println(WiFi.localIP());
}

void sendPost(int latency) {
  String sUrl = domain + "/api/stats/";
  const char *serverUrl = sUrl.c_str();
  const int capacity = JSON_OBJECT_SIZE(2);
  StaticJsonDocument<capacity> body;
  body["latency"] = latency;

  // send https post request

  if ((WiFi.status() == WL_CONNECTED)) {
    std::unique_ptr<BearSSL::WiFiClientSecure> client(
        new BearSSL::WiFiClientSecure);
    client->setInsecure();
    HTTPClient https;
    if (https.begin(*client, serverUrl)) {
      https.addHeader("Content-Type", "application/json");
      https.addHeader("Accept", "application/json");
      https.addHeader("Authorization", deviceToken);
      int httpResponseCode = https.POST(body.as<String>());

      // Serial.println(body.as<String>());

      if (httpResponseCode >= 200 && httpResponseCode <= 300) {
        Serial.print("HTTP Response code: ");
        Serial.println(httpResponseCode);
        String response = https.getString();
        Serial.println(response);
      } else {
        Serial.print("Error code: ");
        Serial.println(httpResponseCode);
        // print http response
        Serial.println(https.getString());
        Serial.println(https.errorToString(httpResponseCode).c_str());
      }
      https.end();
    } else {
      Serial.println("Error in HTTPS connection");
    }
  }
}

int getPing() {

  if (Ping.ping(remote_ip, 2)) {
    int avg_time_ms = Ping.averageTime();
    Serial.println(avg_time_ms);
    return avg_time_ms;
  } else {
    Serial.println("Error");
    return -1;
  }
}

void loop() {

  int latency = getPing();
  if (latency == -1) {
    // TODO: LED blink
    // digitalWrite(LED_BUILTIN, LOW);
  } else {
    sendPost(latency);
  }
}
