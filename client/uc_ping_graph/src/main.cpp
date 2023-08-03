#include <Arduino.h>
#include <ArduinoJSON.h>
#include <ESP8266HTTPClient.h>
#include <ESP8266Ping.h>
#include <ESP8266WiFi.h>

const char *ssid = "";
const char *password = "";

const IPAddress remote_ip(1, 1, 1, 1);

void setup() {
  // put your setup code here, to run once:
  Serial.begin(9600);
  delay(10);
  Serial.println();
  Serial.println("Connecting to WiFi");

  WiFi.begin(ssid, password);

  while (WiFi.status() != WL_CONNECTED) {
    delay(100);
    Serial.print(".");
  }

  Serial.println();
  Serial.print("WiFi connected with ip ");
  Serial.println(WiFi.localIP());

  Serial.print("Pinging ip ");
  Serial.println(remote_ip);

  if (Ping.ping(remote_ip)) {
    Serial.println("Success!!");
  } else {
    Serial.println("Error :(");
  }
}

void loop() {
  // put your main code here, to run repeatedly:
  Ping.ping(remote_ip);
  int avg_time_ms = Ping.averageTime();
  Serial.println(avg_time_ms);

  const char *serverUrl = "https://postman-echo.com/post";
  const int capacity = JSON_OBJECT_SIZE(3);
  StaticJsonDocument<capacity> body;

  body["ping"] = avg_time_ms;
  body["ip"] = WiFi.localIP().toString();
  body["mac"] = WiFi.macAddress();

  // send https post request
  HTTPClient http;
  WiFiClientSecure client;

  http.begin(client, serverUrl);
  // http.addHeader("Content-Type", "application/json");
  // http.addHeader("Accept", "application/json");
  //   http.addHeader("User-Agent", "ESP8266");
  //    http.addHeader("Connection", "close");
  //     http.addHeader("Content-Length",
  //                    measureJson(body) + 1); // +1 for null terminator

  int httpResponseCode = http.POST(body.as<String>());

  if (httpResponseCode > 0) {
    Serial.print("HTTP Response code: ");
    Serial.println(httpResponseCode);
    String response = http.getString();
    Serial.println(response);
  } else {
    Serial.print("Error code: ");
    Serial.println(httpResponseCode);
    // print http response
    Serial.println(http.getString());
    Serial.println(http.errorToString(httpResponseCode).c_str());
  }
  http.end();
}
