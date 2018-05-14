from gpiozero import MCP3008
from time import sleep

values = []
# retrieve each value from the ADC pins
for i in range(0,7):
    dev = MCP3008(i)
    values.append(dev.value*100)

# return the values to the calling function
print("\t".join(map(str,values)))

   
