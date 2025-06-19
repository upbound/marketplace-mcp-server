package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

// Proxy represents a lightweight proxy to wrap the MCP server,
// allowing it to be reachable over HTTP
type Proxy struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.Reader
	mu     sync.Mutex
	reqID  int
}

func NewProxy(serverBinary string, args ...string) (*Proxy, error) {
	cmd := exec.Command(serverBinary, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd.Stderr = log.Writer()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &Proxy{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

func (p *Proxy) Forward(method string, params interface{}) ([]byte, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.reqID++

	var rawParams json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		rawParams = json.RawMessage(b)
	}
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      p.reqID,
		Method:  method,
		Params:  rawParams,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Write to open pipe
	_, err = p.stdin.Write(append(data, '\n'))
	if err != nil {
		return nil, err
	}

	// Read response from stdout pipe
	buf := make([]byte, 4096)
	n, err := p.stdout.Read(buf)
	if err != nil {
		return nil, err
	}

	// Return bytes and unmarshal in the handlers
	return buf[:n], nil
}
