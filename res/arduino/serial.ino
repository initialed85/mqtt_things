void handleSerialRead() {
  if (!Serial.available()) {
    return;
  }

  int relayNumber = Serial.parseInt();

  if (relayNumber == 0) {
    Serial.println("ERROR: expected int, got something else");
    return;
  }

  int relayIndex = getRelayIndexFromNumber(relayNumber);
  if (relayIndex == -1) {
    Serial.print("ERROR: relay ");
    Serial.print(relayNumber);
    Serial.println(" unknown");
    return;
  }

  char readBuffer[16];

  Serial.readBytesUntil(',', readBuffer, 16);
  int bytesRead = Serial.readBytesUntil('\n', readBuffer, 16);
  readBuffer[bytesRead] = '\x0';
  String readData = String(readBuffer);
  readData.trim();
  readData.toLowerCase();

  int relayState = -1;
  String relayFriendlyState;

  if (readData.startsWith("on")) {
    relayState = 1;
    relayFriendlyState = "on";
  } else if (readData.startsWith("off")) {
    relayState = 0;
    relayFriendlyState = "off";
  }

  if (relayState == -1) {
    Serial.print("ERROR: state \"");
    Serial.print(readData);
    Serial.println("\" unknown- expected \"on\" of \"off\"");
    return;
  }

  Serial.print("INFO: requesting relay ");
  Serial.print(relayNumber);
  Serial.print(" change to state ");
  Serial.print(relayFriendlyState);
  Serial.println("");
  setRelayState(relayIndex, relayState);
}
