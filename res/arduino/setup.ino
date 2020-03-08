const int numRelays = 8;
const int relayTimeout = 30;

int relayPins[numRelays] = {2, 3, 4, 5, 6, 7, 8, 9};
int relayNumbers[numRelays] = {1, 2, 3, 4, 5, 6, 7, 8};
int relayStates[numRelays] = {0, 0, 0, 0, 0, 0, 0, 0};
int relayCounters[numRelays] = {0, 0, 0, 0, 0, 0, 0, 0};

void setup() {
  for (int i = 0; i < numRelays; i++ ) {
    int pin = relayPins[i];
    pinMode(pin, OUTPUT);
    digitalWrite(pin, LOW);
  }

  Serial.begin(9600);
  Serial.setTimeout(10);
  Serial.println("DEBUG: setup() complete");
}