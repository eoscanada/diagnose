package eos

import (
	"html/template"

	"github.com/eoscanada/derr"
)

var alreadyRunningTemplate *template.Template

func init() {
	var err error

	alreadyRunningTemplate, err = template.New("already_running").Parse(alreadyRunning)
	derr.Check("unable to create template already_running", err)
}

var alreadyRunning = `
<html>
<head>
    <title>dfuse diagnose</title>
    <link rel="stylesheet" type="text/css" href="/dfuse.css">
</head>
<body>
    <div style="width:90%; margin: 2rem auto;">
		<h2>Already running, try later</h2>
    </div>
</body>
</html>
`
