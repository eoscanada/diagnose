package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
)

type tplData struct {
	Protocol           string
	Namespace          string
	BlockStore         string
	SearchIndexesStore string
	KVDBConnection     string
}

func (d *Diagnose) renderTemplate(w http.ResponseWriter, data *tplData) {
	cnt, err := ioutil.ReadFile("index.html")
	if err != nil {
		zlog.Fatal("failed reading index.html", zap.Error(err))
	}

	tpl, err := template.New("index.html").Funcs(template.FuncMap{
		"json": func(input interface{}) (template.JS, error) {
			cnt, err := json.MarshalIndent(input, "", "  ")
			if err != nil {
				return "", err
			}
			return template.JS(cnt), nil
		},
	}).Parse(string(cnt))

	if err != nil {
		zlog.Fatal("failed parsing template", zap.Error(err))
		//http.Error(w, fmt.Sprintf("error parsing template: %s", err), 500)
		return
	}

	err = tpl.Execute(w, data)

	if err != nil {
		http.Error(w, fmt.Sprintf("error processing template: %s", err), 500)
	}
}
