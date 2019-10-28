package renderer

import (
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
)

var lock sync.Mutex

func PutPreambule(w http.ResponseWriter, title string) {
	PutLine(w, `<html><head><title>%s</title><link rel="stylesheet" type="text/css" href="/dfuse.css"></head><body><div style="width:90%%; margin: 2rem auto;"><h1>%s</h1>`, title, title)
}

func PutErrorLine(w http.ResponseWriter, prefix string, err error) {
	PutLine(w, "<p><strong>%s: %s</strong></p>\n", prefix, err.Error())
}

func PutSyncLine(w http.ResponseWriter, format string, v ...interface{}) {
	line := fmt.Sprintf(format, v...)

	flush := w.(http.Flusher)

	lock.Lock()
	defer lock.Unlock()

	fmt.Fprint(w, line)
	flush.Flush()

	zlog.Info("html output line", zap.String("line", line))
}

func PutLine(w http.ResponseWriter, format string, v ...interface{}) {
	line := fmt.Sprintf(format, v...)

	flush := w.(http.Flusher)
	fmt.Fprint(w, line)
	flush.Flush()

	zlog.Info("html output line", zap.String("line", line))
}

func FlushWriter(w http.ResponseWriter) {
	flusher := w.(http.Flusher)
	flusher.Flush()
}
