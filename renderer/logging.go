package renderer

import (
	"go.uber.org/zap"
)

var zlog = zap.NewNop()

func SetLogger(l *zap.Logger) {
	zlog = l
}
