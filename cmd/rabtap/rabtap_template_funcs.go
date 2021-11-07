// templating functions for rabtap
// (c) copyright 2021 by Jan Delgado
package main

import (
	"math"
	"text/template"
)

var RabtapTemplateFuncs = rabtapTemplateFuncs{}.GetFuncMap()

type rabtapTemplateFuncs struct {
}

// toPercent converts the given float value to a rounded int percentage value
func (s rabtapTemplateFuncs) toPercent(x float64) int {
	return int(math.Round(x * 100.))
}

func (s rabtapTemplateFuncs) GetFuncMap() template.FuncMap {
	return template.FuncMap{
		"ToPercent": s.toPercent,
	}
}
