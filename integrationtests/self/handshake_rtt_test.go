package self_test

import (
<<<<<<< HEAD
	"context"
	"crypto/tls"
	"io"
	"net"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	quicproxy "github.com/quic-go/quic-go/integrationtests/tools/proxy"

	"github.com/stretchr/testify/require"
)

func handshakeWithRTT(t *testing.T, serverAddr net.Addr, tlsConf *tls.Config, quicConf *quic.Config, rtt time.Duration) quic.Connection {
	t.Helper()

	proxy := quicproxy.Proxy{
		Conn:        newUDPConnLocalhost(t),
		ServerAddr:  serverAddr.(*net.UDPAddr),
		DelayPacket: func(quicproxy.Direction, net.Addr, net.Addr, []byte) time.Duration { return rtt / 2 },
	}
	require.NoError(t, proxy.Start())
	t.Cleanup(func() { proxy.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*rtt)
	defer cancel()
	conn, err := quic.Dial(
		ctx,
		newUDPConnLocalhost(t),
		proxy.LocalAddr(),
		tlsConf,
		quicConf,
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.CloseWithError(0, "") })
	return conn
}

func TestHandshakeRTTWithoutRetry(t *testing.T) {
	ln, err := quic.Listen(newUDPConnLocalhost(t), getTLSConfig(), getQuicConfig(nil))
	require.NoError(t, err)
	defer ln.Close()

	clientConfig := getQuicConfig(&quic.Config{
		GetConfigForClient: func(info *quic.ClientInfo) (*quic.Config, error) {
			require.False(t, info.AddrVerified)
			return nil, nil
		},
	})

	const rtt = 400 * time.Millisecond
	start := time.Now()
	handshakeWithRTT(t, ln.Addr(), getTLSClientConfig(), clientConfig, rtt)
	rtts := time.Since(start).Seconds() / rtt.Seconds()
	require.GreaterOrEqual(t, rtts, float64(1))
	require.Less(t, rtts, float64(2))
}

func TestHandshakeRTTWithRetry(t *testing.T) {
	tr := &quic.Transport{
		Conn:                newUDPConnLocalhost(t),
		VerifySourceAddress: func(net.Addr) bool { return true },
	}
	addTracer(tr)
	defer tr.Close()
	ln, err := tr.Listen(getTLSConfig(), getQuicConfig(nil))
	require.NoError(t, err)
	defer ln.Close()

	clientConfig := getQuicConfig(&quic.Config{
		GetConfigForClient: func(info *quic.ClientInfo) (*quic.Config, error) {
			require.True(t, info.AddrVerified)
			return nil, nil
		},
	})
	const rtt = 400 * time.Millisecond
	start := time.Now()
	handshakeWithRTT(t, ln.Addr(), getTLSClientConfig(), clientConfig, rtt)
	rtts := time.Since(start).Seconds() / rtt.Seconds()
	require.GreaterOrEqual(t, rtts, float64(2))
	require.Less(t, rtts, float64(3))
}

func TestHandshakeRTTWithHelloRetryRequest(t *testing.T) {
	tlsConf := getTLSConfig()
	tlsConf.CurvePreferences = []tls.CurveID{tls.CurveP384}

	ln, err := quic.Listen(newUDPConnLocalhost(t), tlsConf, getQuicConfig(nil))
	require.NoError(t, err)
	defer ln.Close()

	const rtt = 400 * time.Millisecond
	start := time.Now()
	handshakeWithRTT(t, ln.Addr(), getTLSClientConfig(), getQuicConfig(nil), rtt)
	rtts := time.Since(start).Seconds() / rtt.Seconds()
	require.GreaterOrEqual(t, rtts, float64(2))
	require.Less(t, rtts, float64(3))
}

func TestHandshakeRTTReceiveMessage(t *testing.T) {
	sendAndReceive := func(t *testing.T, serverConn, clientConn quic.Connection) {
		t.Helper()
		serverStr, err := serverConn.OpenUniStream()
		require.NoError(t, err)
		_, err = serverStr.Write([]byte("foobar"))
		require.NoError(t, err)
		require.NoError(t, serverStr.Close())

		str, err := clientConn.AcceptUniStream(context.Background())
		require.NoError(t, err)
		data, err := io.ReadAll(str)
		require.NoError(t, err)
		require.Equal(t, []byte("foobar"), data)
	}

	t.Run("using Listen", func(t *testing.T) {
		ln, err := quic.Listen(newUDPConnLocalhost(t), getTLSConfig(), getQuicConfig(nil))
		require.NoError(t, err)
		defer ln.Close()

		connChan := make(chan quic.Connection, 1)
		go func() {
			conn, err := ln.Accept(context.Background())
			if err != nil {
				t.Logf("failed to accept connection: %s", err)
				close(connChan)
				return
			}
			connChan <- conn
		}()

		const rtt = 400 * time.Millisecond
		start := time.Now()
		conn := handshakeWithRTT(t, ln.Addr(), getTLSClientConfig(), getQuicConfig(nil), rtt)
		serverConn := <-connChan
		if serverConn == nil {
			t.Fatal("serverConn is nil")
		}
		sendAndReceive(t, serverConn, conn)

		rtts := time.Since(start).Seconds() / rtt.Seconds()
		require.GreaterOrEqual(t, rtts, float64(2))
		require.Less(t, rtts, float64(3))
	})

	t.Run("using ListenEarly", func(t *testing.T) {
		ln, err := quic.ListenEarly(newUDPConnLocalhost(t), getTLSConfig(), getQuicConfig(nil))
		require.NoError(t, err)
		defer ln.Close()

		connChan := make(chan quic.Connection, 1)
		go func() {
			conn, err := ln.Accept(context.Background())
			if err != nil {
				t.Logf("failed to accept connection: %s", err)
				close(connChan)
				return
			}
			connChan <- conn
		}()

		const rtt = 400 * time.Millisecond
		start := time.Now()
		conn := handshakeWithRTT(t, ln.Addr(), getTLSClientConfig(), getQuicConfig(nil), rtt)
		serverConn := <-connChan
		if serverConn == nil {
			t.Fatal("serverConn is nil")
		}
		sendAndReceive(t, serverConn, conn)

		took := time.Since(start)
		rtts := float64(took) / float64(rtt)
		require.GreaterOrEqual(t, rtts, float64(1))
		require.Less(t, rtts, float64(2))
	})
}
=======
	"crypto/tls"
	"fmt"
	"net"
	"time"

	quic "github.com/project-faster/mp-quic-go"
	"github.com/project-faster/mp-quic-go/integrationtests/tools/proxy"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/internal/utils"
	"github.com/project-faster/mp-quic-go/qerr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/project-faster/mp-quic-go/internal/testdata"
)

var _ = Describe("Handshake RTT tests", func() {
	var (
		proxy         *quicproxy.QuicProxy
		server        quic.Listener
		serverConfig  *quic.Config
		testStartedAt time.Time
		acceptStopped chan struct{}
	)

	rtt := 400 * time.Millisecond

	BeforeEach(func() {
		acceptStopped = make(chan struct{})
		serverConfig = &quic.Config{}
	})

	AfterEach(func() {
		Expect(proxy.Close()).To(Succeed())
		Expect(server.Close()).To(Succeed())
		<-acceptStopped
	})

	runServerAndProxy := func() {
		var err error
		// start the server
		server, err = quic.ListenAddr("localhost:0", testdata.GetTLSConfig(), serverConfig)
		Expect(err).ToNot(HaveOccurred())
		// start the proxy
		proxy, err = quicproxy.NewQuicProxy("localhost:0", protocol.VersionWhatever, &quicproxy.Opts{
			RemoteAddr:  server.Addr().String(),
			DelayPacket: func(_ quicproxy.Direction, _ uint64) time.Duration { return rtt / 2 },
		})
		Expect(err).ToNot(HaveOccurred())

		testStartedAt = time.Now()

		go func() {
			defer GinkgoRecover()
			defer close(acceptStopped)
			for {
				_, err := server.Accept()
				if err != nil {
					return
				}
			}
		}()
	}

	expectDurationInRTTs := func(num int) {
		testDuration := time.Since(testStartedAt)
		expectedDuration := time.Duration(num) * rtt
		Expect(testDuration).To(SatisfyAll(
			BeNumerically(">=", expectedDuration),
			BeNumerically("<", expectedDuration+rtt),
		))
	}

	It("fails when there's no matching version, after 1 RTT", func() {
		Expect(len(protocol.SupportedVersions)).To(BeNumerically(">", 1))
		serverConfig.Versions = protocol.SupportedVersions[:1]
		runServerAndProxy()
		clientConfig := &quic.Config{
			Versions: protocol.SupportedVersions[1:2],
		}
		_, err := quic.DialAddr(proxy.LocalAddr().String(), nil, clientConfig)
		Expect(err).To(HaveOccurred())
		Expect(err.(qerr.ErrorCode)).To(Equal(qerr.InvalidVersion))
		expectDurationInRTTs(1)
	})

	// 1 RTT for verifying the source address
	// 1 RTT to become secure
	// 1 RTT to become forward-secure
	It("is forward-secure after 3 RTTs", func() {
		runServerAndProxy()
		_, err := quic.DialAddr(proxy.LocalAddr().String(), &tls.Config{InsecureSkipVerify: true}, nil)
		Expect(err).ToNot(HaveOccurred())
		expectDurationInRTTs(3)
	})

	It("does version negotiation in 1 RTT", func() {
		Expect(len(protocol.SupportedVersions)).To(BeNumerically(">", 1))
		// the server doesn't support the highest supported version, which is the first one the client will try
		serverConfig.Versions = protocol.SupportedVersions[1:]
		runServerAndProxy()
		_, err := quic.DialAddr(proxy.LocalAddr().String(), &tls.Config{InsecureSkipVerify: true}, nil)
		Expect(err).ToNot(HaveOccurred())
		expectDurationInRTTs(4)
	})

	// 1 RTT for verifying the source address
	// 1 RTT to become secure
	// TODO (marten-seemann): enable this test (see #625)
	PIt("is secure after 2 RTTs", func() {
		utils.SetLogLevel(utils.LogLevelDebug)
		runServerAndProxy()
		_, err := quic.DialAddrNonFWSecure(proxy.LocalAddr().String(), &tls.Config{InsecureSkipVerify: true}, nil)
		fmt.Println("#### is non fw secure ###")
		Expect(err).ToNot(HaveOccurred())
		expectDurationInRTTs(2)
	})

	It("is forward-secure after 2 RTTs when the server doesn't require a Cookie", func() {
		serverConfig.AcceptCookie = func(_ net.Addr, _ *quic.Cookie) bool {
			return true
		}
		runServerAndProxy()
		_, err := quic.DialAddr(proxy.LocalAddr().String(), &tls.Config{InsecureSkipVerify: true}, nil)
		Expect(err).ToNot(HaveOccurred())
		expectDurationInRTTs(2)
	})

	It("doesn't complete the handshake when the server never accepts the Cookie", func() {
		serverConfig.AcceptCookie = func(_ net.Addr, _ *quic.Cookie) bool {
			return false
		}
		runServerAndProxy()
		_, err := quic.DialAddr(proxy.LocalAddr().String(), &tls.Config{InsecureSkipVerify: true}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.CryptoTooManyRejects))
	})

	It("doesn't complete the handshake when the handshake timeout is too short", func() {
		serverConfig.HandshakeTimeout = 2 * rtt
		runServerAndProxy()
		_, err := quic.DialAddr(proxy.LocalAddr().String(), &tls.Config{InsecureSkipVerify: true}, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.HandshakeTimeout))
		// 2 RTTs during the timeout
		// plus 1 RTT: the timer starts 0.5 RTTs after sending the first packet, and the CONNECTION_CLOSE needs another 0.5 RTTs to reach the client
		expectDurationInRTTs(3)
	})
})
>>>>>>> project-faster/main
