package logger

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func (l Level) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

type Entry struct {
	Time   string                 `json:"time"`
	Level  string                 `json:"level"`
	Msg    string                 `json:"msg"`
	ReqID  string                 `json:"req_id,omitempty"`
	Method string                 `json:"method,omitempty"`
	Path   string                 `json:"path,omitempty"`
	IP     string                 `json:"ip,omitempty"`
	Status int                    `json:"status,omitempty"`
	Dur    string                 `json:"dur,omitempty"`
	Err    string                 `json:"err,omitempty"`
	Source string                 `json:"src,omitempty"`
	Extra  map[string]interface{} `json:"-"`
}

type Logger struct {
	minLevel Level
}

func New() *Logger {
	return &Logger{minLevel: DebugLevel}
}

func generateRequestID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func (l *Logger) getSourceLocation(skip int) (string, int, string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", 0, "unknown"
	}

	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = fn.Name()
		if idx := strings.LastIndex(funcName, "."); idx != -1 {
			funcName = funcName[idx+1:]
		}
	}

	return file, line, funcName
}

func (l *Logger) log(level Level, message string, err error, ctx map[string]interface{}, skip int) {
	if level < l.minLevel {
		return
	}

	file, line, _ := l.getSourceLocation(skip + 1)
	entry := Entry{
		Time:   time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Level:  strings.ToLower(level.String()),
		Msg:    message,
		Source: fmt.Sprintf("%s:%d", strings.TrimSuffix(file, ".go"), line),
	}

	if ctx != nil {
		// Extract standard fields
		if v, ok := ctx["req_id"].(string); ok {
			entry.ReqID = v
		}
		if v, ok := ctx["method"].(string); ok {
			entry.Method = v
		}
		if v, ok := ctx["path"].(string); ok {
			entry.Path = v
		}
		if v, ok := ctx["ip"].(string); ok {
			entry.IP = v
		}
		if v, ok := ctx["status"].(int); ok {
			entry.Status = v
		}
		if v, ok := ctx["duration"].(time.Duration); ok {
			entry.Dur = v.String()
		}
		
		// Store all other fields in Extra map
		entry.Extra = make(map[string]interface{})
		standardFields := map[string]bool{
			"req_id":   true,
			"method":   true,
			"path":     true,
			"ip":       true,
			"status":   true,
			"duration": true,
		}
		for k, v := range ctx {
			if !standardFields[k] {
				entry.Extra[k] = v
			}
		}
		// If no extra fields, set to nil to avoid empty object
		if len(entry.Extra) == 0 {
			entry.Extra = nil
		}
	}

	if err != nil {
		entry.Err = err.Error()
	}

	// Use Go's log with no prefix for clean JSON output
	log.SetFlags(0)
	
	// Create a map for JSON output that includes all fields
	output := make(map[string]interface{})
	output["time"] = entry.Time
	output["level"] = entry.Level
	output["msg"] = entry.Msg
	output["src"] = entry.Source
	
	if entry.ReqID != "" {
		output["req_id"] = entry.ReqID
	}
	if entry.Method != "" {
		output["method"] = entry.Method
	}
	if entry.Path != "" {
		output["path"] = entry.Path
	}
	if entry.IP != "" {
		output["ip"] = entry.IP
	}
	if entry.Status != 0 {
		output["status"] = entry.Status
	}
	if entry.Dur != "" {
		output["dur"] = entry.Dur
	}
	if entry.Err != "" {
		output["err"] = entry.Err
	}
	
	// Add all extra fields
	if entry.Extra != nil {
		for k, v := range entry.Extra {
			output[k] = v
		}
	}
	
	if jsonData, marshalErr := json.Marshal(output); marshalErr == nil {
		log.Println(string(jsonData))
	} else {
		log.Printf(`{"time":"%s","level":"%s","msg":"%s"}`, entry.Time, entry.Level, message)
	}
}

func (l *Logger) WithRequest(r *http.Request) map[string]interface{} {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" && r.RemoteAddr != "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}

	return map[string]interface{}{
		"req_id": requestID,
		"path":   r.URL.Path,
		"method": r.Method,
		"ip":     ip,
	}
}

func (l *Logger) Debug(message string, ctx map[string]interface{}) {
	l.log(DebugLevel, message, nil, ctx, 1)
}

func (l *Logger) Info(message string, ctx map[string]interface{}) {
	l.log(InfoLevel, message, nil, ctx, 1)
}

func (l *Logger) Warn(message string, ctx map[string]interface{}) {
	l.log(WarnLevel, message, nil, ctx, 1)
}

func (l *Logger) Error(message string, err error, ctx map[string]interface{}) {
	l.log(ErrorLevel, message, err, ctx, 1)
}

var Log = New()

func Debug(message string, ctx map[string]interface{}) {
	Log.Debug(message, ctx)
}

func Info(message string, ctx map[string]interface{}) {
	Log.Info(message, ctx)
}

func Warn(message string, ctx map[string]interface{}) {
	Log.Warn(message, ctx)
}

func Error(message string, err error, ctx map[string]interface{}) {
	Log.Error(message, err, ctx)
}

func LogRequestStart(r *http.Request) {
	Log.Info("Request started", Log.WithRequest(r))
}

func LogRequestComplete(r *http.Request, statusCode int, duration time.Duration) {
	ctx := Log.WithRequest(r)
	ctx["status"] = statusCode
	ctx["duration"] = duration
	Log.Info("Request completed", ctx)
}

func LogRequestError(r *http.Request, err error, statusCode int) {
	ctx := Log.WithRequest(r)
	ctx["status"] = statusCode
	Log.Error("Request failed", err, ctx)
}
