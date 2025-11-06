package main

import (
	"17/internal/parser"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	opt, err := parser.ParseFlags()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error while parsing flags:", err)
		os.Exit(1)
	}
	if opt.Host == "" || opt.Port == "" {
		fmt.Fprintln(os.Stderr, "host and port are required (use --host and --port)")
		os.Exit(1)
	}

	addr := net.JoinHostPort(opt.Host, opt.Port)

	conn, err := net.DialTimeout("tcp", addr, opt.Timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while connecting to %s (timeout %s): %v\n", addr, opt.Timeout, err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Fprintln(os.Stderr, "Connected:", addr)

	done := make(chan struct{}, 2)

	go Writer(conn, done)
	go Reader(conn, done)

	<-done
	<-done
}

func Writer(conn net.Conn, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	_, err := io.Copy(conn, os.Stdin)

	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	} else {
		_ = conn.Close()
	}

	if err != nil && !errors.Is(err, io.EOF) &&
		!errors.Is(err, net.ErrClosed) &&
		!isUseOfClosedConnErr(err) {
		fmt.Fprintln(os.Stderr, "write error:", err)
	}
}

func Reader(conn net.Conn, done chan<- struct{}) {
	defer func() { done <- struct{}{} }()

	_, err := io.Copy(os.Stdout, conn)

	if err != nil && !errors.Is(err, io.EOF) &&
		!errors.Is(err, net.ErrClosed) &&
		!isUseOfClosedConnErr(err) {
		fmt.Fprintln(os.Stderr, "read error:", err)
	}
}

func isUseOfClosedConnErr(err error) bool {
	return strings.Contains(err.Error(), "use of closed network connection")
}
