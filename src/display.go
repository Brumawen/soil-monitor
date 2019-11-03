package main

import (
	"fmt"
	"time"

	gopitools "github.com/brumawen/gopi-tools/src"
)

// Display manages the display of information on the 2x8 character LCD display
type Display struct {
	Items     []*DisplayItem // The list of display items
	ShowTime  int            // The amount of time (in secs) that each screen will display
	isRunning bool           // Indicates if the display is running
	idx       int            // The current index
}

// DisplayItem item holds the text that will be displayed on the LCD display
type DisplayItem struct {
	Name  string // The name of the item
	Line1 string // Line 1 of the display
	Line2 string // Line 2 of the display
}

// Start will start the display cycling through the screens
func (d *Display) Start() {
	if d.isRunning {
		return
	}
	d.isRunning = true
	go func() {
		if d.ShowTime <= 0 {
			d.ShowTime = 5
		}
		d.idx = -1
		for d.isRunning {
			d.ShowNextItem()
			time.Sleep(time.Duration(d.ShowTime) * time.Second)
		}
	}()
}

// Stop will stop the display cycling through the screens.
func (d *Display) Stop() {
	d.isRunning = false
	cd := gopitools.CharDisplay{}
	if err := cd.Clear(); err != nil {
		d.logError("Error clearing the display.", err.Error())
	}
}

// SetItem sets the display item ready to be displayed
func (d *Display) SetItem(name string, l1 string, l2 string) {
	for _, x := range d.Items {
		if x.Name == name {
			x.Line1 = l1
			x.Line2 = l2
			d.RefreshCurrentItem()
			return
		}
	}
	d.Items = append(d.Items, &DisplayItem{
		Name:  name,
		Line1: l1,
		Line2: l2,
	})
}

// ShowNextItem increments the display index and shows the next item.
func (d *Display) ShowNextItem() {
	d.idx = d.idx + 1
	l := len(d.Items)
	if d.idx >= l {
		d.idx = 0
	}
	d.RefreshCurrentItem()
}

// RefreshCurrentItem displays the item associtated with the current display index.
func (d *Display) RefreshCurrentItem() {
	i := d.Items[d.idx]

	cd := gopitools.CharDisplay{}
	if err := cd.Message(i.GetMessage()); err != nil {
		d.logError("Error setting display.", err.Error())
	}
}

// GetMessage will return the display lines formatted for the Character Display
func (i *DisplayItem) GetMessage() string {
	if i.Line1 == "" && i.Line2 == "" {
		return ""
	}
	if i.Line2 == "" {
		return i.Line1
	}
	return i.Line1 + "\n" + i.Line2
}

// logError logs an error message to the logger
func (d *Display) logError(v ...interface{}) {
	a := fmt.Sprint(v...)
	logger.Error("Display: ", a[1:len(a)-1])
}
