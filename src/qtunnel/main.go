package main

import (
    "os"
    "os/signal"
    "syscall"
    "log"
    "flag"
    "qtunnel/tunnel"
)

func waitSignal() {
    var sigChan = make(chan os.Signal, 1)
    signal.Notify(sigChan)
    for sig := range sigChan {
        if sig == syscall.SIGINT || sig == syscall.SIGTERM {
            log.Printf("terminated by signal %v\n", sig)
            return
        } else {
            log.Printf("received signal: %v, ignore\n", sig)
        }
    }
}

func main() {
    var faddr, baddr, cryptoMethod, secret, logTo string
    var clientMode, useQUIC bool
    flag.StringVar(&logTo, "logto", "stdout", "stdout or syslog")
    flag.StringVar(&faddr, "listen", ":9001", "host:port qtunnel listen on")
    flag.StringVar(&baddr, "backend", "127.0.0.1:6400", "host:port of the backend")
    flag.StringVar(&cryptoMethod, "crypto", "rc4", "encryption method")
    flag.StringVar(&secret, "secret", "secret", "password used to encrypt the data")
    flag.BoolVar(&clientMode, "clientmode", false, "if running at client mode")
    flag.BoolVar(&useQUIC, "quic", false, "use QUIC protocol instead of TCP")
    flag.Parse()

    log.SetOutput(os.Stdout)
    if logTo == "syslog" {
        err := setupSyslog()
        if err != nil {
            log.Fatal(err)
        }
    }

    var t *tunnel.Tunnel
    if useQUIC {
        t = tunnel.NewTunnelWithQUIC(faddr, baddr, clientMode, cryptoMethod, secret, 4096, true)
        log.Println("qtunnel started with QUIC protocol.")
    } else {
        t = tunnel.NewTunnel(faddr, baddr, clientMode, cryptoMethod, secret, 4096)
        log.Println("qtunnel started with TCP protocol.")
    }
    go t.Start()
    waitSignal()
}
