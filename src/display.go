package main

import gopitools "github.com/brumawen/gopi-tools/src"

// Display manages the display of information on the 2x8 character LCD display
type Display struct {
	Idx   int                   // The current index
	Items []DisplayItem         // The list of display items
	Disp  gopitools.CharDisplay // The character display module
}

// DisplayItem item holds the text that will be displayed on the LCD display
type DisplayItem struct {
	Name  string // The name of the item
	Line1 string // Line 1 of the display
	Line2 string // Line 2 of the display
}

// SetItem sets the display item ready to be displayed
func (d *Display) SetItem(i DisplayItem) {
	for _, x := range d.Items {
		if x.Name == i.Name {
			x.Line1 = i.Line1
			x.Line2 = i.Line2
			d.RefreshCurrentItem()
			break
		}
	}
}

// ShowNextItem increments the display index and shows the next item.
func (d *Display) ShowNextItem() {
	d.Idx = d.idx + 1
	l := len(d.Items)
	if d.Idx >= l {
		d.Idx = 0
	}
	d.RefreshCurrentItem()
}

// RefreshCurrentItem displays the item associtated with the current display index.
func (d *Display) RefreshCurrentItem() {

}
