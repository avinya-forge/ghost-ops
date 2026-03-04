package sidecar

import (
	"context"
	"io"
	"log"
	"net"
	"sync"
)

// Proxy represents a sidecar proxy for network interception.
type Proxy struct {
	targetAddr string
	listener   net.Listener
	mu         sync.Mutex
	active     bool
}

// NewProxy creates a new sidecar proxy.
func NewProxy(targetAddr string) *Proxy {
	return &Proxy{
		targetAddr: targetAddr,
	}
}

// Start starts the proxy on the given address.
func (p *Proxy) Start(ctx context.Context, listenAddr string) error {
	var err error
	p.mu.Lock()
	p.listener, err = net.Listen("tcp", listenAddr)
	if err != nil {
		p.mu.Unlock()
		return err
	}
	p.active = true

	// Create a channel to signal when the listener actually stops
	stopCh := make(chan struct{})
	p.mu.Unlock()

	go func() {
		select {
		case <-ctx.Done():
			p.Stop()
		case <-stopCh:
			// Proxy stopped manually
		}
	}()

	for {
		clientConn, err := p.listener.Accept()
		if err != nil {
			p.mu.Lock()
			active := p.active
			p.mu.Unlock()
			if !active {
				close(stopCh)
				return nil
			}
			log.Printf("Proxy accept error: %v", err)
			continue
		}

		go p.handleConnection(clientConn)
	}
}

// Stop stops the proxy.
func (p *Proxy) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.active {
		p.active = false
		if p.listener != nil {
			p.listener.Close()
		}
	}
}

func (p *Proxy) handleConnection(clientConn net.Conn) {
	// Intercept and forward
	targetConn, err := net.Dial("tcp", p.targetAddr)
	if err != nil {
		log.Printf("Proxy dial error: %v", err)
		clientConn.Close()
		return
	}

	errCh := make(chan error, 2)

	go func() {
		_, err := io.Copy(targetConn, clientConn)
		errCh <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, targetConn)
		errCh <- err
	}()

	<-errCh
	clientConn.Close()
	targetConn.Close()
}
