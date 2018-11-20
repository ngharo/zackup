package graylog

import (
	"sync"

	"github.com/sirupsen/logrus"
	graylog "gopkg.in/gemnasium/logrus-graylog-hook.v2"
)

var (
	log = logrus.WithField("prefix", "middleware")

	_ Middleware  = (*middleware)(nil) // type check
	_ logrus.Hook = (*middleware)(nil) // type check
)

// Middleware implements a runtime-configurable Graylog middleware. This
// allows you can change the log level and graylog endpoint at runtime.
// You need to call Flush() before exiting the program.
type Middleware interface {
	logrus.Hook
	Flush()
	SetLevel(string)
	SetEndpoint(string)
}

// NewMiddleware returns a new Middleware.
func NewMiddleware(componentName string) Middleware {
	return &middleware{name: componentName}
}

type middleware struct {
	name     string
	endpoint string
	level    logrus.Level
	hook     *graylog.GraylogHook

	sync.RWMutex
}

func (gl *middleware) Fire(ent *logrus.Entry) error {
	gl.RLock()
	defer gl.RUnlock()

	if gl.hook == nil || ent.Level < gl.level {
		return nil
	}
	return gl.hook.Fire(ent)
}

func (gl *middleware) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		// TraceLevel,
	}
}

func (gl *middleware) Flush() {
	gl.Lock()
	defer gl.Unlock()

	if gl.hook != nil {
		gl.hook.Flush()
	}
}

func (gl *middleware) SetLevel(s string) {
	gl.Lock()
	defer gl.Unlock()

	lvl, err := logrus.ParseLevel(s)
	if err == nil && lvl != gl.level {
		// log level has changed
		log.Level = lvl
		gl.level = lvl
	}
}

func (gl *middleware) SetEndpoint(s string) {
	gl.Lock()
	defer gl.Unlock()

	if gl.endpoint == s {
		// no change
		return
	}

	if s == "" {
		// disable endpoint
		gl.hook = nil
	}

	newHook := graylog.NewAsyncGraylogHook(s, map[string]interface{}{
		"component": gl.name,
	})
	if gl.hook == nil {
		// install endpoint
		gl.hook = newHook
	} else {
		// update endpoint
		oldHook := gl.hook
		gl.hook = newHook
		oldHook.Flush()
	}
}
