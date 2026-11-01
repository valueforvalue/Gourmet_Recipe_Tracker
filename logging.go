package main

import (
	"io"
	"log"
	"os"
)

// logMaxBytes is the size at which tracker.log rolls over to tracker.log.1.
const logMaxBytes = 10 * 1024 * 1024

// rotatingLogWriter writes to a log file and rolls it over to <path>.1 when
// the next write would exceed maxBytes. The log package serializes writes, so
// no internal locking is needed.
type rotatingLogWriter struct {
	path     string
	maxBytes int64
	file     *os.File
	size     int64
}

// Write appends p to the current file, rotating first if the write would
// exceed maxBytes.
func (w *rotatingLogWriter) Write(p []byte) (int, error) {
	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}
	if w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Close closes the underlying file, if open.
func (w *rotatingLogWriter) Close() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *rotatingLogWriter) open() error {
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	w.file = f
	w.size = info.Size()
	return nil
}

func (w *rotatingLogWriter) rotate() error {
	if err := w.Close(); err != nil {
		return err
	}
	// os.Rename fails on Windows when the destination exists, so drop any
	// previous .1 backup first.
	_ = os.Remove(w.path + ".1")
	if err := os.Rename(w.path, w.path+".1"); err != nil {
		return err
	}
	w.size = 0
	return w.open()
}

// setupLogging directs the standard logger to both stdout and a rotating
// tracker.log. It returns the writer so the caller can close it on shutdown,
// or an error when the log file could not be opened.
func setupLogging() (*rotatingLogWriter, error) {
	w := &rotatingLogWriter{path: "tracker.log", maxBytes: logMaxBytes}
	if err := w.open(); err != nil {
		return nil, err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, w))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	return w, nil
}

// logInfo logs at INFO level.
func logInfo(format string, args ...any) { log.Printf("[INFO] "+format, args...) }

// logWarn logs at WARN level.
func logWarn(format string, args ...any) { log.Printf("[WARN] "+format, args...) }

// logError logs at ERROR level.
func logError(format string, args ...any) { log.Printf("[ERROR] "+format, args...) }
