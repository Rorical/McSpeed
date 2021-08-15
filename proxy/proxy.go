package proxy

import (
	"fmt"
	"io"
	"net"

	"github.com/mgutz/ansi"
)

// Logger - Interface to pass into Proxy for it to log messages
type Logger interface {
	Trace(f string, args ...interface{})
	Debug(f string, args ...interface{})
	Info(f string, args ...interface{})
	Warn(f string, args ...interface{})
}

// NullLogger - An empty logger that ignores everything
type NullLogger struct{}

// Trace - no-op
func (l NullLogger) Trace(f string, args ...interface{}) {}

// Debug - no-op
func (l NullLogger) Debug(f string, args ...interface{}) {}

// Info - no-op
func (l NullLogger) Info(f string, args ...interface{}) {}

// Warn - no-op
func (l NullLogger) Warn(f string, args ...interface{}) {}

// ColorLogger - A Logger that logs to stdout in color
type ColorLogger struct {
	VeryVerbose bool
	Verbose     bool
	Prefix      string
	Color       bool
}

// Trace - Log a very verbose trace message
func (l ColorLogger) Trace(f string, args ...interface{}) {
	if !l.VeryVerbose {
		return
	}
	l.output("blue", f, args...)
}

// Debug - Log a debug message
func (l ColorLogger) Debug(f string, args ...interface{}) {
	if !l.Verbose {
		return
	}
	l.output("green", f, args...)
}

// Info - Log a general message
func (l ColorLogger) Info(f string, args ...interface{}) {
	l.output("green", f, args...)
}

// Warn - Log a warning
func (l ColorLogger) Warn(f string, args ...interface{}) {
	l.output("red", f, args...)
}

func (l ColorLogger) output(color, f string, args ...interface{}) {
	if l.Color && color != "" {
		f = ansi.Color(f, color)
	}
	fmt.Printf(fmt.Sprintf("%s%s\n", l.Prefix, f), args...)
}

// https://github.com/jpillora/go-tcp-proxy/blob/master/proxy.go
// Proxy - Manages a Proxy connection, piping data between local and remote.
type Proxy struct {
	sentBytes     uint64
	receivedBytes uint64
	lconn, rconn  io.ReadWriteCloser
	erred         bool
	errsig        chan bool
	tlsUnwrapp    bool
	tlsAddress    string

	Matcher  func([]byte)
	Replacer func([]byte) []byte

	// Settings
	Nagles    bool
	Log       Logger
	OutputHex bool
}

// New - Create a new Proxy instance. Takes over local connection passed in,
// and closes it when finished.
func New(lconn, rconn net.Conn) *Proxy {
	return &Proxy{
		lconn:  lconn,
		rconn:  rconn,
		erred:  false,
		errsig: make(chan bool),
		Log:    NullLogger{},
	}
}

type setNoDelayer interface {
	SetNoDelay(bool) error
}

// Start - open connection to remote and start proxying data.
func (p *Proxy) Start() {
	defer p.lconn.Close()

	var err error
	if err != nil {
		p.Log.Warn("Remote connection failed: %s", err)
		return
	}
	defer p.rconn.Close()

	//nagles?
	if p.Nagles {
		if conn, ok := p.lconn.(setNoDelayer); ok {
			conn.SetNoDelay(true)
		}
		if conn, ok := p.rconn.(setNoDelayer); ok {
			conn.SetNoDelay(true)
		}
	}

	//display both ends
	//bidirectional copy
	go p.pipe(p.lconn, p.rconn)
	go p.pipe(p.rconn, p.lconn)

	//wait for close...
	<-p.errsig
	p.Log.Info("Closed (%d bytes sent, %d bytes recieved)", p.sentBytes, p.receivedBytes)
}

func (p *Proxy) err(s string, err error) {
	if p.erred {
		return
	}
	if err != io.EOF {
		p.Log.Warn(s, err)
	}
	p.errsig <- true
	p.erred = true
}

func (p *Proxy) pipe(src, dst io.ReadWriter) {
	islocal := src == p.lconn

	var dataDirection string
	if islocal {
		dataDirection = ">>> %d bytes sent%s"
	} else {
		dataDirection = "<<< %d bytes recieved%s"
	}

	var byteFormat string
	if p.OutputHex {
		byteFormat = "%x"
	} else {
		byteFormat = "%s"
	}

	//directional copy (64k buffer)
	buff := make([]byte, 0xffff)
	for {
		n, err := src.Read(buff)
		if err != nil {
			p.err("Read failed '%s'\n", err)
			return
		}
		b := buff[:n]

		//execute match
		if p.Matcher != nil {
			p.Matcher(b)
		}

		//execute replace
		if p.Replacer != nil {
			b = p.Replacer(b)
		}

		//show output
		p.Log.Debug(dataDirection, n, "")
		p.Log.Trace(byteFormat, b)

		//write out result
		n, err = dst.Write(b)
		if err != nil {
			p.err("Write failed '%s'\n", err)
			return
		}
		if islocal {
			p.sentBytes += uint64(n)
		} else {
			p.receivedBytes += uint64(n)
		}
	}
}
