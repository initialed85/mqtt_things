int getRelayPinFromIndex(int index) {
  return relayPins[index];
}

int getRelayNumberFromIndex(int index) {
  return relayNumbers[index];
}

int getRelayIndexFromNumber(int number) {
  for (int i = 0; i < numRelays; i++) {
    if (relayNumbers[i] == number) {
      return i;
    }
  }

  return -1;
}

bool setRelayState(int index, int state) {
  int *relayState = &relayStates[index];
  int *relayCounter = &relayCounters[index];
  int relayPin = relayPins[index];

  if (state == 1) {
    *relayCounter = relayTimeout;
    if (*relayState != state) {
      digitalWrite(relayPin, HIGH);
      *relayState = 1;
      return true;
    }
  } else if (*relayState != state) {
    *relayCounter = 0;
    digitalWrite(relayPin, LOW);
    *relayState = 0;
    return true;
  }

  return false;
}

void handleRelayCounters() {
  for (int i = 0; i < numRelays; i++) {
    int *relayCounter = &relayCounters[i];

    if (*relayCounter > 0) {
      *relayCounter = *relayCounter - 1;
    } else if (setRelayState(i, 0)) {
      Serial.print("INFO: requesting relay ");
      Serial.print(getRelayNumberFromIndex(i));
      Serial.println(" change to state off (due to timeout)");
    }
  }
}