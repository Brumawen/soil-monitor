from gpiozero import MCP3008

light = MCP3008(0)
moisture = MCP3008(1)

print(str(light.value) + "," +  str(moisture.value))