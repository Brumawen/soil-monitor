package main

import (
	"fmt"
	"time"
	"os"

	gopitools "github.com/brumawen/gopi-tools/src"
)

func main() {
	fmt.Println("Turning on power")
	pwr := gopitools.Pin{GpioNo: 22, TurnOffOnClose: true}
	defer pwr.Close()
	if err := pwr.On(); err != nil {
		fmt.Println("Error turning on power.", err.Error())
		return
	}

	// wait 2 secs to let everthing stabilize
	time.Sleep(2 * time.Second)

	// Read ambient light and moisture content
	fmt.Println("Reading Light and Moisture values")
	mcp := gopitools.Mcp3008{}
	defer mcp.Close()
	
	l := []float64{}
	m := []float64{}
	for i := 0; i < 50; i++ {
		vals, err := mcp.Read()
		if err != nil {
			fmt.Println("Error reading values.", i)
			break
		} 
		l = append(l, 100 - vals[1])
		m = append(m, vals[0])
	}

	printLines("light.dat", l)
	printLines("moisture.dat", m)

	// Switch off the power to the soil components
	fmt.Println("Turning off power")
	if err := pwr.Off(); err != nil {
		fmt.Println("Error turning off power.", err.Error())
	}
	fmt.Println("Done")
}

func printLines(filePath string, values []float64) error {
    f, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer f.Close()
    for _, value := range values {
       fmt.Fprintln(f, value)  // print values to f, one per line
    }
    return nil
}