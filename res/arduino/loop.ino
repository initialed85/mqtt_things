int loopCounter = 100;

bool handleLoopCounter() {
  if (loopCounter == 0) {
    loopCounter = 100;
    return true;
  }

  loopCounter--;

  return false;
}

void loop() {
  if (handleLoopCounter()) {
    handleRelayCounters();
  }

  handleSerialRead();

  delay(10);
}
