package sidecar

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	// Start a dummy target server
	targetListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start target server: %v", err)
	}
	defer targetListener.Close()

	targetAddr := targetListener.Addr().String()
	go func() {
		for {
			conn, err := targetListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				io.Copy(c, c) // Echo server
			}(conn)
		}
	}()

	// Start proxy
	proxy := NewProxy(targetAddr)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen for proxy: %v", err)
	}
	proxyAddr := proxyListener.Addr().String()
	proxyListener.Close() // Let proxy listen on it

	errCh := make(chan error, 1)
	go func() {
		errCh <- proxy.Start(ctx, proxyAddr)
	}()

	// Wait for proxy to start
	time.Sleep(100 * time.Millisecond)

	// Test connection
	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("Failed to connect to proxy: %v", err)
	}
	defer conn.Close()

	msg := "hello proxy"
	fmt.Fprintf(conn, "%s", msg)
	fmt.Fprintf(conn, "%s", msg)

	buf := make([]byte, len(msg))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		t.Fatalf("Failed to read from proxy: %v", err)
	}

	if string(buf) != msg {
		t.Errorf("Expected %q, got %q", msg, string(buf))
	}

	proxy.Stop()
	cancel()
	<-errCh
}

func TestProxy_ListenError(t *testing.T) {
	// Start a listener on a port to cause an error
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	proxy := NewProxy("127.0.0.1:1234")
	ctx := context.Background()

	err = proxy.Start(ctx, l.Addr().String())
	if err == nil {
		t.Fatal("Expected error when starting proxy on an already used address")
	}
}

func TestProxy_DialError(t *testing.T) {
	// Start proxy pointing to a non-existent server
	proxy := NewProxy("127.0.0.1:9999") // unlikely to be running
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen for proxy: %v", err)
	}
	proxyAddr := proxyListener.Addr().String()
	proxyListener.Close()

	go func() {
		_ = proxy.Start(ctx, proxyAddr)
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("Failed to connect to proxy: %v", err)
	}
	defer conn.Close()

	msg := "hello proxy"
	fmt.Fprintf(conn, "%s", msg)

	// Since target server is not there, connection should eventually be closed
	buf := make([]byte, 100)
	err = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}

	_, err = conn.Read(buf)
	if err == nil {
		t.Errorf("Expected an error reading from closed connection, but got nil")
	}

	proxy.Stop()
}
