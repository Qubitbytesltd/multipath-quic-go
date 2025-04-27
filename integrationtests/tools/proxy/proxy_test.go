package quicproxy

import (
<<<<<<< HEAD
	"net"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/wire"

	"github.com/stretchr/testify/require"
)

func TestPacketQueue(t *testing.T) {
	q := newQueue()

	getPackets := func() []string {
		packets := make([]string, 0, len(q.Packets))
		for _, p := range q.Packets {
			packets = append(packets, string(p.Raw))
		}
		return packets
	}

	require.Empty(t, getPackets())
	now := time.Now()

	q.Add(packetEntry{Time: now, Raw: []byte("p3")})
	require.Equal(t, []string{"p3"}, getPackets())
	q.Add(packetEntry{Time: now.Add(time.Second), Raw: []byte("p4")})
	require.Equal(t, []string{"p3", "p4"}, getPackets())
	q.Add(packetEntry{Time: now.Add(-time.Second), Raw: []byte("p1")})
	require.Equal(t, []string{"p1", "p3", "p4"}, getPackets())
	q.Add(packetEntry{Time: now.Add(time.Second), Raw: []byte("p5")})
	require.Equal(t, []string{"p1", "p3", "p4", "p5"}, getPackets())
	q.Add(packetEntry{Time: now.Add(-time.Second), Raw: []byte("p2")})
	require.Equal(t, []string{"p1", "p2", "p3", "p4", "p5"}, getPackets())
}

func newUPDConnLocalhost(t testing.TB) *net.UDPConn {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return conn
}

func makePacket(t *testing.T, p protocol.PacketNumber, payload []byte) []byte {
	t.Helper()
	hdr := wire.ExtendedHeader{
		Header: wire.Header{
			Type:             protocol.PacketTypeInitial,
			Version:          protocol.Version1,
			Length:           4 + protocol.ByteCount(len(payload)),
			DestConnectionID: protocol.ParseConnectionID([]byte{0xde, 0xad, 0xbe, 0xef, 0, 0, 0x13, 0x37}),
			SrcConnectionID:  protocol.ParseConnectionID([]byte{0xde, 0xad, 0xbe, 0xef, 0, 0, 0x13, 0x37}),
		},
		PacketNumber:    p,
		PacketNumberLen: protocol.PacketNumberLen4,
	}
	b, err := hdr.Append(nil, protocol.Version1)
	require.NoError(t, err)
	b = append(b, payload...)
	return b
}

func readPacketNumber(t *testing.T, b []byte) protocol.PacketNumber {
	t.Helper()
	hdr, data, _, err := wire.ParsePacket(b)
	require.NoError(t, err)
	require.Equal(t, protocol.PacketTypeInitial, hdr.Type)
	extHdr, err := hdr.ParseExtended(data)
	require.NoError(t, err)
	return extHdr.PacketNumber
}

// Set up a dumb UDP server.
// In production this would be a QUIC server.
func runServer(t *testing.T) (*net.UDPAddr, chan []byte) {
	done := make(chan struct{})
	t.Cleanup(func() {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	})

	serverConn := newUPDConnLocalhost(t)
	serverReceivedPackets := make(chan []byte, 100)
	go func() {
		defer close(done)
		for {
			buf := make([]byte, protocol.MaxPacketBufferSize)
			// the ReadFromUDP will error as soon as the UDP conn is closed
			n, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			serverReceivedPackets <- buf[:n]
			// echo the packet
			if _, err := serverConn.WriteToUDP(buf[:n], addr); err != nil {
				return
			}
		}
	}()

	return serverConn.LocalAddr().(*net.UDPAddr), serverReceivedPackets
}

func TestProxyingBackAndForth(t *testing.T) {
	serverAddr, _ := runServer(t)
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	// send the first packet
	_, err = clientConn.Write(makePacket(t, 1, []byte("foobar")))
	require.NoError(t, err)
	// send the second packet
	_, err = clientConn.Write(makePacket(t, 2, []byte("decafbad")))
	require.NoError(t, err)

	buf := make([]byte, 1024)
	n, err := clientConn.Read(buf)
	require.NoError(t, err)
	require.Contains(t, string(buf[:n]), "foobar")
	n, err = clientConn.Read(buf)
	require.NoError(t, err)
	require.Contains(t, string(buf[:n]), "decafbad")
}

func TestDropIncomingPackets(t *testing.T) {
	const numPackets = 6
	serverAddr, serverReceivedPackets := runServer(t)
	var counter atomic.Int32
	var fromAddr, toAddr atomic.Pointer[net.Addr]
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
		DropPacket: func(d Direction, from, to net.Addr, _ []byte) bool {
			if d != DirectionIncoming {
				return false
			}
			fromAddr.Store(&from)
			toAddr.Store(&to)
			return counter.Add(1)%2 == 1
		},
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	for i := 1; i <= numPackets; i++ {
		_, err := clientConn.Write(makePacket(t, protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
		require.NoError(t, err)
	}

	for i := 0; i < numPackets/2; i++ {
		select {
		case <-serverReceivedPackets:
		case <-time.After(time.Second):
			t.Fatalf("timeout")
		}
	}
	select {
	case <-serverReceivedPackets:
		t.Fatalf("received unexpected packet")
	case <-time.After(100 * time.Millisecond):
	}

	require.Equal(t, *fromAddr.Load(), clientConn.LocalAddr())
	require.Equal(t, *toAddr.Load(), serverAddr)
}

func TestDropOutgoingPackets(t *testing.T) {
	const numPackets = 6
	serverAddr, serverReceivedPackets := runServer(t)
	var counter atomic.Int32
	var fromAddr, toAddr atomic.Pointer[net.Addr]
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
		DropPacket: func(d Direction, from, to net.Addr, _ []byte) bool {
			if d != DirectionOutgoing {
				return false
			}
			fromAddr.Store(&from)
			toAddr.Store(&to)
			return counter.Add(1)%2 == 1
		},
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	clientReceivedPackets := make(chan struct{}, numPackets)
	// receive the packets echoed by the server on client side
	go func() {
		for {
			buf := make([]byte, protocol.MaxPacketBufferSize)
			if _, _, err := clientConn.ReadFromUDP(buf); err != nil {
				return
			}
			clientReceivedPackets <- struct{}{}
		}
	}()

	for i := 1; i <= numPackets; i++ {
		_, err := clientConn.Write(makePacket(t, protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
		require.NoError(t, err)
	}

	for i := 0; i < numPackets/2; i++ {
		select {
		case <-clientReceivedPackets:
		case <-time.After(time.Second):
			t.Fatalf("timeout")
		}
	}
	select {
	case <-clientReceivedPackets:
		t.Fatalf("received unexpected packet")
	case <-time.After(100 * time.Millisecond):
	}
	require.Len(t, serverReceivedPackets, numPackets)

	require.Equal(t, *fromAddr.Load(), serverAddr)
	require.Equal(t, *toAddr.Load(), clientConn.LocalAddr())
}

func TestDelayIncomingPackets(t *testing.T) {
	const numPackets = 3
	const delay = 200 * time.Millisecond
	serverAddr, serverReceivedPackets := runServer(t)
	var counter atomic.Int32
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
		DelayPacket: func(d Direction, _, _ net.Addr, _ []byte) time.Duration {
			// delay packet 1 by 200 ms
			// delay packet 2 by 400 ms
			// ...
			if d == DirectionOutgoing {
				return 0
			}
			p := counter.Add(1)
			return time.Duration(p) * delay
		},
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	start := time.Now()
	for i := 1; i <= numPackets; i++ {
		_, err := clientConn.Write(makePacket(t, protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
		require.NoError(t, err)
	}

	for i := 1; i <= numPackets; i++ {
		select {
		case data := <-serverReceivedPackets:
			require.WithinDuration(t, start.Add(time.Duration(i)*delay), time.Now(), delay/2)
			require.Equal(t, protocol.PacketNumber(i), readPacketNumber(t, data))
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for packet %d", i)
		}
	}
}

func TestPacketReordering(t *testing.T) {
	const delay = 200 * time.Millisecond
	expectDelay := func(startTime time.Time, numRTTs int) {
		expectedReceiveTime := startTime.Add(time.Duration(numRTTs) * delay)
		now := time.Now()
		require.True(t, now.After(expectedReceiveTime) || now.Equal(expectedReceiveTime))
		require.True(t, now.Before(expectedReceiveTime.Add(delay/2)))
	}

	serverAddr, serverReceivedPackets := runServer(t)
	var counter atomic.Int32
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
		DelayPacket: func(d Direction, _, _ net.Addr, _ []byte) time.Duration {
			// delay packet 1 by 600 ms
			// delay packet 2 by 400 ms
			// delay packet 3 by 200 ms
			if d == DirectionOutgoing {
				return 0
			}
			p := counter.Add(1)
			return 600*time.Millisecond - time.Duration(p-1)*delay
		},
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	// send 3 packets
	start := time.Now()
	for i := 1; i <= 3; i++ {
		_, err := clientConn.Write(makePacket(t, protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
		require.NoError(t, err)
	}
	for i := 1; i <= 3; i++ {
		select {
		case packet := <-serverReceivedPackets:
			expectDelay(start, i)
			expectedPacketNumber := protocol.PacketNumber(4 - i) // 3, 2, 1 in reverse order
			require.Equal(t, expectedPacketNumber, readPacketNumber(t, packet))
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for packet %d", i)
		}
	}
}

func TestConstantDelay(t *testing.T) { // no reordering expected here
	serverAddr, serverReceivedPackets := runServer(t)
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
		DelayPacket: func(d Direction, _, _ net.Addr, _ []byte) time.Duration {
			if d == DirectionOutgoing {
				return 0
			}
			return 100 * time.Millisecond
		},
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	// send 100 packets
	for i := 0; i < 100; i++ {
		_, err := clientConn.Write(makePacket(t, protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
		require.NoError(t, err)
	}
	require.Eventually(t, func() bool { return len(serverReceivedPackets) == 100 }, 5*time.Second, 10*time.Millisecond)
	timeout := time.After(5 * time.Second)
	for i := 0; i < 100; i++ {
		select {
		case packet := <-serverReceivedPackets:
			require.Equal(t, protocol.PacketNumber(i), readPacketNumber(t, packet))
		case <-timeout:
			t.Fatalf("timeout waiting for packet %d", i)
		}
	}
}

func TestDelayOutgoingPackets(t *testing.T) {
	const numPackets = 3
	const delay = 200 * time.Millisecond

	serverAddr, serverReceivedPackets := runServer(t)
	var counter atomic.Int32
	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverAddr,
		DelayPacket: func(d Direction, _, _ net.Addr, _ []byte) time.Duration {
			// delay packet 1 by 200 ms
			// delay packet 2 by 400 ms
			// ...
			if d == DirectionIncoming {
				return 0
			}
			p := counter.Add(1)
			return time.Duration(p) * delay
		},
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()
	clientConn, err := net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	clientReceivedPackets := make(chan []byte, numPackets)
	// receive the packets echoed by the server on client side
	go func() {
		for {
			buf := make([]byte, protocol.MaxPacketBufferSize)
			n, _, err := clientConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			clientReceivedPackets <- buf[:n]
		}
	}()

	start := time.Now()
	for i := 1; i <= numPackets; i++ {
		_, err := clientConn.Write(makePacket(t, protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
		require.NoError(t, err)
	}
	// the packets should have arrived immediately at the server
	for i := 0; i < numPackets; i++ {
		select {
		case <-serverReceivedPackets:
		case <-time.After(time.Second):
			t.Fatalf("timeout")
		}
	}
	require.WithinDuration(t, start, time.Now(), delay/2)

	for i := 1; i <= numPackets; i++ {
		select {
		case packet := <-clientReceivedPackets:
			require.Equal(t, protocol.PacketNumber(i), readPacketNumber(t, packet))
			require.WithinDuration(t, start.Add(time.Duration(i)*delay), time.Now(), delay/2)
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for packet %d", i)
		}
	}
}

func TestProxySwitchConn(t *testing.T) {
	serverConn := newUPDConnLocalhost(t)

	type packet struct {
		Data []byte
		Addr *net.UDPAddr
	}
	serverReceivedPackets := make(chan packet, 1)
	go func() {
		for {
			buf := make([]byte, 1000)
			n, addr, err := serverConn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			serverReceivedPackets <- packet{Data: buf[:n], Addr: addr}
		}
	}()

	proxy := Proxy{
		Conn:       newUPDConnLocalhost(t),
		ServerAddr: serverConn.LocalAddr().(*net.UDPAddr),
	}
	require.NoError(t, proxy.Start())
	defer proxy.Close()

	clientConn := newUPDConnLocalhost(t)
	_, err := clientConn.WriteToUDP([]byte("hello"), proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)
	clientConn.SetReadDeadline(time.Now().Add(time.Second))

	var firstConnAddr *net.UDPAddr
	select {
	case p := <-serverReceivedPackets:
		require.Equal(t, "hello", string(p.Data))
		require.NotEqual(t, clientConn.LocalAddr(), p.Addr)
		firstConnAddr = p.Addr
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}

	_, err = serverConn.WriteToUDP([]byte("hi"), firstConnAddr)
	require.NoError(t, err)
	buf := make([]byte, 1000)
	n, addr, err := clientConn.ReadFromUDP(buf)
	require.NoError(t, err)
	require.Equal(t, "hi", string(buf[:n]))
	require.Equal(t, proxy.LocalAddr(), addr)

	newConn := newUPDConnLocalhost(t)
	require.NoError(t, proxy.SwitchConn(clientConn.LocalAddr().(*net.UDPAddr), newConn))

	_, err = clientConn.WriteToUDP([]byte("foobar"), proxy.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)

	select {
	case p := <-serverReceivedPackets:
		require.Equal(t, "foobar", string(p.Data))
		require.NotEqual(t, clientConn.LocalAddr(), p.Addr)
		require.NotEqual(t, firstConnAddr, p.Addr)
		require.Equal(t, newConn.LocalAddr(), p.Addr)
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}

	// the old connection doesn't deliver any packets to the client anymore
	_, err = serverConn.WriteTo([]byte("invalid"), firstConnAddr)
	require.NoError(t, err)
	_, err = serverConn.WriteTo([]byte("foobaz"), newConn.LocalAddr())
	require.NoError(t, err)
	n, addr, err = clientConn.ReadFromUDP(buf)
	require.NoError(t, err)
	require.Equal(t, "foobaz", string(buf[:n])) // "invalid" is not delivered
	require.Equal(t, proxy.LocalAddr(), addr)
}
=======
	"bytes"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/internal/wire"
)

type packetData []byte

var _ = Describe("QUIC Proxy", func() {
	makePacket := func(p protocol.PacketNumber, payload []byte) []byte {
		b := &bytes.Buffer{}
		hdr := wire.PublicHeader{
			PacketNumber:         p,
			PacketNumberLen:      protocol.PacketNumberLen6,
			ConnectionID:         1337,
			TruncateConnectionID: false,
		}
		hdr.Write(b, protocol.VersionWhatever, protocol.PerspectiveServer)
		raw := b.Bytes()
		raw = append(raw, payload...)
		return raw
	}

	Context("Proxy setup and teardown", func() {
		It("sets up the UDPProxy", func() {
			proxy, err := NewQuicProxy("localhost:0", protocol.VersionWhatever, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(proxy.clientDict).To(HaveLen(0))

			// check that the proxy port is in use
			addr, err := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(proxy.LocalPort()))
			Expect(err).ToNot(HaveOccurred())
			_, err = net.ListenUDP("udp", addr)
			Expect(err).To(MatchError(fmt.Sprintf("listen udp 127.0.0.1:%d: bind: address already in use", proxy.LocalPort())))
			Expect(proxy.Close()).To(Succeed()) // stopping is tested in the next test
		})

		It("stops the UDPProxy", func() {
			proxy, err := NewQuicProxy("localhost:0", protocol.VersionWhatever, nil)
			Expect(err).ToNot(HaveOccurred())
			port := proxy.LocalPort()
			err = proxy.Close()
			Expect(err).ToNot(HaveOccurred())

			// check that the proxy port is not in use anymore
			addr, err := net.ResolveUDPAddr("udp", "localhost:"+strconv.Itoa(port))
			Expect(err).ToNot(HaveOccurred())
			// sometimes it takes a while for the OS to free the port
			Eventually(func() error {
				ln, err := net.ListenUDP("udp", addr)
				defer ln.Close()
				return err
			}).ShouldNot(HaveOccurred())
		})

		It("has the correct LocalAddr and LocalPort", func() {
			proxy, err := NewQuicProxy("localhost:0", protocol.VersionWhatever, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(proxy.LocalAddr().String()).To(Equal("127.0.0.1:" + strconv.Itoa(proxy.LocalPort())))
			Expect(proxy.LocalPort()).ToNot(BeZero())

			Expect(proxy.Close()).To(Succeed())
		})
	})

	Context("Proxy tests", func() {
		var (
			serverConn            *net.UDPConn
			serverNumPacketsSent  int32
			serverReceivedPackets chan packetData
			clientConn            *net.UDPConn
			proxy                 *QuicProxy
		)

		startProxy := func(opts *Opts) {
			var err error
			proxy, err = NewQuicProxy("localhost:0", protocol.VersionWhatever, opts)
			Expect(err).ToNot(HaveOccurred())
			clientConn, err = net.DialUDP("udp", nil, proxy.LocalAddr().(*net.UDPAddr))
			Expect(err).ToNot(HaveOccurred())
		}

		// getClientDict returns a copy of the clientDict map
		getClientDict := func() map[string]*connection {
			d := make(map[string]*connection)
			proxy.mutex.Lock()
			defer proxy.mutex.Unlock()
			for k, v := range proxy.clientDict {
				d[k] = v
			}
			return d
		}

		BeforeEach(func() {
			serverReceivedPackets = make(chan packetData, 100)
			atomic.StoreInt32(&serverNumPacketsSent, 0)

			// setup a dump UDP server
			// in production this would be a QUIC server
			raddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
			Expect(err).ToNot(HaveOccurred())
			serverConn, err = net.ListenUDP("udp", raddr)
			Expect(err).ToNot(HaveOccurred())

			go func() {
				for {
					buf := make([]byte, protocol.MaxPacketSize)
					// the ReadFromUDP will error as soon as the UDP conn is closed
					n, addr, err2 := serverConn.ReadFromUDP(buf)
					if err2 != nil {
						return
					}
					data := buf[0:n]
					serverReceivedPackets <- packetData(data)
					// echo the packet
					serverConn.WriteToUDP(data, addr)
					atomic.AddInt32(&serverNumPacketsSent, 1)
				}
			}()
		})

		AfterEach(func() {
			err := proxy.Close()
			Expect(err).ToNot(HaveOccurred())
			err = serverConn.Close()
			Expect(err).ToNot(HaveOccurred())
			err = clientConn.Close()
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(200 * time.Millisecond)
		})

		Context("no packet drop", func() {
			It("relays packets from the client to the server", func() {
				startProxy(&Opts{RemoteAddr: serverConn.LocalAddr().String()})
				// send the first packet
				_, err := clientConn.Write(makePacket(1, []byte("foobar")))
				Expect(err).ToNot(HaveOccurred())

				Eventually(getClientDict).Should(HaveLen(1))
				var conn *connection
				for _, conn = range getClientDict() {
					Expect(atomic.LoadUint64(&conn.incomingPacketCounter)).To(Equal(uint64(1)))
				}

				// send the second packet
				_, err = clientConn.Write(makePacket(2, []byte("decafbad")))
				Expect(err).ToNot(HaveOccurred())

				Eventually(serverReceivedPackets).Should(HaveLen(2))
				Expect(getClientDict()).To(HaveLen(1))
				Expect(string(<-serverReceivedPackets)).To(ContainSubstring("foobar"))
				Expect(string(<-serverReceivedPackets)).To(ContainSubstring("decafbad"))
			})

			It("relays packets from the server to the client", func() {
				startProxy(&Opts{RemoteAddr: serverConn.LocalAddr().String()})
				// send the first packet
				_, err := clientConn.Write(makePacket(1, []byte("foobar")))
				Expect(err).ToNot(HaveOccurred())

				Eventually(getClientDict).Should(HaveLen(1))
				var key string
				var conn *connection
				for key, conn = range getClientDict() {
					Eventually(func() uint64 { return atomic.LoadUint64(&conn.outgoingPacketCounter) }).Should(Equal(uint64(1)))
				}

				// send the second packet
				_, err = clientConn.Write(makePacket(2, []byte("decafbad")))
				Expect(err).ToNot(HaveOccurred())

				Expect(getClientDict()).To(HaveLen(1))
				Eventually(func() uint64 {
					conn := getClientDict()[key]
					return atomic.LoadUint64(&conn.outgoingPacketCounter)
				}).Should(BeEquivalentTo(2))

				clientReceivedPackets := make(chan packetData, 2)
				// receive the packets echoed by the server on client side
				go func() {
					for {
						buf := make([]byte, protocol.MaxPacketSize)
						// the ReadFromUDP will error as soon as the UDP conn is closed
						n, _, err2 := clientConn.ReadFromUDP(buf)
						if err2 != nil {
							return
						}
						data := buf[0:n]
						clientReceivedPackets <- packetData(data)
					}
				}()

				Eventually(serverReceivedPackets).Should(HaveLen(2))
				Expect(atomic.LoadInt32(&serverNumPacketsSent)).To(BeEquivalentTo(2))
				Eventually(clientReceivedPackets).Should(HaveLen(2))
				Expect(string(<-clientReceivedPackets)).To(ContainSubstring("foobar"))
				Expect(string(<-clientReceivedPackets)).To(ContainSubstring("decafbad"))
			})
		})

		Context("Drop Callbacks", func() {
			It("drops incoming packets", func() {
				opts := &Opts{
					RemoteAddr: serverConn.LocalAddr().String(),
					DropPacket: func(d Direction, p uint64) bool {
						return d == DirectionIncoming && p%2 == 0
					},
				}
				startProxy(opts)

				for i := 1; i <= 6; i++ {
					_, err := clientConn.Write(makePacket(protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
					Expect(err).ToNot(HaveOccurred())
				}
				Eventually(serverReceivedPackets).Should(HaveLen(3))
				Consistently(serverReceivedPackets).Should(HaveLen(3))
			})

			It("drops outgoing packets", func() {
				const numPackets = 6
				opts := &Opts{
					RemoteAddr: serverConn.LocalAddr().String(),
					DropPacket: func(d Direction, p uint64) bool {
						return d == DirectionOutgoing && p%2 == 0
					},
				}
				startProxy(opts)

				clientReceivedPackets := make(chan packetData, numPackets)
				// receive the packets echoed by the server on client side
				go func() {
					for {
						buf := make([]byte, protocol.MaxPacketSize)
						// the ReadFromUDP will error as soon as the UDP conn is closed
						n, _, err2 := clientConn.ReadFromUDP(buf)
						if err2 != nil {
							return
						}
						data := buf[0:n]
						clientReceivedPackets <- packetData(data)
					}
				}()

				for i := 1; i <= numPackets; i++ {
					_, err := clientConn.Write(makePacket(protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
					Expect(err).ToNot(HaveOccurred())
				}

				Eventually(clientReceivedPackets).Should(HaveLen(numPackets / 2))
				Consistently(clientReceivedPackets).Should(HaveLen(numPackets / 2))
			})
		})

		Context("Delay Callback", func() {
			expectDelay := func(startTime time.Time, rtt time.Duration, numRTTs int) {
				expectedReceiveTime := startTime.Add(time.Duration(numRTTs) * rtt)
				Expect(time.Now()).To(SatisfyAll(
					BeTemporally(">=", expectedReceiveTime),
					BeTemporally("<", expectedReceiveTime.Add(rtt/2)),
				))
			}

			It("delays incoming packets", func() {
				delay := 300 * time.Millisecond
				opts := &Opts{
					RemoteAddr: serverConn.LocalAddr().String(),
					// delay packet 1 by 200 ms
					// delay packet 2 by 400 ms
					// ...
					DelayPacket: func(d Direction, p uint64) time.Duration {
						if d == DirectionOutgoing {
							return 0
						}
						return time.Duration(p) * delay
					},
				}
				startProxy(opts)

				// send 3 packets
				start := time.Now()
				for i := 1; i <= 3; i++ {
					_, err := clientConn.Write(makePacket(protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
					Expect(err).ToNot(HaveOccurred())
				}
				Eventually(serverReceivedPackets).Should(HaveLen(1))
				expectDelay(start, delay, 1)
				Eventually(serverReceivedPackets).Should(HaveLen(2))
				expectDelay(start, delay, 2)
				Eventually(serverReceivedPackets).Should(HaveLen(3))
				expectDelay(start, delay, 3)
			})

			It("delays outgoing packets", func() {
				const numPackets = 3
				delay := 300 * time.Millisecond
				opts := &Opts{
					RemoteAddr: serverConn.LocalAddr().String(),
					// delay packet 1 by 200 ms
					// delay packet 2 by 400 ms
					// ...
					DelayPacket: func(d Direction, p uint64) time.Duration {
						if d == DirectionIncoming {
							return 0
						}
						return time.Duration(p) * delay
					},
				}
				startProxy(opts)

				clientReceivedPackets := make(chan packetData, numPackets)
				// receive the packets echoed by the server on client side
				go func() {
					for {
						buf := make([]byte, protocol.MaxPacketSize)
						// the ReadFromUDP will error as soon as the UDP conn is closed
						n, _, err2 := clientConn.ReadFromUDP(buf)
						if err2 != nil {
							return
						}
						data := buf[0:n]
						clientReceivedPackets <- packetData(data)
					}
				}()

				start := time.Now()
				for i := 1; i <= numPackets; i++ {
					_, err := clientConn.Write(makePacket(protocol.PacketNumber(i), []byte("foobar"+strconv.Itoa(i))))
					Expect(err).ToNot(HaveOccurred())
				}
				// the packets should have arrived immediately at the server
				Eventually(serverReceivedPackets).Should(HaveLen(3))
				expectDelay(start, delay, 0)
				Eventually(clientReceivedPackets).Should(HaveLen(1))
				expectDelay(start, delay, 1)
				Eventually(clientReceivedPackets).Should(HaveLen(2))
				expectDelay(start, delay, 2)
				Eventually(clientReceivedPackets).Should(HaveLen(3))
				expectDelay(start, delay, 3)
			})
		})
	})
})
>>>>>>> project-faster/main
