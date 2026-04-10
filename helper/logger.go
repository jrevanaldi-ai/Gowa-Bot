package helper

import (
	"fmt"
	"time"
)


const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorWhite  = "\033[97m"


	ColorBoldRed    = "\033[1;31m"
	ColorBoldGreen  = "\033[1;32m"
	ColorBoldYellow = "\033[1;33m"
	ColorBoldBlue   = "\033[1;34m"
	ColorBoldPurple = "\033[1;35m"
	ColorBoldCyan   = "\033[1;36m"


	ColorBgRed    = "\033[41m"
	ColorBgGreen  = "\033[42m"
	ColorBgYellow = "\033[43m"
	ColorBgBlue   = "\033[44m"
)


type Logger struct {
	prefix string
}


func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}


func formatTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}


func (l *Logger) Info(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	fmt.Printf("%s[%s] %s[INFO]%s %s\n", ColorGray, formatTimestamp(), ColorCyan, ColorReset, msg)
}


func (l *Logger) Success(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	fmt.Printf("%s[%s] %s[SUCCESS]%s %s\n", ColorGray, formatTimestamp(), ColorBoldGreen, ColorReset, msg)
}


func (l *Logger) Warning(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	fmt.Printf("%s[%s] %s[WARNING]%s %s\n", ColorGray, formatTimestamp(), ColorBoldYellow, ColorReset, msg)
}


func (l *Logger) Error(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	fmt.Printf("%s[%s] %s[ERROR]%s %s\n", ColorGray, formatTimestamp(), ColorBoldRed, ColorReset, msg)
}


func (l *Logger) Debug(format string, v ...interface{}) {
	msg := format
	if len(v) > 0 {
		msg = fmt.Sprintf(format, v...)
	}
	fmt.Printf("%s[%s] %s[DEBUG]%s %s\n", ColorGray, formatTimestamp(), ColorPurple, ColorReset, msg)
}


func (l *Logger) Message(pushName, number, cmd, chatType string) {
	fmt.Printf("%s[%s] %s[MESSAGE]%s %s | %s | %s | %s\n",
		ColorGray, formatTimestamp(), ColorBoldCyan, ColorReset,
		ColorWhite+pushName+ColorReset,
		ColorYellow+number+ColorReset,
		ColorGreen+cmd+ColorReset,
		ColorBlue+chatType+ColorReset)
}


func Banner() {
	fmt.Printf(`
%s╔═══════════════════════════════════════════╗
║%s  ██████╗  ██████╗ ██╗  ██╗██╗   ██╗%s
║  ██╔══██╗██╔═══██╗╚██╗██╔╝╚██╗ ██╔╝%s
║  ██████╔╝██║   ██║ ╚███╔╝  ╚████╔╝ %s
║  ██╔══██╗██║   ██║ ██╔██╗   ╚██╔╝  %s
║  ██████╔╝╚██████╔╝██╔╝ ██╗   ██║   %s
║  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝   ╚═╝   %s
║%s         Gowa-Bot - WhatsApp Bot         %s
║%s         Created with Gowa Library       %s
╚═══════════════════════════════════════════╝%s

`,
		ColorBoldCyan, ColorWhite, ColorReset,
		ColorBoldCyan, ColorReset,
		ColorBoldCyan, ColorReset,
		ColorBoldCyan, ColorReset,
		ColorBoldCyan, ColorReset,
		ColorBoldCyan, ColorReset,
		ColorCyan, ColorReset,
		ColorCyan, ColorReset,
		ColorReset)
}
