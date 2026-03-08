package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Transport handles reading/writing JSON-RPC messages over stdio pipes.
type Transport struct {
	writer io.Writer
	reader *bufio.Scanner
	mu     sync.Mutex
}

func NewTransport(r io.Reader, w io.Writer) *Transport {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer
	return &Transport{
		writer: w,
		reader: scanner,
	}
}

// Send writes a JSON-RPC message (one line of JSON + newline).
func (t *Transport) Send(msg any) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')
	_, err = t.writer.Write(data)
	return err
}

// Receive reads the next JSON-RPC message from the stream.
func (t *Transport) Receive() (json.RawMessage, error) {
	if !t.reader.Scan() {
		if err := t.reader.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	return json.RawMessage(t.reader.Bytes()), nil
}
