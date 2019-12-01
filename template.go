package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"go.uber.org/zap"
)

type tplData struct {
	Router             *mux.Router
	Protocol           string
	Namespace          string
	BlockStore         string
	SearchIndexesStore string
	KVDBConnection     string
}

func (d *RootServer) renderTemplate(w http.ResponseWriter, data *tplData) {
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
		"isEOS": func(protocol string) bool {
			return true
		},
		"isETH": func(protocol string) bool {
			return false
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

func (t *tplData) IsEOS() bool {
	return t.Protocol == "EOS"
}

func (t *tplData) IsETH() bool {
	return t.Protocol == "ETH"
}
