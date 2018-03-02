// Copyright (c) 2017 Jan Delgado
package main

// print colors on the console, supporting golang templating

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

const (
	colorURL      = color.FgHiWhite
	colorVHost    = color.FgMagenta
	colorExchange = color.FgHiBlue
	colorQueue    = color.FgHiYellow
	colorConsumer = color.FgHiGreen
	colorMessage  = color.FgHiYellow
	colorKey      = color.FgHiCyan
)

// ColorPrinterFunc takes fmt.Sprint like arguments and add colors
type ColorPrinterFunc func(a ...interface{}) string

// ColorPrinter allows to print various items colorized
type ColorPrinter struct {
	URL      ColorPrinterFunc
	VHost    ColorPrinterFunc
	Exchange ColorPrinterFunc
	Queue    ColorPrinterFunc
	Consumer ColorPrinterFunc
	Message  ColorPrinterFunc
	Key      ColorPrinterFunc
}

// GetFuncMap returns a function map that can be used in a template.
func (s ColorPrinter) GetFuncMap() map[string]interface{} {
	return map[string]interface{}{
		"QueueColor":    s.Queue,
		"ExchangeColor": s.Exchange,
		"URLColor":      s.URL,
		"VHostColor":    s.VHost,
		"ConsumerColor": s.Consumer,
		"MessageColor":  s.Message,
		"KeyColor":      s.Key}
}

// NewColorableWriter returns a colorable writer for the given file (e.g
// os.Stdout). Needed for windows platform to make colors work.
func NewColorableWriter(file *os.File) io.Writer {
	return colorable.NewColorable(file)
}

// NewColorPrinter returns a ColorPrinter used to color the console
// output. If noColor is set to true, a no-op color printer is returned.
func NewColorPrinter(noColor bool) ColorPrinter {
	if noColor {
		nullPrinter := func(a ...interface{}) string {
			return fmt.Sprint(a...)
		}
		return ColorPrinter{nullPrinter, nullPrinter,
			nullPrinter, nullPrinter, nullPrinter, nullPrinter, nullPrinter}
	}
	return ColorPrinter{
		URL:      color.New(colorURL).SprintFunc(),
		VHost:    color.New(colorVHost).SprintFunc(),
		Exchange: color.New(colorExchange).SprintFunc(),
		Queue:    color.New(colorQueue).SprintFunc(),
		Consumer: color.New(colorConsumer).SprintFunc(),
		Message:  color.New(colorMessage).SprintFunc(),
		Key:      color.New(colorKey).SprintFunc(),
	}
}
