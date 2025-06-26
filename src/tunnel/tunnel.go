package tunnel

import (
    "context"
    "crypto/rand"
    "crypto/rsa"
    "crypto/tls"
    "crypto/x509"
    "crypto/x509/pkix"
    "encoding/pem"
    "io"
    "math/big"
    "net"
    "log"
    "time"
    "sync/atomic"

    "github.com/quic-go/quic-go"
)

type Tunnel struct {
    faddr, baddr *net.TCPAddr
    clientMode bool
    cryptoMethod string
    secret []byte
    sessionsCount int32
    pool *recycler
    useQUIC bool
    tlsConfig *tls.Config
}

func NewTunnel(faddr, baddr string, clientMode bool, cryptoMethod, secret string, size uint32) *Tunnel {
    return NewTunnelWithQUIC(faddr, baddr, clientMode, cryptoMethod, secret, size, false)
}

func NewTunnelWithQUIC(faddr, baddr string, clientMode bool, cryptoMethod, secret string, size uint32, useQUIC bool) *Tunnel {
    a1, err := net.ResolveTCPAddr("tcp", faddr)
    if err != nil {
        log.Fatalln("resolve frontend error:", err)
    }
    a2, err := net.ResolveTCPAddr("tcp", baddr)
    if err != nil {
        log.Fatalln("resolve backend error:", err)
    }
    
    var tlsConfig *tls.Config
    if useQUIC {
        cert, err := generateSelfSignedCert()
        if err != nil {
            log.Fatal("Failed to generate certificate:", err)
        }
        tlsConfig = &tls.Config{
            Certificates:       []tls.Certificate{cert},
            InsecureSkipVerify: true,
            NextProtos:         []string{"qtunnel-quic"},
            MinVersion:         tls.VersionTLS12,
        }
    }
    
    return &Tunnel{
        faddr: a1,
        baddr: a2,
        clientMode: clientMode,
        cryptoMethod: cryptoMethod,
        secret: []byte(secret),
        sessionsCount: 0,
        pool: NewRecycler(size),
        useQUIC: useQUIC,
        tlsConfig: tlsConfig,
    }
}

func (t *Tunnel) pipe(dst, src *Conn, c chan int64) {
    defer func() {
        dst.CloseWrite()
        src.CloseRead()
    }()
    n, err := io.Copy(dst, src)
    if err != nil {
        log.Print(err)
    }
    c <- n
}

func (t *Tunnel) pipeQUIC(dst, src interface{}, c chan int64) {
    defer func() {
        if d, ok := dst.(*Conn); ok {
            d.CloseWrite()
        } else if d, ok := dst.(*QUICConn); ok {
            d.CloseWrite()
        }
        if s, ok := src.(*Conn); ok {
            s.CloseRead()
        } else if s, ok := src.(*QUICConn); ok {
            s.CloseRead()
        }
    }()
    
    var n int64
    var err error
    
    if dstConn, ok := dst.(*Conn); ok {
        if srcConn, ok := src.(*Conn); ok {
            n, err = io.Copy(dstConn, srcConn)
        } else if srcQUIC, ok := src.(*QUICConn); ok {
            n, err = io.Copy(dstConn, srcQUIC)
        }
    } else if dstQUIC, ok := dst.(*QUICConn); ok {
        if srcConn, ok := src.(*Conn); ok {
            n, err = io.Copy(dstQUIC, srcConn)
        } else if srcQUIC, ok := src.(*QUICConn); ok {
            n, err = io.Copy(dstQUIC, srcQUIC)
        }
    }
    
    if err != nil {
        log.Print(err)
    }
    c <- n
}

func (t *Tunnel) transport(conn net.Conn) {
    if t.useQUIC {
        t.transportQUIC(conn)
    } else {
        t.transportTCP(conn)
    }
}

func (t *Tunnel) transportTCP(conn net.Conn) {
    start := time.Now()
    conn2, err := net.DialTCP("tcp", nil, t.baddr)
    if err != nil {
        log.Print(err)
        return
    }
    connectTime := time.Now().Sub(start)
    start = time.Now()
    cipher := NewCipher(t.cryptoMethod, t.secret)
    readChan := make(chan int64)
    writeChan := make(chan int64)
    var readBytes, writeBytes int64
    atomic.AddInt32(&t.sessionsCount, 1)
    var bconn, fconn *Conn
    if t.clientMode {
        fconn = NewConn(conn, nil, t.pool)
        bconn = NewConn(conn2, cipher, t.pool)
    } else {
        fconn = NewConn(conn, cipher, t.pool)
        bconn = NewConn(conn2, nil, t.pool)
    }
    go t.pipe(bconn, fconn, writeChan)
    go t.pipe(fconn, bconn, readChan)
    readBytes = <-readChan
    writeBytes = <-writeChan
    transferTime := time.Now().Sub(start)
    log.Printf("r:%d w:%d ct:%.3f t:%.3f [#%d]", readBytes, writeBytes,
        connectTime.Seconds(), transferTime.Seconds(), t.sessionsCount)
    atomic.AddInt32(&t.sessionsCount, -1)
}

func (t *Tunnel) transportQUIC(conn net.Conn) {
    start := time.Now()
    quicConn, err := quic.DialAddr(context.Background(), t.baddr.String(), t.tlsConfig, nil)
    if err != nil {
        log.Print(err)
        return
    }
    defer quicConn.CloseWithError(0, "")
    
    stream, err := quicConn.OpenStreamSync(context.Background())
    if err != nil {
        log.Print(err)
        return
    }
    
    connectTime := time.Now().Sub(start)
    start = time.Now()
    cipher := NewCipher(t.cryptoMethod, t.secret)
    readChan := make(chan int64)
    writeChan := make(chan int64)
    var readBytes, writeBytes int64
    atomic.AddInt32(&t.sessionsCount, 1)
    
    var bconn, fconn interface{}
    if t.clientMode {
        fconn = NewConn(conn, nil, t.pool)
        bconn = NewQUICConn(stream, cipher, t.pool)
    } else {
        fconn = NewConn(conn, cipher, t.pool)
        bconn = NewQUICConn(stream, nil, t.pool)
    }
    
    go t.pipeQUIC(bconn, fconn, writeChan)
    go t.pipeQUIC(fconn, bconn, readChan)
    readBytes = <-readChan
    writeBytes = <-writeChan
    transferTime := time.Now().Sub(start)
    log.Printf("QUIC r:%d w:%d ct:%.3f t:%.3f [#%d]", readBytes, writeBytes,
        connectTime.Seconds(), transferTime.Seconds(), t.sessionsCount)
    atomic.AddInt32(&t.sessionsCount, -1)
}

func (t *Tunnel) Start() {
    if t.useQUIC && !t.clientMode {
        // Server mode with QUIC: listen on QUIC for client connections
        t.startQUIC()
    } else {
        // Client mode or TCP mode: use TCP frontend
        t.startTCP()
    }
}

func (t *Tunnel) startTCP() {
    ln, err := net.ListenTCP("tcp", t.faddr)
    if err != nil {
        log.Fatal(err)
    }
    defer ln.Close()

    for {
        conn, err := ln.AcceptTCP()
        if err != nil {
            log.Println("accept:", err)
            continue
        }
        go t.transport(conn)
    }
}

func (t *Tunnel) startQUIC() {
    listener, err := quic.ListenAddr(t.faddr.String(), t.tlsConfig, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept(context.Background())
        if err != nil {
            log.Println("QUIC accept:", err)
            continue
        }
        go t.handleQUICConnection(conn)
    }
}

func (t *Tunnel) handleQUICConnection(conn quic.Connection) {
    for {
        stream, err := conn.AcceptStream(context.Background())
        if err != nil {
            log.Println("QUIC stream accept:", err)
            return
        }
        go t.transportQUICServer(stream)
    }
}

func (t *Tunnel) transportQUICServer(stream quic.Stream) {
    start := time.Now()
    conn2, err := net.DialTCP("tcp", nil, t.baddr)
    if err != nil {
        log.Print(err)
        return
    }
    connectTime := time.Now().Sub(start)
    start = time.Now()
    cipher := NewCipher(t.cryptoMethod, t.secret)
    readChan := make(chan int64)
    writeChan := make(chan int64)
    var readBytes, writeBytes int64
    atomic.AddInt32(&t.sessionsCount, 1)
    
    var bconn, fconn interface{}
    if t.clientMode {
        fconn = NewQUICConn(stream, nil, t.pool)
        bconn = NewConn(conn2, cipher, t.pool)
    } else {
        fconn = NewQUICConn(stream, cipher, t.pool)
        bconn = NewConn(conn2, nil, t.pool)
    }
    
    go t.pipeQUIC(bconn, fconn, writeChan)
    go t.pipeQUIC(fconn, bconn, readChan)
    readBytes = <-readChan
    writeBytes = <-writeChan
    transferTime := time.Now().Sub(start)
    log.Printf("QUIC-Server r:%d w:%d ct:%.3f t:%.3f [#%d]", readBytes, writeBytes,
        connectTime.Seconds(), transferTime.Seconds(), t.sessionsCount)
    atomic.AddInt32(&t.sessionsCount, -1)
}

func generateSelfSignedCert() (tls.Certificate, error) {
    priv, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        return tls.Certificate{}, err
    }

    template := x509.Certificate{
        SerialNumber: big.NewInt(1),
        Subject: pkix.Name{
            Organization:  []string{"qtunnel"},
            Country:       []string{"US"},
            Province:      []string{""},
            Locality:      []string{""},
            StreetAddress: []string{""},
            PostalCode:    []string{""},
        },
        NotBefore:    time.Now(),
        NotAfter:     time.Now().Add(365 * 24 * time.Hour),
        KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
        ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
        IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
        DNSNames:     []string{"localhost"},
    }

    certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
    if err != nil {
        return tls.Certificate{}, err
    }

    certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
    keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

    return tls.X509KeyPair(certPEM, keyPEM)
}
