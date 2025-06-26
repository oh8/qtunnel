package tunnel

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestTunnelQUICRC4(t *testing.T) {
	done := make(chan bool)
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	go backendServer(9482, data, done, t)
	
	// Create QUIC backend tunnel (server mode) - uses TCP frontend, QUIC backend
	b := NewTunnelWithQUIC("127.0.0.1:9450", "127.0.0.1:9482", false, "rc4", "secret", 4, true)
	// Create QUIC frontend tunnel (client mode) - uses TCP frontend, QUIC backend
	f := NewTunnelWithQUIC("127.0.0.1:9451", "127.0.0.1:9450", true, "rc4", "secret", 4, true)
	
	go b.Start()
	go f.Start()
	
	// Wait for servers to start
	time.Sleep(500 * time.Millisecond)
	
	conn, err := net.Dial("tcp", "127.0.0.1:9451")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	
	_, err = conn.Write(data)
	if err != nil {
		t.Error(err)
		return
	}
	// Wait for transmission complete
	time.Sleep(500 * time.Millisecond)
	close(done)
}

func TestTunnelQUICAES256CFB(t *testing.T) {
	done := make(chan bool)
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	go backendServer(9483, data, done, t)
	
	// Create QUIC backend tunnel (server mode) - uses TCP frontend, QUIC backend
	b := NewTunnelWithQUIC("127.0.0.1:9452", "127.0.0.1:9483", false, "aes256cfb", "secret", 4, true)
	// Create QUIC frontend tunnel (client mode) - uses TCP frontend, QUIC backend
	f := NewTunnelWithQUIC("127.0.0.1:9453", "127.0.0.1:9452", true, "aes256cfb", "secret", 4, true)
	
	go b.Start()
	go f.Start()
	
	// Wait for servers to start
	time.Sleep(500 * time.Millisecond)
	
	conn, err := net.Dial("tcp", "127.0.0.1:9453")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	
	_, err = conn.Write(data)
	if err != nil {
		t.Error(err)
		return
	}
	// Wait for transmission complete
	time.Sleep(500 * time.Millisecond)
	close(done)
}

func TestMixedTCPQUIC(t *testing.T) {
	done := make(chan bool)
	data := []byte{9, 8, 7, 6, 5, 4, 3, 2, 1}
	go backendServer(9484, data, done, t)
	
	// Create QUIC backend tunnel (server mode) - uses TCP frontend, QUIC backend
	b := NewTunnelWithQUIC("127.0.0.1:9454", "127.0.0.1:9484", false, "rc4", "secret", 4, true)
	// Create QUIC frontend tunnel (client mode) - uses TCP frontend, QUIC backend
	f := NewTunnelWithQUIC("127.0.0.1:9455", "127.0.0.1:9454", true, "rc4", "secret", 4, true)
	
	go b.Start()
	go f.Start()
	
	// Wait for servers to start
	time.Sleep(500 * time.Millisecond)
	
	conn, err := net.Dial("tcp", "127.0.0.1:9455")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()
	
	_, err = conn.Write(data)
	if err != nil {
		t.Error(err)
		return
	}
	// Wait for transmission complete
	time.Sleep(500 * time.Millisecond)
	close(done)
}

func backendServerQUIC(port int, data []byte, done chan bool, t *testing.T) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Error(err)
	}
	defer ln.Close()
	conn, err := ln.Accept()
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	buf := make([]byte, len(data))
	conn.Read(buf)
	if bytes.Compare(buf, data) != 0 {
		t.Errorf("Expected %v, got %v", data, buf)
	}
	<-done
}