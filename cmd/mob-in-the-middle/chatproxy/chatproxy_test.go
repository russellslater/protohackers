package chatproxy_test

import (
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/russellslater/protohackers"
	"github.com/russellslater/protohackers/cmd/mob-in-the-middle/chatproxy"
)

const upstreamEchoSvrPort = 5000

func init() {
	startUpstreamEchoServer(upstreamEchoSvrPort)
}

func startUpstreamEchoServer(port int) {
	go protohackers.ListenAndAccept(port, func(c net.Conn) error {
		defer c.Close()
		_, err := io.Copy(c, c)
		return err
	})

	waitForSvr(port)
}

func waitForSvr(port int) {
	for {
		c, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			c.Close()
			break
		}
		time.Sleep(time.Second)
	}
}

func startProxySvr(port int) *chatproxy.ChatProxy {
	proxySvr := chatproxy.NewChatProxy(port, fmt.Sprintf(":%d", upstreamEchoSvrPort))
	go proxySvr.Start()

	waitForSvr(port)

	return proxySvr
}

func TestChatProxy(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	proxySvrPort := 5001

	proxySvr := startProxySvr(proxySvrPort)
	defer proxySvr.Close()

	// connect to proxy
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", proxySvrPort))

	if err != nil {
		t.Errorf("could not connect to proxy server: %v", err)
	}

	defer conn.Close()

	// expect 'Hello, World!' to be echoed back
	conn.Write([]byte("Hello, World!\n"))

	got := make([]byte, 1000)

	if n, err := conn.Read(got); err == nil {
		is.Equal(string(got[:n]), "Hello, World!\n") // response did not match
	} else {
		t.Errorf("could not read from connection: %v", err)
	}

	// newline required for proxy to complete read
	conn.Write([]byte("Alpha"))
	conn.Write([]byte("Beta"))
	conn.Write([]byte("Gamma"))
	conn.Write([]byte("\n"))

	if n, err := conn.Read(got); err == nil {
		is.Equal(string(got[:n]), "AlphaBetaGamma\n") // response did not match
	} else {
		t.Errorf("could not read from connection: %v", err)
	}
}

type testRewriter struct {
	targetValue  string
	rewriteValue string
}

func (s *testRewriter) Rewrite(src string) string {
	return strings.ReplaceAll(src, s.targetValue, s.rewriteValue)
}

func (s *testRewriter) RewriteBytes(src []byte) []byte { return []byte(s.Rewrite(string(src))) }

func TestChatProxyRewrite(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	port := 5002

	proxySvr := startProxySvr(port)
	defer proxySvr.Close()

	// connect to proxy
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		t.Errorf("could not connect to proxy server: %v", err)
	}

	defer conn.Close()

	proxySvr.Rewriters = []chatproxy.Rewriter{
		&testRewriter{targetValue: "Hello", rewriteValue: "Goodbye"},
		&testRewriter{targetValue: "World", rewriteValue: "Banana"},
	}

	conn.Write([]byte("Hello, World!\n"))

	got := make([]byte, 1000)

	if n, err := conn.Read(got); err == nil {
		is.Equal(string(got[:n]), "Goodbye, Banana!\n") // rewriters did not transform message
	} else {
		t.Errorf("could not read from connection: %v", err)
	}

	// switch out rewriters
	proxySvr.Rewriters = []chatproxy.Rewriter{
		&testRewriter{targetValue: "Three", rewriteValue: "One Hundred"},
	}

	conn.Write([]byte("Hello! One, Two, Three.\n"))

	if n, err := conn.Read(got); err == nil {
		is.Equal(string(got[:n]), "Hello! One, Two, One Hundred.\n") // rewriter did not transform message
	} else {
		t.Errorf("could not read from connection: %v", err)
	}
}
