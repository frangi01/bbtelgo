package logx

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	Debug Level = iota
	Info
	Warn
	Error
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
var (
	colReset  = "\x1b[0m"
	colGray   = "\x1b[90m"
	colBlue   = "\x1b[34m"
	colYellow = "\x1b[33m"
	colRed    = "\x1b[31m"
)

type Level int
type Options struct {
	Level      Level
	MaxSizeMB  int
	IncludeSrc bool
	TimeFormat string
}
type Logger struct {
	mu         sync.Mutex
	level      Level
	console    *log.Logger
	file       *log.Logger
	fileHandle *os.File
	filePath   string
	maxSize    int64
	includeSrc bool
	timeFormat string
}

func New(fp string, opts Options) (*Logger, error) {

	if opts.TimeFormat == "" {
		opts.TimeFormat = time.RFC3339
	}

	if err := os.MkdirAll(filepath.Dir(fp), 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	fh, err := os.OpenFile(fp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	l := &Logger{
		level:      opts.Level,
		console:    log.New(os.Stdout, "", 0),
		file:       log.New(fh, "", 0),
		fileHandle: fh,
		filePath:   fp,
		maxSize:    int64(opts.MaxSizeMB) * 1024 * 1024,
		includeSrc: opts.IncludeSrc,
		timeFormat: opts.TimeFormat,
	}
	return l, nil
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fileHandle != nil {
		return l.fileHandle.Close()
	}
	return nil
}

func (l *Logger) rotateIfNeeded() {
	if l.maxSize <= 0 || l.fileHandle == nil {
		return
	}
	info, err := l.fileHandle.Stat()
	if err != nil || info.Size() < l.maxSize {
		return
	}
	_ = l.fileHandle.Close()

	ts := time.Now().Format("20060102-150405")
	backup := fmt.Sprintf("%s.%s", l.filePath, ts)
	_ = os.Rename(l.filePath, backup)

	newFH, err := os.OpenFile(l.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		// If we fail to reopen, fallback to stderr to avoid panics.
		l.file = log.New(os.Stderr, "", 0)
		l.fileHandle = nil
		return
	}
	l.file = log.New(newFH, "", 0)
	l.fileHandle = newFH
}

func colorFor(lv Level) string {
	switch lv {
	case Debug:
		return colGray
	case Info:
		return colBlue
	case Warn:
		return colYellow
	case Error:
		return colRed
	default:
		return colReset
	}
}

func (l *Logger) prefix(lv Level) (fileline string, ts string, lvl string) {
	ts = time.Now().Format(l.timeFormat)
	lvl = lv.String()
	if l.includeSrc {
		// skip 3 frames: this, logf, public method
		if _, file, line, ok := runtime.Caller(3); ok {
			short := file
			if idx := strings.LastIndexByte(file, '/'); idx >= 0 {
				short = file[idx+1:]
			}
			fileline = fmt.Sprintf("%s:%d", short, line)
		}
	}
	return
}

func (l *Logger) logf(lv Level, format string, args ...any) {
	if lv < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	fileline, ts, lvl := l.prefix(lv)

	// Console (colored, compact)
	var cbuf strings.Builder
	cbuf.WriteString(colorFor(lv))
	cbuf.WriteString(ts)
	cbuf.WriteString(" ")
	cbuf.WriteString(lvl)
	if fileline != "" {
		cbuf.WriteString(" [")
		cbuf.WriteString(fileline)
		cbuf.WriteString("]")
	}
	cbuf.WriteString(colReset)
	cbuf.WriteString(" - ")
	fmt.Fprintf(&cbuf, format, args...)
	l.console.Print(cbuf.String())

	// File (no color, stable formatting)
	var fbuf strings.Builder
	fbuf.WriteString(ts)
	fbuf.WriteString(" ")
	fbuf.WriteString(lvl)
	if fileline != "" {
		fbuf.WriteString(" [")
		fbuf.WriteString(fileline)
		fbuf.WriteString("]")
	}
	fbuf.WriteString(" - ")
	fmt.Fprintf(&fbuf, format, args...)
	l.file.Print(fbuf.String())

	l.rotateIfNeeded()
}

func (l *Logger) Writer(lv Level) io.Writer {
	pr, pw := io.Pipe()
	go func() {
		defer pr.Close()
		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if n > 0 {
				msg := strings.TrimRight(string(buf[:n]), "\n")
				l.logf(lv, "%s", msg)
			}
			if err != nil {
				return
			}
		}
	}()
	return pw
}

func (l *Logger) Debugf(format string, args ...any) { l.logf(Debug, format, args...) }
func (l *Logger) Infof(format string, args ...any)  { l.logf(Info, format, args...) }
func (l *Logger) Warnf(format string, args ...any)  { l.logf(Warn, format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.logf(Error, format, args...) }

func (l *Logger) SetLevel(lv Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = lv
}

func (l *Logger) SetIncludeSrc(on bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.includeSrc = on
}

func (l *Logger) SetTimeFormat(tf string) {
	if tf == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timeFormat = tf
}

func (l *Logger) DisableFile() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fileHandle != nil {
		_ = l.fileHandle.Close()
		l.fileHandle = nil
	}
	l.file = log.New(io.Discard, "", 0)
}

func (l *Logger) EnableFile(path string, maxSizeMB int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if path == "" {
		path = l.filePath
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	fh, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	if l.fileHandle != nil {
		_ = l.fileHandle.Close()
	}
	l.fileHandle = fh
	l.filePath = path
	l.maxSize = int64(maxSizeMB) * 1024 * 1024
	l.file = log.New(fh, "", 0)
	return nil
}

// HOW TO USE
/*
	logger, err := logx.New("logs/bot.log", logx.Options{
		Level:      logx.Debug, // minimum level to print
		MaxSizeMB:  10,         // rotate after ~10MB (0 to disable)
		IncludeSrc: true,       // show file:line
		// TimeFormat: time.RFC3339, // or customize
	})
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	logger.Debugf("debug message: %v", 123)
	logger.Infof("server starting on %s", ":8080")
	logger.Warnf("cache miss for key=%s", "user:42")
	logger.Errorf("cannot open connection: %v", errors.New("connection refused"))


	// 3) Example: pipe standard library log into our logger
	std := log.New(logger.Writer(logx.Info), "", 0)
	std.Println("this goes through our logger")

	// 4) Demo loop (simulate app running)
	for i := 0; i < 5; i++ {
		logger.Infof("tick %d", i)
		time.Sleep(1 * time.Millisecond)
	}

*/
