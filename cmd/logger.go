package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/jwalton/gchalk"
)

func SleepyLogLn(format string, v ...any) {
	SleepyPrint("[info]", 2, format+"\n", v...)
}

func SleepyLog(format string, v ...any) {
	SleepyPrint("[info]", 2, format, v...)
}

func SleepyWarnLn(format string, v ...any) {
	SleepyPrint(gchalk.RGB(255, 136, 0)("[warn]"), 2, format+"\n", v...)
}

func SleepyWarn(format string, v ...any) {
	SleepyPrint(gchalk.RGB(255, 136, 0)("[warn]"), 2, format, v...)
}

func SleepyErrorLn(format string, v ...any) {
	SleepyPrint(gchalk.Red("[error]"), 2, format+"\n", v...)
}

func SleepyError(format string, v ...any) {
	SleepyPrint(gchalk.Red("[error]"), 2, format, v...)
}

func SleepyPrint(level string, depth int, format string, v ...any) {
	hour, min, sec := time.Now().Clock()
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}
	file = file[(strings.LastIndex(file, "/") + 1):]

	fmt.Printf("[%s] [%s] %s: %s", gchalk.Red(fmt.Sprintf("%02d:%02d:%02d", hour, min, sec)), gchalk.Blue(fmt.Sprintf("%s:%v", file, line)), level, fmt.Sprintf(format, v...))
}
