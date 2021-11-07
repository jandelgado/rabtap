// rabtap templating
// Copyright (C) 2017-2021 Jan Delgado
package main

import (
	"bytes"
	"text/template"
)

// MergeTemplateFuncs merges the given list of template funcs into a single one
func MergeTemplateFuncs(maps ...template.FuncMap) template.FuncMap {
	res := template.FuncMap{}
	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// resolveTemplate resolves a template for use in the broker info printer,
// with support for colored output. name is just an informational name
// passed to the template ctor. tpl is the actual template and args
// the arguments used during rendering.
func resolveTemplate(name string, tpl string, args interface{}, funcs map[string]interface{}) string {
	tmpl := template.Must(template.New(name).Funcs(funcs).Parse(tpl))
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, args)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
