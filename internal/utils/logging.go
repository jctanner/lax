package utils

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"
)

// GetGoroutineID gets the ID of the current goroutine.
func GetGoroutineID() uint64 {
	b := make([]byte, 64)
	n := runtime.Stack(b, false)
	b = bytes.TrimPrefix(b[:n], []byte("goroutine "))
	idField := bytes.Fields(b)[0]
	id, err := strconv.ParseUint(string(idField), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

// GoroutineIDHook is a logrus hook that adds the goroutine ID to log entries.
type GoroutineIDHook struct{}

func (hook *GoroutineIDHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *GoroutineIDHook) Fire(entry *logrus.Entry) error {
	entry.Data["goroutine_id"] = GetGoroutineID()
	return nil
}
