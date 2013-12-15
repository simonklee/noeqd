package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"github.com/simonz05/noeqd/snowflake"
	"github.com/simonz05/util/log"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidAuth    = errors.New("invalid auth")
)

var (
	token = os.Getenv("NOEQ_TOKEN")
)

var (
	help       = flag.Bool("h", false, "show help text")
	wid        = flag.Int64("w", 0, "worker id")
	laddr      = flag.String("l", "0.0.0.0:4444", "the address to listen on")
	lts        = flag.Int64("t", -1, "the last timestamp in milliseconds")
	version    = flag.Bool("version", false, "show version number and exit")
	cpuprofile = flag.String("debug.cpuprofile", "", "write cpu profile to file")
)

var Version = "0.1.0"

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Description:
  Fult-tolerant network service for GUID generation. 

  `)
}

var (
	sf *snowflake.Snowflake
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Fprintln(os.Stdout, Version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	setupServer()
	listenAndServe(*laddr)
}

func setupServer() {
	var err error

	if sf, err = snowflake.New(*wid); err != nil {
		log.Fatalln(err)
	}
}

func sigTrapCloser(l net.Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for _ = range c {
			// Once we close the listener the main loop will exit
			l.Close()
			log.Printf("Closed listener %s", l.Addr())
		}
	}()
}

func listenAndServe(laddr string) error {
	l, err := net.Listen("tcp", laddr)

	if err != nil {
		return err
	}

	sigTrapCloser(l)

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Errorln(err)
		}

		go serveConn(conn)
	}
}

func serveConn(conn net.Conn) {
	err := serve(conn, conn)

	if err != io.EOF {
		log.Errorln(err)
	}

	conn.Close()
}

func serve(r io.Reader, w io.Writer) error {
	if token != "" {
		err := auth(r)

		if err != nil {
			return err
		}
	}

	c := make([]byte, 1)

	for {
		// Wait for 1 byte request
		_, err := io.ReadFull(r, c)

		if err != nil {
			return err
		}

		n := uint(c[0])

		if n == 0 {
			// No authing at this point
			return ErrInvalidRequest
		}

		b := make([]byte, n*8)

		for i := uint(0); i < n; i++ {
			id, err := sf.Next()

			if err != nil {
				return err
			}

			off := i * 8
			b[off+0] = byte(id >> 56)
			b[off+1] = byte(id >> 48)
			b[off+2] = byte(id >> 40)
			b[off+3] = byte(id >> 32)
			b[off+4] = byte(id >> 24)
			b[off+5] = byte(id >> 16)
			b[off+6] = byte(id >> 8)
			b[off+7] = byte(id)
		}

		_, err = w.Write(b)

		if err != nil {
			return err
		}
	}
}

func auth(r io.Reader) error {
	b := make([]byte, 2)

	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}

	if b[0] != 0 {
		return ErrInvalidRequest
	}

	b = make([]byte, b[1])

	if _, err := io.ReadFull(r, b); err != nil {
		return err
	}

	if string(b) != token {
		return ErrInvalidAuth
	}

	return nil
}
