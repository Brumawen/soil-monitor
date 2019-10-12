import Adafruit_DHT

humidity, temperature = Adafruit_DHT.read_retry(Adafruit_DHT.DHT11, 17)
if humidity is not None and temperature is not None:
    print(str(temperature) + "," + str(humidity))
else:
    print('-1,-1')