#include <Arduino.h>
#include <ArduinoJSON.h>
#include <ESP8266HTTPClient.h>
#include <ESP8266Ping.h>
#include <ESP8266WiFi.h>
#include <WiFiClientSecureBearSSL.h>

#define LED D0

const char *ssid = "";
const char *password = "";

const IPAddress remote_ip(1, 1, 1, 1);

void setup() {
  // put your setup code here, to run once:
  Serial.begin(9600);
  Serial.setDebugOutput(true);
  pinMode(LED_BUILTIN, OUTPUT);
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

  Serial.print("Pinging ip ");
  Serial.println(remote_ip);

  if (Ping.ping(remote_ip)) {
    Serial.println("Success!!");
  } else {
    Serial.println("Error :(");
  }
}

void loop() {

  if ((WiFi.status() == WL_CONNECTED)) {
    std::unique_ptr<BearSSL::WiFiClientSecure> client(
        new BearSSL::WiFiClientSecure);
    client->setInsecure();
    HTTPClient https;
    digitalWrite(LED_BUILTIN, HIGH);

    Ping.ping(remote_ip);
    int avg_time_ms = Ping.averageTime();
    Serial.println(avg_time_ms);

    const char *serverUrl = "https://postman-echo.com/get?foo1=bar1&foo2=bar2";

    if (https.begin(*client, serverUrl)) { // HTTPS
      Serial.print("[HTTPS] GET...\n");
      // start connection and send HTTP header
      int httpCode = https.GET();
      // httpCode will be negative on error
      if (httpCode > 0) {
        // HTTP header has been send and Server response header has been handled
        Serial.printf("[HTTPS] GET... code: %d\n", httpCode);
        // file found at server
        if (httpCode == HTTP_CODE_OK ||
            httpCode == HTTP_CODE_MOVED_PERMANENTLY) {
          String payload = https.getString();
          Serial.println(payload);
        }
      } else {
        Serial.printf("[HTTPS] GET... failed, error: %s\n",
                      https.errorToString(httpCode).c_str());
      }

      https.end();
      digitalWrite(LED_BUILTIN, LOW);
    } else {
      Serial.printf("[HTTPS] Unable to connect\n");
    }
  }

  // const char *serverUrl = "https://postman-echo.com/post";
  // const char *serverUrl = "https://postman-echo.com/get?foo1=bar1&foo2=bar2";
  // const int capacity = JSON_OBJECT_SIZE(3);
  //  StaticJsonDocument<capacity> body;

  // body["ping"] = avg_time_ms;
  // body["ip"] = WiFi.localIP().toString();
  // body["mac"] = WiFi.macAddress();

  // send https post request
  // HTTPClient https;
  // WiFiClientSecure client;

  // https->setInsecure();

  // http.begin(client, serverUrl);
  //  http.addHeader("Content-Type", "application/json");
  //  http.addHeader("Accept", "application/json");
  //    http.addHeader("User-Agent", "ESP8266");
  //     http.addHeader("Connection", "close");
  //      http.addHeader("Content-Length",
  //                     measureJson(body) + 1); // +1 for null terminator

  // int httpResponseCode = http.POST(body.as<String>());

  // Serial.println(body.as<String>());

  // if (httpResponseCode > 0) {
  //   Serial.print("HTTP Response code: ");
  //   Serial.println(httpResponseCode);
  //   String response = http.getString();
  //   Serial.println(response);
  // } else {
  //   Serial.print("Error code: ");
  //   Serial.println(httpResponseCode);
  //   // print http response
  //   Serial.println(http.getString());
  //   Serial.println(http.errorToString(httpResponseCode).c_str());
  //}

  // http GET request
  // https.begin(client, serverUrl);
  // int httpResponseCode = https.GET();
  // Serial.println(httpResponseCode);
  // Serial.println(https.errorToString(httpResponseCode).c_str());

  // https.end();
}
