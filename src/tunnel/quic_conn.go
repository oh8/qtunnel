package tunnel

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/quic-go/quic-go"
)

type QUICConn struct {
	stream quic.Stream
	cipher *Cipher
	pool   *recycler
}

func NewQUICConn(stream quic.Stream, cipher *Cipher, pool *recycler) *QUICConn {
	return &QUICConn{
		stream: stream,
		cipher: cipher,
		pool:   pool,
	}
}

func (c *QUICConn) Read(b []byte) (int, error) {
	c.stream.SetReadDeadline(time.Now().Add(30 * time.Minute))
	if c.cipher == nil {
		return c.stream.Read(b)
	}
	n, err := c.stream.Read(b)
	if n > 0 {
		c.cipher.decrypt(b[0:n], b[0:n])
	}
	return n, err
}

func (c *QUICConn) Write(b []byte) (int, error) {
	if c.cipher == nil {
		return c.stream.Write(b)
	}
	c.cipher.encrypt(b, b)
	return c.stream.Write(b)
}

func (c *QUICConn) Close() error {
	return c.stream.Close()
}

func (c *QUICConn) CloseRead() {
	c.stream.CancelRead(0)
}

func (c *QUICConn) CloseWrite() {
	c.stream.Close()
}

// QUICListener wraps quic.Listener to provide Accept method compatible with net.Listener
type QUICListener struct {
	listener quic.Listener
}

func NewQUICListener(listener quic.Listener) *QUICListener {
	return &QUICListener{listener: listener}
}

func (l *QUICListener) Accept() (quic.Connection, error) {
	return l.listener.Accept(context.Background())
}

func (l *QUICListener) Close() error {
	return l.listener.Close()
}

// generateTLSConfig creates a TLS config for QUIC
func generateTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"qtunnel-quic"},
	}
}