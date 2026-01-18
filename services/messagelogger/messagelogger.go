package messagelogger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/appnet-org/arpc/pkg/rpc/element"
)

// ClientMessageLogger implements RPC element interface for client-side message logging
type ClientMessageLogger struct {
	file     *os.File
	filename string
	logDir   string
	mu       sync.Mutex
}

// ServerMessageLogger implements RPC element interface for server-side message logging
type ServerMessageLogger struct {
	file     *os.File
	filename string
	logDir   string
	mu       sync.Mutex
}

// NewClientMessageLogger creates a new client-side message logging element
func NewClientMessageLogger(serviceName string) (element.RPCElement, error) {
	logDir := os.Getenv("MESSAGE_LOG_DIR")
	if logDir == "" {
		logDir = "/var/log/arpc-messages"
	}

	// Create log file name with service name and timestamp but don't open file yet (lazy initialization)
	filename := filepath.Join(logDir, fmt.Sprintf("client-messages-%s-%s.jsonl", serviceName, time.Now().Format("20060102-150405")))

	log.Printf("Client message logging will be written to: %s", filename)
	return &ClientMessageLogger{
		filename: filename,
		logDir:   logDir,
	}, nil
}

// NewServerMessageLogger creates a new server-side message logging element
func NewServerMessageLogger(serviceName string) (element.RPCElement, error) {
	logDir := os.Getenv("MESSAGE_LOG_DIR")
	if logDir == "" {
		logDir = "/var/log/arpc-messages"
	}

	// Create log file name with service name and timestamp but don't open file yet (lazy initialization)
	filename := filepath.Join(logDir, fmt.Sprintf("server-messages-%s-%s.jsonl", serviceName, time.Now().Format("20060102-150405")))

	log.Printf("Server message logging will be written to: %s", filename)
	return &ServerMessageLogger{
		filename: filename,
		logDir:   logDir,
	}, nil
}

// ClientMessageLogger methods
func (l *ClientMessageLogger) Name() string {
	return "client-message-logger"
}

// ensureFileOpen opens the log file if it's not already open (lazy initialization)
func (l *ClientMessageLogger) ensureFileOpen() error {
	if l.file != nil {
		return nil
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open the log file
	file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	log.Printf("Client message log file created: %s", l.filename)
	return nil
}

func (l *ClientMessageLogger) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, context.Context, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Lazy-create the log file on first write
	if err := l.ensureFileOpen(); err != nil {
		log.Printf("Failed to open log file: %v", err)
		return req, ctx, nil // Don't fail the RPC call
	}

	// Compute serialization sizes
	sizes := ComputeSizes(req.Payload)

	// Create log entry
	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Direction:   "request",
		Method:      req.Method,
		MessageType: GetMessageTypeName(req.Payload),
		Sizes:       sizes,
		Payload:     req.Payload,
	}

	data, err := MarshalLogEntry(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry to JSON: %v", err)
		return req, ctx, nil // Don't fail the RPC call
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		log.Printf("Failed to write request to log file: %v", err)
	}

	return req, ctx, nil
}

func (l *ClientMessageLogger) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, context.Context, error) {
	// Don't log receives - only log sends
	return resp, ctx, nil
}

func (l *ClientMessageLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// ServerMessageLogger methods
func (l *ServerMessageLogger) Name() string {
	return "server-message-logger"
}

// ensureFileOpen opens the log file if it's not already open (lazy initialization)
func (l *ServerMessageLogger) ensureFileOpen() error {
	if l.file != nil {
		return nil
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(l.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open the log file
	file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	l.file = file
	log.Printf("Server message log file created: %s", l.filename)
	return nil
}

func (l *ServerMessageLogger) ProcessRequest(ctx context.Context, req *element.RPCRequest) (*element.RPCRequest, context.Context, error) {
	// Don't log receives - only log sends
	return req, ctx, nil
}

func (l *ServerMessageLogger) ProcessResponse(ctx context.Context, resp *element.RPCResponse) (*element.RPCResponse, context.Context, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Lazy-create the log file on first write
	if err := l.ensureFileOpen(); err != nil {
		log.Printf("Failed to open log file: %v", err)
		return resp, ctx, nil // Don't fail the RPC call
	}

	var entry LogEntry

	if resp.Error != nil {
		// Log error response without size computation
		entry = LogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Direction:   "response",
			MessageType: "error",
			Payload:     map[string]string{"error": resp.Error.Error()},
		}
	} else {
		// Compute serialization sizes for successful response
		sizes := ComputeSizes(resp.Result)

		entry = LogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			Direction:   "response",
			MessageType: GetMessageTypeName(resp.Result),
			Sizes:       sizes,
			Payload:     resp.Result,
		}
	}

	data, err := MarshalLogEntry(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry to JSON: %v", err)
		return resp, ctx, nil // Don't fail the RPC call
	}

	if _, err := l.file.Write(append(data, '\n')); err != nil {
		log.Printf("Failed to write response to log file: %v", err)
	}

	return resp, ctx, nil
}

func (l *ServerMessageLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
