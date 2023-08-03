#include <Arduino.h>
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

  if (Ping.ping(remote_ip, 10)) {
    Serial.println("Success!!");
  } else {
    Serial.println("Error :(");
  }
  int avg_time_ms = Ping.averageTime();
  Serial.println(avg_time_ms);
}

void loop() {
  // put your main code here, to run repeatedly:
  if (Ping.ping(remote_ip)) {
    Serial.println("Success!!");
  } else {
    Serial.println("Error :(");
  }
  int avg_time_ms = Ping.averageTime();
  Serial.println(avg_time_ms);
}
