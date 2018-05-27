import Adafruit_CharLCD as LCD
import argparse
import string

parser = argparse.ArgumentParser(description='Control a LCD display.')
parser.add_argument('-a', dest='action', default='display', help='The action to perform. ("display" or "clear"")')
parser.add_argument('-l1', dest='line1', nargs='+', type=str, default=[], help='The text to display on line 1.')
parser.add_argument('-l2', dest='line2', nargs='+', type=str, default=[], help='The text to display on line 2.')
args = parser.parse_args()

# Raspberry Pi pin configuration:
lcd_rs        = 21  
lcd_en        = 20
lcd_d4        = 26
lcd_d5        = 19
lcd_d6        = 13
lcd_d7        = 6
lcd_backlight = 16


# Define LCD column and row size for 8x2 LCD.
lcd_columns = 8
lcd_rows    = 2

# Initialize the LCD using the pins above.
lcd = LCD.Adafruit_CharLCD(lcd_rs, lcd_en, lcd_d4, lcd_d5, lcd_d6, lcd_d7,
                           lcd_columns, lcd_rows, lcd_backlight)

if args.action == 'display':
    line1 = string.join(args.line1, ' ')
    line2 = string.join(args.line2, ' ')
    lcd.message(line1 + '\n' + line2)
else:
    lcd.clear()