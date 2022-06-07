package action

import (
	"container/ring"
	"fmt"
	"github.com/fluxcd/pkg/runtime/logger"
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"strings"
	"sync"
)

const defaultBufferSize = 5

func DebugLogr(log logr.Logger) action.DebugLog {
	return func(format string, v ...interface{}) {
		log.V(logger.InfoLevel).Info(fmt.Sprintf(format, v...))
	}
}

type LogBuffer struct {
	mu     sync.Mutex
	log    action.DebugLog
	buffer *ring.Ring
}

func NewLogBuffer(log action.DebugLog, size int) *LogBuffer {
	if size <= 0 {
		size = defaultBufferSize
	}
	return &LogBuffer{
		log:    log,
		buffer: ring.New(size),
	}
}

func (l *LogBuffer) Log(format string, v ...interface{}) {
	l.mu.Lock()

	// Filter out duplicate log lines, this happens for example when
	// Helm is waiting on workloads to become ready.
	msg := fmt.Sprintf(format, v...)
	if prev := l.buffer.Prev(); prev.Value != msg {
		l.buffer.Value = msg
		l.buffer = l.buffer.Next()
	}

	l.mu.Unlock()
	l.log(format, v...)
}

func (l *LogBuffer) Reset() {
	l.mu.Lock()
	l.buffer = ring.New(l.buffer.Len())
	l.mu.Unlock()
}

func (l *LogBuffer) String() string {
	var str string
	l.mu.Lock()
	l.buffer.Do(func(s interface{}) {
		if s == nil {
			return
		}
		str += s.(string) + "\n"
	})
	l.mu.Unlock()
	return strings.TrimSpace(str)
}
