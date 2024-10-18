package logger

import (
	"fmt"
)

// ANSI renk kodları
const (
	Reset   = "\033[0m"
	Blue    = "\033[1;34m" // Açık mavi
	Green   = "\033[1;32m" // Açık yeşil
	Yellow  = "\033[1;33m" // Açık sarı
	Red     = "\033[1;31m" // Açık kırmızı
	Cyan    = "\033[1;36m" // Yumuşak cyan
	Magenta = "\033[1;35m" // Yumuşak mor
)

// Log prints a message with a tag and color.
func Log(color string, tag string, message string) {
	fmt.Printf("%s[%s] %s%s\n", color, tag, message, Reset)
}

func Info(message string) {
	Log(Cyan, "INFO", message)
}

func Success(message string) {
	Log(Green, "SUCCESS", message)
}

func Warning(message string) {
	Log(Yellow, "WARNING", message)
}

func Error(message string) {
	Log(Red, "ERROR", message)
}
