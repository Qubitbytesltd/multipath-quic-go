package quic

import (
<<<<<<< HEAD
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/quic-go/quic-go/internal/handshake"
	mocklogging "github.com/quic-go/quic-go/internal/mocks/logging"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/qerr"
	"github.com/quic-go/quic-go/internal/utils"
	"github.com/quic-go/quic-go/internal/wire"
	"github.com/quic-go/quic-go/logging"
=======
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"net"
	"reflect"
	"time"

	"github.com/project-faster/mp-quic-go/internal/crypto"
	"github.com/project-faster/mp-quic-go/internal/handshake"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/internal/utils"
	"github.com/project-faster/mp-quic-go/internal/wire"
	"github.com/project-faster/mp-quic-go/qerr"
>>>>>>> project-faster/main

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

<<<<<<< HEAD
type testServer struct{ *baseServer }

type serverOpts struct {
	tracer                    *logging.Tracer
	config                    *Config
	tokenGeneratorKey         TokenGeneratorKey
	maxTokenAge               time.Duration
	useRetry                  bool
	disableVersionNegotiation bool
	acceptEarly               bool
	newConn                   func(
		context.Context,
		context.CancelCauseFunc,
		sendConn,
		*Transport,
		protocol.ConnectionID, // original dest connection ID
		*protocol.ConnectionID, // retry src connection ID
		protocol.ConnectionID, // client dest connection ID
		protocol.ConnectionID, // destination connection ID
		protocol.ConnectionID, // source connection ID
		ConnectionIDGenerator,
		*statelessResetter,
		*Config,
		*tls.Config,
		*handshake.TokenGenerator,
		bool, /* client address validated by an address validation token */
		*logging.ConnectionTracer,
		utils.Logger,
		protocol.Version,
	) quicConn
=======
type mockSession struct {
	connectionID      protocol.ConnectionID
	packetCount       int
	closed            bool
	closeReason       error
	closedRemote      bool
	stopRunLoop       chan struct{} // run returns as soon as this channel receives a value
	handshakeChan     chan handshakeEvent
	handshakeComplete chan error // for WaitUntilHandshakeComplete
	remoteAddr        net.Addr
>>>>>>> project-faster/main
}

func newTestServer(t *testing.T, serverOpts *serverOpts) *testServer {
	t.Helper()
	c, err := wrapConn(newUDPConnLocalhost(t))
	require.NoError(t, err)
	verifySourceAddress := func(net.Addr) bool { return serverOpts.useRetry }
	config := populateConfig(serverOpts.config)
	s := newServer(
		c,
		&Transport{handlerMap: newPacketHandlerMap(nil, utils.DefaultLogger)},
		&protocol.DefaultConnectionIDGenerator{},
		&statelessResetter{},
		func(ctx context.Context) context.Context { return ctx },
		&tls.Config{},
		config,
		serverOpts.tracer,
		func() {},
		serverOpts.tokenGeneratorKey,
		serverOpts.maxTokenAge,
		verifySourceAddress,
		serverOpts.disableVersionNegotiation,
		serverOpts.acceptEarly,
	)
	s.newConn = serverOpts.newConn
	t.Cleanup(func() { s.Close() })
	return &testServer{s}
}

<<<<<<< HEAD
func getLongHeaderPacketEncrypted(t *testing.T, remoteAddr net.Addr, extHdr *wire.ExtendedHeader, data []byte) receivedPacket {
	t.Helper()
	hdr := extHdr.Header
	if hdr.Type != protocol.PacketTypeInitial {
		t.Fatal("can only encrypt Initial packets")
	}
	p := getLongHeaderPacket(t, remoteAddr, extHdr, data)
	sealer, _ := handshake.NewInitialAEAD(hdr.DestConnectionID, protocol.PerspectiveClient, hdr.Version)
	n := len(p.data) - len(data) // length of the header
	p.data = slices.Grow(p.data, 16)
	_ = sealer.Seal(p.data[n:n], p.data[n:], extHdr.PacketNumber, p.data[:n])
	p.data = p.data[:len(p.data)+16]
	sealer.EncryptHeader(p.data[n:n+16], &p.data[0], p.data[n-int(extHdr.PacketNumberLen):n])
	return p
}

func randConnID(l int) protocol.ConnectionID {
	b := make([]byte, l)
	rand.Read(b)
	return protocol.ParseConnectionID(b)
}

func getValidInitialPacket(t *testing.T, raddr net.Addr, srcConnID, destConnID protocol.ConnectionID) receivedPacket {
	t.Helper()
	return getLongHeaderPacket(t,
		raddr,
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketTypeInitial,
				SrcConnectionID:  srcConnID,
				DestConnectionID: destConnID,
				Length:           protocol.MinInitialPacketSize,
				Version:          protocol.Version1,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, protocol.MinInitialPacketSize),
	)
}

type sentPacketCallArgs struct {
	hdr    *logging.Header
	frames []logging.Frame
}

// checkConnectionClose checks
//  1. the arguments of the SentPacket tracer call, and
//  2. reads and parses the packet sent by the server
func checkConnectionClose(
	t *testing.T,
	conn *net.UDPConn,
	argChan <-chan sentPacketCallArgs,
	expectedSrcConnID protocol.ConnectionID,
	expectedDestConnID protocol.ConnectionID,
	expectedErrorCode qerr.TransportErrorCode,
) {
	t.Helper()
	select {
	case args := <-argChan:
		require.Equal(t, protocol.PacketTypeInitial, args.hdr.Type)
		require.Equal(t, expectedSrcConnID, args.hdr.SrcConnectionID)
		require.Equal(t, expectedDestConnID, args.hdr.DestConnectionID)
		require.Len(t, args.frames, 1)
		require.IsType(t, &wire.ConnectionCloseFrame{}, args.frames[0])
		ccf := args.frames[0].(*logging.ConnectionCloseFrame)
		require.EqualValues(t, expectedErrorCode, ccf.ErrorCode)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	b := make([]byte, 1500)
	n, _, err := conn.ReadFromUDP(b)
	require.NoError(t, err)
	parsedHdr, _, _, err := wire.ParsePacket(b[:n])
	require.NoError(t, err)
	require.Equal(t, protocol.PacketTypeInitial, parsedHdr.Type)
	require.Equal(t, expectedSrcConnID, parsedHdr.SrcConnectionID)
	require.Equal(t, expectedDestConnID, parsedHdr.DestConnectionID)
}

func checkRetry(t *testing.T,
	conn *net.UDPConn,
	argChan <-chan sentPacketCallArgs,
	expectedDestConnID protocol.ConnectionID,
) {
	t.Helper()
	select {
	case args := <-argChan:
		require.Equal(t, protocol.PacketTypeRetry, args.hdr.Type)
		require.Equal(t, expectedDestConnID, args.hdr.DestConnectionID)
		require.NotNil(t, args.hdr.Token)
		require.Empty(t, args.frames)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	conn.SetReadDeadline(time.Now().Add(time.Second))
	b := make([]byte, 1500)
	n, _, err := conn.ReadFromUDP(b)
	require.NoError(t, err)
	parsedHdr, _, _, err := wire.ParsePacket(b[:n])
	require.NoError(t, err)
	require.Equal(t, protocol.PacketTypeRetry, parsedHdr.Type)
	require.Equal(t, expectedDestConnID, parsedHdr.DestConnectionID)
	require.NotNil(t, parsedHdr.Token)
}

func TestListen(t *testing.T) {
	_, err := ListenAddr("localhost:0", nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "quic: tls.Config not set")

	_, err = Listen(nil, &tls.Config{}, &Config{Versions: []protocol.Version{0x1234}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid QUIC version: 0x1234")
}

func TestListenAddr(t *testing.T) {
	_, err := ListenAddr("127.0.0.1", &tls.Config{}, &Config{})
	require.Error(t, err)
	require.IsType(t, &net.AddrError{}, err)

	_, err = ListenAddr("1.1.1.1:1111", &tls.Config{}, &Config{})
	require.Error(t, err)
	require.IsType(t, &net.OpError{}, err)

	ln, err := ListenAddr("127.0.0.1:0", &tls.Config{}, &Config{})
	require.NoError(t, err)
	defer ln.Close()
}

func TestServerPacketDropping(t *testing.T) {
	t.Run("destination connection ID too short", func(t *testing.T) {
		conn := newUDPConnLocalhost(t)
		testServerDroppedPacket(t,
			conn,
			getValidInitialPacket(t, conn.LocalAddr(), randConnID(5), randConnID(7)),
			logging.PacketTypeInitial,
			logging.PacketDropUnexpectedPacket,
		)
	})

	t.Run("Initial packet too small", func(t *testing.T) {
		conn := newUDPConnLocalhost(t)
		p := getLongHeaderPacket(t,
			conn.LocalAddr(),
			&wire.ExtendedHeader{
				Header: wire.Header{
					Type:             protocol.PacketTypeInitial,
					DestConnectionID: randConnID(8),
					Version:          protocol.Version1,
				},
				PacketNumberLen: 2,
			},
			make([]byte, protocol.MinInitialPacketSize-100),
		)
		require.Greater(t, len(p.data), protocol.MinInitialPacketSize-100)
		require.Less(t, len(p.data), protocol.MinInitialPacketSize)
		testServerDroppedPacket(t,
			conn,
			p,
			logging.PacketTypeInitial,
			logging.PacketDropUnexpectedPacket,
		)
	})

	// we should not send a Version Negotiation packet if the packet is smaller than 1200 bytes
	t.Run("packet of unknown version, too small", func(t *testing.T) {
		conn := newUDPConnLocalhost(t)
		p := getLongHeaderPacket(t,
			conn.LocalAddr(),
			&wire.ExtendedHeader{
				Header: wire.Header{
					Type:             protocol.PacketTypeInitial,
					DestConnectionID: randConnID(8),
					Version:          0x42,
				},
				PacketNumberLen: 2,
			},
			make([]byte, protocol.MinUnknownVersionPacketSize-100),
		)
		require.Greater(t, len(p.data), protocol.MinUnknownVersionPacketSize-100)
		require.Less(t, len(p.data), protocol.MinUnknownVersionPacketSize)
		testServerDroppedPacket(t,
			conn,
			p,
			logging.PacketTypeNotDetermined,
			logging.PacketDropUnexpectedPacket,
		)
	})

	t.Run("not an Initial packet", func(t *testing.T) {
		conn := newUDPConnLocalhost(t)
		testServerDroppedPacket(t,
			conn,
			getLongHeaderPacket(t,
				conn.LocalAddr(),
				&wire.ExtendedHeader{
					Header: wire.Header{
						Type:             protocol.PacketTypeHandshake,
						DestConnectionID: protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8}),
						Version:          protocol.Version1,
					},
					PacketNumberLen: 2,
				},
				nil,
			),
			logging.PacketTypeHandshake,
			logging.PacketDropUnexpectedPacket,
		)
	})

	// as a server, we should never receive a Version Negotiation packet
	t.Run("Version Negotiation packet", func(t *testing.T) {
		conn := newUDPConnLocalhost(t)
		data := wire.ComposeVersionNegotiation(
			protocol.ArbitraryLenConnectionID{1, 2, 3, 4},
			protocol.ArbitraryLenConnectionID{4, 3, 2, 1},
			[]protocol.Version{1, 2, 3},
		)
		testServerDroppedPacket(t,
			conn,
			receivedPacket{
				remoteAddr: conn.LocalAddr(),
				data:       data,
				buffer:     getPacketBuffer(),
			},
			logging.PacketTypeVersionNegotiation,
			logging.PacketDropUnexpectedPacket,
		)
	})
}

func testServerDroppedPacket(t *testing.T,
	conn *net.UDPConn,
	p receivedPacket,
	expectedPacketType logging.PacketType,
	expectedDropReason logging.PacketDropReason,
) {
	readChan := make(chan struct{})
	go func() {
		defer close(readChan)
		conn.ReadFrom(make([]byte, 1000))
	}()
	mockCtrl := gomock.NewController(t)
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{tracer: tracer})

	mockTracer.EXPECT().DroppedPacket(p.remoteAddr, expectedPacketType, p.Size(), expectedDropReason)
	server.handlePacket(p)

	select {
	case <-readChan:
		t.Fatal("didn't expect to receive a packet")
	case <-time.After(scaleDuration(5 * time.Millisecond)):
	}
}

func TestServerVersionNegotiation(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		testServerVersionNegotiation(t, true)
	})
	t.Run("disabled", func(t *testing.T) {
		testServerVersionNegotiation(t, false)
	})
}

func testServerVersionNegotiation(t *testing.T, enabled bool) {
	mockCtrl := gomock.NewController(t)
	conn := newUDPConnLocalhost(t)
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{
		tracer:                    tracer,
		disableVersionNegotiation: !enabled,
	})

	srcConnID := protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5})
	destConnID := protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6})
	packet := getLongHeaderPacket(t, conn.LocalAddr(),
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketTypeHandshake,
				SrcConnectionID:  srcConnID,
				DestConnectionID: destConnID,
				Version:          0x42,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, protocol.MinUnknownVersionPacketSize),
	)

	switch enabled {
	case true:
		mockTracer.EXPECT().SentVersionNegotiationPacket(packet.remoteAddr, gomock.Any(), gomock.Any(), gomock.Any())
	case false:
		mockTracer.EXPECT().DroppedPacket(packet.remoteAddr, logging.PacketTypeNotDetermined, packet.Size(), logging.PacketDropUnexpectedVersion)
	}
=======
func (s *mockSession) run() error {
	<-s.stopRunLoop
	return s.closeReason
}
func (s *mockSession) WaitUntilHandshakeComplete() error {
	return <-s.handshakeComplete
}
func (s *mockSession) Close(e error) error {
	if s.closed {
		return nil
	}
	s.closeReason = e
	s.closed = true
	close(s.stopRunLoop)
	return nil
}
func (s *mockSession) closeRemote(e error) {
	s.closeReason = e
	s.closed = true
	s.closedRemote = true
	close(s.stopRunLoop)
}
func (s *mockSession) OpenStream() (Stream, error) {
	return &stream{streamID: 1337}, nil
}
func (s *mockSession) AcceptStream() (Stream, error)    { panic("not implemented") }
func (s *mockSession) OpenStreamSync() (Stream, error)  { panic("not implemented") }
func (s *mockSession) LocalAddr() net.Addr              { panic("not implemented") }
func (s *mockSession) RemoteAddr() net.Addr             { return s.remoteAddr }
func (*mockSession) Context() context.Context           { panic("not implemented") }
func (*mockSession) GetVersion() protocol.VersionNumber { return protocol.VersionWhatever }

var _ Session = &mockSession{}
var _ NonFWSession = &mockSession{}

func newMockSession(
	_ connection,
	_ *pconnManager,
	_ bool,
	_ protocol.VersionNumber,
	connectionID protocol.ConnectionID,
	_ *handshake.ServerConfig,
	_ *tls.Config,
	_ *Config,
) (packetHandler, <-chan handshakeEvent, error) {
	s := mockSession{
		connectionID:      connectionID,
		handshakeChan:     make(chan handshakeEvent),
		handshakeComplete: make(chan error),
		stopRunLoop:       make(chan struct{}),
	}
	return &s, s.handshakeChan, nil
}

var _ = Describe("Server", func() {
	var (
		conn     *mockPacketConn
		config   *Config
		pconnMgr *pconnManager
		udpAddr  = &net.UDPAddr{IP: net.IPv4(192, 168, 100, 200), Port: 1337}
	)

	BeforeEach(func() {
		pconnMgr = &pconnManager{}
		conn = &mockPacketConn{addr: &net.UDPAddr{}}
		pconnMgr.setup(conn, nil, nil)
		config = &Config{Versions: protocol.SupportedVersions}
	})
>>>>>>> project-faster/main

	written := make(chan []byte, 1)
	go func() {
		b := make([]byte, 1500)
		n, _, _ := conn.ReadFrom(b)
		written <- b[:n]
	}()
	server.handlePacket(packet)

<<<<<<< HEAD
	switch enabled {
	case true:
		select {
		case b := <-written:
			require.True(t, wire.IsVersionNegotiationPacket(b))
			dest, src, versions, err := wire.ParseVersionNegotiationPacket(b)
			require.NoError(t, err)
			require.Equal(t, protocol.ArbitraryLenConnectionID(srcConnID.Bytes()), dest)
			require.Equal(t, protocol.ArbitraryLenConnectionID(destConnID.Bytes()), src)
			require.NotContains(t, versions, protocol.Version(0x42))
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	case false:
		select {
		case <-written:
			t.Fatal("expected no version negotiation packet")
		case <-time.After(scaleDuration(10 * time.Millisecond)):
		}
	}
}

func TestServerRetry(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{tracer: tracer, useRetry: true})

	conn := newUDPConnLocalhost(t)

	packet := getLongHeaderPacket(t, conn.LocalAddr(),
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketTypeInitial,
				SrcConnectionID:  protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5}),
				DestConnectionID: protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				Version:          protocol.Version1,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, protocol.MinUnknownVersionPacketSize),
	)
	argsChan := make(chan sentPacketCallArgs, 1)
	mockTracer.EXPECT().SentPacket(packet.remoteAddr, gomock.Any(), gomock.Any(), nil).Do(
		func(_ net.Addr, hdr *logging.Header, _ logging.ByteCount, _ []logging.Frame) {
			argsChan <- sentPacketCallArgs{hdr: hdr}
		},
	)
	server.handlePacket(packet)
	checkRetry(t, conn, argsChan, protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5}))
}

func TestServerTokenValidation(t *testing.T) {
	var tokenGeneratorKey handshake.TokenProtectorKey
	rand.Read(tokenGeneratorKey[:])
	tg := handshake.NewTokenGenerator(tokenGeneratorKey)

	t.Run("retry token with invalid address", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		token, err := tg.NewRetryToken(
			&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1337},
			protocol.ConnectionID{},
			protocol.ConnectionID{},
		)
		require.NoError(t, err)
		tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
		server := newTestServer(t, &serverOpts{
			useRetry:          true,
			tracer:            tracer,
			tokenGeneratorKey: tokenGeneratorKey,
		})

		testServerTokenValidation(t, server, mockTracer, newUDPConnLocalhost(t), token, false, true, false)
	})

	t.Run("expired retry token", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		conn := newUDPConnLocalhost(t)
		tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
		server := newTestServer(t, &serverOpts{
			useRetry:          true,
			tracer:            tracer,
			config:            &Config{HandshakeIdleTimeout: time.Millisecond / 2},
			tokenGeneratorKey: tokenGeneratorKey,
		})
=======
		BeforeEach(func() {
			serv = &server{
				sessions:     make(map[protocol.ConnectionID]packetHandler),
				newSession:   newMockSession,
				pconnMgr:     pconnMgr,
				config:       config,
				sessionQueue: make(chan Session, 5),
				errorChan:    make(chan struct{}),
			}
			b := &bytes.Buffer{}
			utils.LittleEndian.WriteUint32(b, protocol.VersionNumberToTag(protocol.SupportedVersions[0]))
			firstPacket = []byte{0x09, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c}
			firstPacket = append(append(firstPacket, b.Bytes()...), 0x01)
		})

		It("returns the address", func() {
			conn.addr = &net.UDPAddr{
				IP:   net.IPv4(192, 168, 13, 37),
				Port: 1234,
			}
			Expect(serv.Addr().String()).To(Equal("192.168.13.37:1234"))
		})

		It("creates new sessions", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			sess := serv.sessions[connID].(*mockSession)
			Expect(sess.connectionID).To(Equal(connID))
			Expect(sess.packetCount).To(Equal(1))
		})

		It("accepts a session once the connection it is forward secure", func(done Done) {
			var acceptedSess Session
			go func() {
				defer GinkgoRecover()
				var err error
				acceptedSess, err = serv.Accept()
				Expect(err).ToNot(HaveOccurred())
			}()
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			sess := serv.sessions[connID].(*mockSession)
			sess.handshakeChan <- handshakeEvent{encLevel: protocol.EncryptionSecure}
			Consistently(func() Session { return acceptedSess }).Should(BeNil())
			sess.handshakeChan <- handshakeEvent{encLevel: protocol.EncryptionForwardSecure}
			Eventually(func() Session { return acceptedSess }).Should(Equal(sess))
			close(done)
		}, 0.5)

		It("doesn't accept session that error during the handshake", func(done Done) {
			var accepted bool
			go func() {
				defer GinkgoRecover()
				serv.Accept()
				accepted = true
			}()
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			sess := serv.sessions[connID].(*mockSession)
			sess.handshakeChan <- handshakeEvent{err: errors.New("handshake failed")}
			Consistently(func() bool { return accepted }).Should(BeFalse())
			close(done)
		})

		It("assigns packets to existing sessions", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			err = serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: []byte{0x08, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0x01}, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID].(*mockSession).connectionID).To(Equal(connID))
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(2))
		})

		It("closes and deletes sessions", func() {
			serv.deleteClosedSessionsAfter = time.Second // make sure that the nil value for the closed session doesn't get deleted in this test
			nullAEAD := crypto.NewNullAEAD(protocol.PerspectiveServer, protocol.VersionWhatever)
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: append(firstPacket, nullAEAD.Seal(nil, nil, 0, firstPacket)...), rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID]).ToNot(BeNil())
			// make session.run() return
			serv.sessions[connID].(*mockSession).stopRunLoop <- struct{}{}
			// The server should now have closed the session, leaving a nil value in the sessions map
			Consistently(func() map[protocol.ConnectionID]packetHandler { return serv.sessions }).Should(HaveLen(1))
			Expect(serv.sessions[connID]).To(BeNil())
		})

		It("deletes nil session entries after a wait time", func() {
			serv.deleteClosedSessionsAfter = 25 * time.Millisecond
			nullAEAD := crypto.NewNullAEAD(protocol.PerspectiveServer, protocol.VersionWhatever)
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: append(firstPacket, nullAEAD.Seal(nil, nil, 0, firstPacket)...), rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions).To(HaveKey(connID))
			// make session.run() return
			serv.sessions[connID].(*mockSession).stopRunLoop <- struct{}{}
			Eventually(func() bool {
				serv.sessionsMutex.Lock()
				_, ok := serv.sessions[connID]
				serv.sessionsMutex.Unlock()
				return ok
			}).Should(BeFalse())
		})

		It("closes sessions and the connection when Close is called", func() {
			session, _, _ := newMockSession(nil, pconnMgr, false, 0, 0, nil, nil, nil)
			serv.sessions[1] = session
			err := serv.Close()
			Expect(err).NotTo(HaveOccurred())
			Expect(session.(*mockSession).closed).To(BeTrue())
			Expect(conn.closed).To(BeTrue())
		})

		It("ignores packets for closed sessions", func() {
			serv.sessions[connID] = nil
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: []byte{0x08, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0x01}, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID]).To(BeNil())
		})

		It("works if no quic.Config is given", func(done Done) {
			ln, err := ListenAddr("127.0.0.1:0", nil, config)
			Expect(err).ToNot(HaveOccurred())
			Expect(ln.Close()).To(Succeed())
			close(done)
		}, 1)

		It("closes properly", func() {
			ln, err := ListenAddr("127.0.0.1:0", nil, config)
			Expect(err).ToNot(HaveOccurred())

			var returned bool
			go func() {
				defer GinkgoRecover()
				_, err := ln.Accept()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("use of closed network connection"))
				returned = true
			}()
			ln.Close()
			Eventually(func() bool { return returned }).Should(BeTrue())
		})

		It("errors when encountering a connection error", func(done Done) {
			testErr := errors.New("connection error")
			conn.readErr = testErr
			go serv.serve()
			_, err := serv.Accept()
			Expect(err).To(MatchError(testErr))
			Expect(serv.Close()).To(Succeed())
			close(done)
		}, 0.5)

		It("closes all sessions when encountering a connection error", func() {
			session, _, _ := newMockSession(nil, pconnMgr, false, 0, 0, nil, nil, nil)
			serv.sessions[0x12345] = session
			Expect(serv.sessions[0x12345].(*mockSession).closed).To(BeFalse())
			testErr := errors.New("connection error")
			conn.readErr = testErr
			go serv.serve()
			Eventually(func() Session { return serv.sessions[connID] }).Should(BeNil())
			Eventually(func() bool { return session.(*mockSession).closed }).Should(BeTrue())
			Expect(serv.Close()).To(Succeed())
		})

		It("ignores delayed packets with mismatching versions", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(1))
			b := &bytes.Buffer{}
			// add an unsupported version
			utils.LittleEndian.WriteUint32(b, protocol.VersionNumberToTag(protocol.SupportedVersions[0]+1))
			data := []byte{0x09, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c}
			data = append(append(data, b.Bytes()...), 0x01)
			err = serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: data, rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			// if we didn't ignore the packet, the server would try to send a version negotation packet, which would make the test panic because it doesn't have a udpConn
			Expect(conn.dataWritten.Bytes()).To(BeEmpty())
			// make sure the packet was *not* passed to session.handlePacket()
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(1))
		})

		It("errors on invalid public header", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: nil, rcvTime: time.Now()})
			Expect(err.(*qerr.QuicError).ErrorCode).To(Equal(qerr.InvalidPacketHeader))
		})

		It("ignores public resets for unknown connections", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: wire.WritePublicReset(999, 1, 1337), rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(BeEmpty())
		})

		It("ignores public resets for known connections", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(1))
			err = serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: wire.WritePublicReset(connID, 1, 1337), rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(1))
		})

		It("ignores invalid public resets for known connections", func() {
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: firstPacket, rcvTime: time.Now()})
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(1))
			data := wire.WritePublicReset(connID, 1, 1337)
			err = serv.handlePacket(&receivedRawPacket{rcvPconn: nil, remoteAddr: nil, data: data[:len(data)-2], rcvTime: time.Now()})
			Expect(err).ToNot(HaveOccurred())
			Expect(serv.sessions).To(HaveLen(1))
			Expect(serv.sessions[connID].(*mockSession).packetCount).To(Equal(1))
		})

		It("doesn't respond with a version negotiation packet if the first packet is too small", func() {
			b := &bytes.Buffer{}
			hdr := wire.PublicHeader{
				VersionFlag:     true,
				ConnectionID:    0x1337,
				PacketNumber:    1,
				PacketNumberLen: protocol.PacketNumberLen2,
			}
			hdr.Write(b, 13 /* not a valid QUIC version */, protocol.PerspectiveClient)
			b.Write(bytes.Repeat([]byte{0}, protocol.ClientHelloMinimumSize-1)) // this packet is 1 byte too small
			err := serv.handlePacket(&receivedRawPacket{rcvPconn: conn, remoteAddr: udpAddr, data: b.Bytes(), rcvTime: time.Now()})
			Expect(err).To(MatchError("dropping small packet with unknown version"))
			Expect(conn.dataWritten.Len()).Should(BeZero())
		})
	})

	It("setups with the right values", func() {
		supportedVersions := []protocol.VersionNumber{1, 3, 5}
		acceptCookie := func(_ net.Addr, _ *Cookie) bool { return true }
		config := Config{
			Versions:         supportedVersions,
			AcceptCookie:     acceptCookie,
			HandshakeTimeout: 1337 * time.Hour,
			IdleTimeout:      42 * time.Minute,
			KeepAlive:        true,
		}
		ln, err := Listen(conn, &tls.Config{}, &config)
		Expect(err).ToNot(HaveOccurred())
		server := ln.(*server)
		Expect(server.deleteClosedSessionsAfter).To(Equal(protocol.ClosedSessionDeleteTimeout))
		Expect(server.sessions).ToNot(BeNil())
		Expect(server.scfg).ToNot(BeNil())
		Expect(server.config.Versions).To(Equal(supportedVersions))
		Expect(server.config.HandshakeTimeout).To(Equal(1337 * time.Hour))
		Expect(server.config.IdleTimeout).To(Equal(42 * time.Minute))
		Expect(reflect.ValueOf(server.config.AcceptCookie)).To(Equal(reflect.ValueOf(acceptCookie)))
		Expect(server.config.KeepAlive).To(BeTrue())
	})

	It("fills in default values if options are not set in the Config", func() {
		ln, err := Listen(conn, &tls.Config{}, &Config{})
		Expect(err).ToNot(HaveOccurred())
		server := ln.(*server)
		Expect(server.config.Versions).To(Equal(protocol.SupportedVersions))
		Expect(server.config.HandshakeTimeout).To(Equal(protocol.DefaultHandshakeTimeout))
		Expect(server.config.IdleTimeout).To(Equal(protocol.DefaultIdleTimeout))
		Expect(reflect.ValueOf(server.config.AcceptCookie)).To(Equal(reflect.ValueOf(defaultAcceptCookie)))
		Expect(server.config.KeepAlive).To(BeFalse())
	})

	It("listens on a given address", func() {
		addr := "127.0.0.1:13579"
		ln, err := ListenAddr(addr, nil, config)
		Expect(err).ToNot(HaveOccurred())
		serv := ln.(*server)
		Expect(serv.Addr().String()).To(Equal(addr))
	})

	It("errors if given an invalid address", func() {
		addr := "127.0.0.1"
		_, err := ListenAddr(addr, nil, config)
		Expect(err).To(BeAssignableToTypeOf(&net.AddrError{}))
	})

	It("errors if given an invalid address", func() {
		addr := "1.1.1.1:1111"
		_, err := ListenAddr(addr, nil, config)
		Expect(err).To(BeAssignableToTypeOf(&net.OpError{}))
	})

	It("setups and responds with version negotiation", func() {
		config.Versions = []protocol.VersionNumber{99}
		b := &bytes.Buffer{}
		hdr := wire.PublicHeader{
			VersionFlag:     true,
			ConnectionID:    0x1337,
			PacketNumber:    1,
			PacketNumberLen: protocol.PacketNumberLen2,
		}
		hdr.Write(b, 13 /* not a valid QUIC version */, protocol.PerspectiveClient)
		b.Write(bytes.Repeat([]byte{0}, protocol.ClientHelloMinimumSize)) // add a fake CHLO
		conn.dataToRead = b.Bytes()
		conn.dataReadFrom = udpAddr
		ln, err := ListenImpl(conn, nil, config, pconnMgr)
		Expect(err).ToNot(HaveOccurred())

		var returned bool
		go func() {
			ln.Accept()
			returned = true
		}()

		Eventually(func() int { return conn.dataWritten.Len() }).ShouldNot(BeZero())
		Expect(conn.dataWrittenTo).To(Equal(udpAddr))
		b = &bytes.Buffer{}
		utils.LittleEndian.WriteUint32(b, protocol.VersionNumberToTag(99))
		expected := append(
			[]byte{0x9, 0x37, 0x13, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			b.Bytes()...,
		)
		Expect(conn.dataWritten.Bytes()).To(Equal(expected))
		Consistently(func() bool { return returned }).Should(BeFalse())
	})

	It("sends a PublicReset for new connections that don't have the VersionFlag set", func() {
		conn.dataReadFrom = udpAddr
		conn.dataToRead = []byte{0x08, 0xf6, 0x19, 0x86, 0x66, 0x9b, 0x9f, 0xfa, 0x4c, 0x01}
		ln, err := ListenImpl(conn, nil, config, pconnMgr)
		Expect(err).ToNot(HaveOccurred())
		go func() {
			defer GinkgoRecover()
			_, err := ln.Accept()
			Expect(err).ToNot(HaveOccurred())
		}()
>>>>>>> project-faster/main

		token, err := tg.NewRetryToken(conn.LocalAddr(), protocol.ConnectionID{}, protocol.ConnectionID{})
		require.NoError(t, err)
		// the maximum retry token age is equivalent to the handshake timeout
		time.Sleep(time.Millisecond) // make sure the token is expired
		testServerTokenValidation(t, server, mockTracer, conn, token, false, true, false)
	})
<<<<<<< HEAD

	// if the packet is corrupted, it will just be dropped (no INVALID_TOKEN nor Retry is sent)
	t.Run("corrupted packet", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
		server := newTestServer(t, &serverOpts{
			useRetry:          true,
			tracer:            tracer,
			config:            &Config{HandshakeIdleTimeout: time.Millisecond / 2},
			tokenGeneratorKey: tokenGeneratorKey,
		})

		conn := newUDPConnLocalhost(t)
		token, err := tg.NewRetryToken(conn.LocalAddr(), protocol.ConnectionID{}, protocol.ConnectionID{})
		require.NoError(t, err)
		time.Sleep(time.Millisecond) // make sure the token is expired
		testServerTokenValidation(t, server, mockTracer, conn, token, true, false, true)
	})

	t.Run("invalid non-retry token", func(t *testing.T) {
		var tokenGeneratorKey2 handshake.TokenProtectorKey
		rand.Read(tokenGeneratorKey2[:])
		mockCtrl := gomock.NewController(t)
		tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
		server := newTestServer(t, &serverOpts{
			tokenGeneratorKey: tokenGeneratorKey2, // use a different key
			useRetry:          true,
			tracer:            tracer,
			maxTokenAge:       time.Millisecond,
		})

		conn := newUDPConnLocalhost(t)
		token, err := tg.NewToken(conn.LocalAddr())
		require.NoError(t, err)
		time.Sleep(3 * time.Millisecond) // make sure the token is expired
		testServerTokenValidation(t, server, mockTracer, conn, token, false, false, true)
	})

	t.Run("expired non-retry token", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
		server := newTestServer(t, &serverOpts{
			tokenGeneratorKey: tokenGeneratorKey,
			useRetry:          true,
			tracer:            tracer,
			maxTokenAge:       time.Millisecond,
		})

		conn := newUDPConnLocalhost(t)
		token, err := tg.NewToken(conn.LocalAddr())
		require.NoError(t, err)
		time.Sleep(3 * time.Millisecond) // make sure the token is expired
		testServerTokenValidation(t, server, mockTracer, conn, token, false, false, true)
	})
}

func testServerTokenValidation(
	t *testing.T,
	server *testServer,
	mockTracer *mocklogging.MockTracer,
	conn *net.UDPConn,
	token []byte,
	corruptedPacket bool,
	expectInvalidTokenConnectionClose bool,
	expectRetry bool,
) {
	hdr := wire.Header{
		Type:             protocol.PacketTypeInitial,
		SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
		DestConnectionID: protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		Token:            token,
		Length:           protocol.MinInitialPacketSize + protocol.ByteCount(protocol.PacketNumberLen4) + 16,
		Version:          protocol.Version1,
	}
	packet := getLongHeaderPacketEncrypted(t,
		conn.LocalAddr(),
		&wire.ExtendedHeader{Header: hdr, PacketNumberLen: protocol.PacketNumberLen4},
		make([]byte, protocol.MinInitialPacketSize),
	)
	if corruptedPacket {
		packet.data[len(packet.data)-10] ^= 0xff // corrupt the packet
		done := make(chan struct{})
		mockTracer.EXPECT().DroppedPacket(packet.remoteAddr, logging.PacketTypeInitial, packet.Size(), logging.PacketDropPayloadDecryptError).Do(
			func(net.Addr, logging.PacketType, protocol.ByteCount, logging.PacketDropReason) {
				close(done)
			},
		)
		server.handlePacket(packet)
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
		return
	}

	argsChan := make(chan sentPacketCallArgs, 1)
	if expectInvalidTokenConnectionClose || expectRetry {
		mockTracer.EXPECT().SentPacket(packet.remoteAddr, gomock.Any(), gomock.Any(), gomock.Any()).Do(
			func(_ net.Addr, hdr *logging.Header, _ logging.ByteCount, frames []logging.Frame) {
				argsChan <- sentPacketCallArgs{hdr: hdr, frames: frames}
			},
		)
	}

	server.handlePacket(packet)
	if expectInvalidTokenConnectionClose {
		checkConnectionClose(t, conn, argsChan, hdr.DestConnectionID, hdr.SrcConnectionID, qerr.InvalidToken)
	}
	if expectRetry {
		checkRetry(t, conn, argsChan, hdr.SrcConnectionID)
	}
}

type connConstructorArgs struct {
	ctx              context.Context
	config           *Config
	origDestConnID   protocol.ConnectionID
	retrySrcConnID   *protocol.ConnectionID
	clientDestConnID protocol.ConnectionID
	destConnID       protocol.ConnectionID
	srcConnID        protocol.ConnectionID
}

type connConstructorRecorder struct {
	ch chan connConstructorArgs

	conns []quicConn
}

func newConnConstructorRecorder(conns ...quicConn) *connConstructorRecorder {
	return &connConstructorRecorder{
		ch:    make(chan connConstructorArgs, len(conns)),
		conns: conns,
	}
}

func (r *connConstructorRecorder) Args() <-chan connConstructorArgs { return r.ch }

func (r *connConstructorRecorder) NewConn(
	ctx context.Context,
	_ context.CancelCauseFunc,
	_ sendConn,
	_ *Transport,
	origDestConnID protocol.ConnectionID,
	retrySrcConnID *protocol.ConnectionID,
	clientDestConnID protocol.ConnectionID,
	destConnID protocol.ConnectionID,
	srcConnID protocol.ConnectionID,
	_ ConnectionIDGenerator,
	_ *statelessResetter,
	config *Config,
	_ *tls.Config,
	_ *handshake.TokenGenerator,
	_ bool,
	_ *logging.ConnectionTracer,
	_ utils.Logger,
	_ protocol.Version,
) quicConn {
	r.ch <- connConstructorArgs{
		ctx:              ctx,
		config:           config,
		origDestConnID:   origDestConnID,
		retrySrcConnID:   retrySrcConnID,
		clientDestConnID: clientDestConnID,
		destConnID:       destConnID,
		srcConnID:        srcConnID,
	}
	c := r.conns[0]
	r.conns = r.conns[1:]
	return c
}

func TestServerCreateConnection(t *testing.T) {
	t.Run("without retry", func(t *testing.T) {
		testServerCreateConnection(t, false)
	})
	t.Run("with retry", func(t *testing.T) {
		testServerCreateConnection(t, true)
	})
}

func testServerCreateConnection(t *testing.T, useRetry bool) {
	mockCtrl := gomock.NewController(t)
	tokenGeneratorKey := TokenGeneratorKey{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	tg := handshake.NewTokenGenerator(tokenGeneratorKey)

	server := newTestServer(t, &serverOpts{
		useRetry:          useRetry,
		tokenGeneratorKey: tokenGeneratorKey,
	})

	done := make(chan struct{}, 3)
	c := NewMockQUICConn(mockCtrl)
	c.EXPECT().run().Do(func() error { done <- struct{}{}; return nil })
	c.EXPECT().Context().DoAndReturn(func() context.Context {
		done <- struct{}{}
		return context.Background()
	})
	c.EXPECT().HandshakeComplete().DoAndReturn(func() <-chan struct{} {
		done <- struct{}{}
		return make(chan struct{})
	})
	recorder := newConnConstructorRecorder(c)
	server.newConn = recorder.NewConn

	conn := newUDPConnLocalhost(t)
	var token []byte
	if useRetry {
		var err error
		token, err = tg.NewRetryToken(
			conn.LocalAddr(),
			protocol.ParseConnectionID([]byte{0xde, 0xad, 0xc0, 0xde}),
			protocol.ParseConnectionID([]byte{0xde, 0xca, 0xfb, 0xad}),
		)
		require.NoError(t, err)
	}
	hdr := wire.Header{
		Type:             protocol.PacketTypeInitial,
		SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
		DestConnectionID: protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
		Length:           protocol.MinInitialPacketSize + protocol.ByteCount(protocol.PacketNumberLen4) + 16,
		Token:            token,
		Version:          protocol.Version1,
	}
	packet := getLongHeaderPacketEncrypted(t,
		conn.LocalAddr(),
		&wire.ExtendedHeader{Header: hdr, PacketNumberLen: protocol.PacketNumberLen4},
		make([]byte, protocol.MinInitialPacketSize),
	)
	c.EXPECT().handlePacket(packet)

	server.handlePacket(packet)

	var args connConstructorArgs
	select {
	case args = <-recorder.Args():
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.Equal(t, protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}), args.destConnID)
	assert.NotEqual(t, args.origDestConnID, args.srcConnID)
	if useRetry {
		assert.Equal(t, protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}), args.destConnID)
		assert.Equal(t, protocol.ParseConnectionID([]byte{0xde, 0xad, 0xc0, 0xde}), args.origDestConnID)
		assert.Equal(t, protocol.ParseConnectionID([]byte{0xde, 0xca, 0xfb, 0xad}), *args.retrySrcConnID)
	} else {
		assert.Equal(t, protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}), args.origDestConnID)
		assert.Zero(t, args.retrySrcConnID)
	}

	for range 3 {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	// shutdown
	c.EXPECT().closeWithTransportError(ConnectionRefused)
	c.EXPECT().destroy(gomock.Any()).AnyTimes()
}

func TestServerClose(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var conns []quicConn
	const numConns = 3
	done := make(chan struct{}, numConns)
	for range numConns {
		conn := NewMockQUICConn(mockCtrl)
		conn.EXPECT().run().MaxTimes(1)
		conn.EXPECT().handlePacket(gomock.Any()).MaxTimes(1)
		conn.EXPECT().Context().Return(context.Background()).MaxTimes(1)
		conn.EXPECT().HandshakeComplete().Return(make(chan struct{})).MaxTimes(1) // doesn't complete handshake
		conn.EXPECT().closeWithTransportError(ConnectionRefused).Do(func(TransportErrorCode) { done <- struct{}{} })
		conns = append(conns, conn)
	}
	recorder := newConnConstructorRecorder(conns...)
	server := newTestServer(t, &serverOpts{newConn: recorder.NewConn})

	for range numConns {
		b := make([]byte, 10)
		rand.Read(b)
		connID := protocol.ParseConnectionID(b)
		server.handlePacket(getValidInitialPacket(t,
			&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
			randConnID(6),
			connID,
		))
		select {
		case <-recorder.Args():
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	server.Close()
	// closing closes all handshakeing connections with CONNECTION_REFUSED
	for range numConns {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	// Accept returns ErrServerClosed after closing
	for range 5 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := server.Accept(ctx)
		require.ErrorIs(t, err, ErrServerClosed)
		require.ErrorIs(t, err, net.ErrClosed)
	}
}

func TestServerGetConfigForClientAccept(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	c := NewMockQUICConn(mockCtrl)
	c.EXPECT().run().MaxTimes(1)
	c.EXPECT().Context().Return(context.Background()).MaxTimes(1)
	c.EXPECT().HandshakeComplete().Return(make(chan struct{})).MaxTimes(1)
	recorder := newConnConstructorRecorder(c)
	server := newTestServer(t, &serverOpts{
		config: &Config{
			GetConfigForClient: func(*ClientInfo) (*Config, error) {
				return &Config{MaxIncomingStreams: 1234}, nil
			},
		},
		newConn: recorder.NewConn,
	})

	conn := newUDPConnLocalhost(t)
	packet := getValidInitialPacket(t,
		conn.LocalAddr(),
		protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
		protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8}),
	)
	c.EXPECT().handlePacket(packet).MaxTimes(1)

	server.handlePacket(packet)

	var args connConstructorArgs
	select {
	case args = <-recorder.Args():
		require.EqualValues(t, 1234, args.config.MaxIncomingStreams)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	assert.Equal(t, protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}), args.destConnID)
	assert.NotEqual(t, args.origDestConnID, args.srcConnID)

	// shutdown
	c.EXPECT().closeWithTransportError(ConnectionRefused)
	c.EXPECT().destroy(gomock.Any()).AnyTimes()
}

func TestServerGetConfigForClientReject(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{
		tracer: tracer,
		config: &Config{
			GetConfigForClient: func(*ClientInfo) (*Config, error) {
				return nil, errors.New("rejected")
			},
		},
	})

	conn := newUDPConnLocalhost(t)
	srcConnID := randConnID(6)
	destConnID := randConnID(8)
	p := getValidInitialPacket(t, conn.LocalAddr(), srcConnID, destConnID)
	argsChan := make(chan sentPacketCallArgs, 1)
	mockTracer.EXPECT().SentPacket(p.remoteAddr, gomock.Any(), gomock.Any(), gomock.Any()).Do(
		func(_ net.Addr, hdr *logging.Header, _ logging.ByteCount, frames []logging.Frame) {
			argsChan <- sentPacketCallArgs{hdr: hdr, frames: frames}
		},
	)
	server.handlePacket(p)

	checkConnectionClose(t, conn, argsChan, destConnID, srcConnID, qerr.ConnectionRefused)
}

func TestServerPacketHandling(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	server := newTestServer(t, &serverOpts{})

	srcConnID := randConnID(6)
	destConnID := randConnID(8)
	conn := NewMockQUICConn(mockCtrl)
	handledPacket := make(chan receivedPacket, 1)
	conn.EXPECT().handlePacket(gomock.Any()).Do(func(p receivedPacket) {
		handledPacket <- p
	})
	server.tr.handlerMap.Add(destConnID, conn)

	server.handlePacket(
		getValidInitialPacket(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42}, srcConnID, destConnID),
	)
	select {
	case p := <-handledPacket:
		require.Equal(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42}, p.remoteAddr)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	// shutdown
	conn.EXPECT().destroy(gomock.Any()).AnyTimes()
}

func TestServerReceiveQueue(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	acceptConn := make(chan struct{})
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{
		tracer: tracer,
		newConn: func(
			_ context.Context,
			_ context.CancelCauseFunc,
			_ sendConn,
			_ *Transport,
			_ protocol.ConnectionID,
			_ *protocol.ConnectionID,
			_ protocol.ConnectionID,
			_ protocol.ConnectionID,
			_ protocol.ConnectionID,
			_ ConnectionIDGenerator,
			_ *statelessResetter,
			_ *Config,
			_ *tls.Config,
			_ *handshake.TokenGenerator,
			_ bool,
			_ *logging.ConnectionTracer,
			_ utils.Logger,
			_ protocol.Version,
		) quicConn {
			<-acceptConn
			conn := NewMockQUICConn(mockCtrl)
			conn.EXPECT().handlePacket(gomock.Any()).MaxTimes(1)
			conn.EXPECT().run().MaxTimes(1)
			conn.EXPECT().Context().Return(context.Background()).MaxTimes(1)
			conn.EXPECT().HandshakeComplete().Return(make(chan struct{})).MaxTimes(1)
			conn.EXPECT().closeWithTransportError(gomock.Any()).MaxTimes(1)
			return conn
		},
	})

	conn := newUDPConnLocalhost(t)
	for range protocol.MaxServerUnprocessedPackets + 1 {
		server.handlePacket(getValidInitialPacket(t, conn.LocalAddr(), randConnID(6), randConnID(8)))
	}

	done := make(chan struct{})
	mockTracer.EXPECT().DroppedPacket(gomock.Any(), logging.PacketTypeNotDetermined, gomock.Any(), logging.PacketDropDOSPrevention).Do(
		func(_ net.Addr, _ logging.PacketType, _ logging.ByteCount, _ logging.PacketDropReason) {
			close(done)
		},
	)
	server.handlePacket(getValidInitialPacket(t, conn.LocalAddr(), randConnID(6), randConnID(8)))
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
	close(acceptConn)
}

func TestServerAccept(t *testing.T) {
	t.Run("without accept early", func(t *testing.T) {
		testServerAccept(t, false)
	})
	t.Run("with accept early", func(t *testing.T) {
		testServerAccept(t, true)
	})
}

func testServerAccept(t *testing.T, acceptEarly bool) {
	mockCtrl := gomock.NewController(t)
	ready := make(chan struct{})
	c := NewMockQUICConn(mockCtrl)
	c.EXPECT().run()
	c.EXPECT().handlePacket(gomock.Any())
	c.EXPECT().Context().Return(context.Background())
	if acceptEarly {
		c.EXPECT().earlyConnReady().Return(ready)
	} else {
		c.EXPECT().HandshakeComplete().Return(ready)
	}
	recorder := newConnConstructorRecorder(c)
	tracer, _ := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{
		acceptEarly: acceptEarly,
		tracer:      tracer,
		newConn:     recorder.NewConn,
	})

	// Accept should respect the context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := server.Accept(ctx)
	require.ErrorIs(t, err, context.Canceled)

	// establish a new connection, which then starts handshaking
	server.handlePacket(getValidInitialPacket(t,
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
		randConnID(6),
		randConnID(8),
	))

	accepted := make(chan error, 1)
	go func() {
		_, err := server.Accept(context.Background())
		accepted <- err
	}()

	select {
	case <-accepted:
		t.Fatal("server accepted the connection too early")
	case <-time.After(scaleDuration(5 * time.Millisecond)):
	}
	// now complete the handshake
	close(ready)

	select {
	case err := <-accepted:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestServerAcceptHandshakeFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := NewMockQUICConn(gomock.NewController(t))
	c.EXPECT().run()
	c.EXPECT().handlePacket(gomock.Any())
	c.EXPECT().Context().Return(ctx)
	c.EXPECT().HandshakeComplete().Return(make(chan struct{}))
	recorder := newConnConstructorRecorder(c)
	server := newTestServer(t, &serverOpts{newConn: recorder.NewConn})

	// establish a new connection, which then starts handshaking
	server.handlePacket(getValidInitialPacket(t,
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
		randConnID(6),
		randConnID(8),
	))

	accepted := make(chan error, 1)
	go func() {
		_, err := server.Accept(context.Background())
		accepted <- err
	}()

	cancel()
	select {
	case <-accepted:
		t.Fatal("server should not have accepted the connection")
	case <-time.After(scaleDuration(5 * time.Millisecond)):
	}
}

func TestServerAcceptQueue(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var conns []quicConn
	var rejectedConn *MockQUICConn
	for i := range protocol.MaxAcceptQueueSize + 2 {
		conn := NewMockQUICConn(mockCtrl)
		conn.EXPECT().handlePacket(gomock.Any())
		conn.EXPECT().run()
		c := make(chan struct{})
		close(c)
		conn.EXPECT().HandshakeComplete().Return(c)
		conn.EXPECT().Context().Return(context.Background())
		conns = append(conns, conn)
		if i == protocol.MaxAcceptQueueSize {
			rejectedConn = conn
			continue
		}
		defer func(conn *MockQUICConn) {
			conn.EXPECT().closeWithTransportError(ConnectionRefused).MaxTimes(1)
		}(conn)
	}
	recorder := newConnConstructorRecorder(conns...)
	server := newTestServer(t, &serverOpts{newConn: recorder.NewConn})

	for range protocol.MaxAcceptQueueSize {
		b := make([]byte, 16)
		rand.Read(b)
		connID := protocol.ParseConnectionID(b)
		server.handlePacket(
			getValidInitialPacket(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42}, randConnID(6), connID),
		)
		select {
		case args := <-recorder.Args():
			require.Equal(t, connID, args.origDestConnID)
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
	// wait for the connection to be enqueued
	time.Sleep(scaleDuration(10 * time.Millisecond))

	done := make(chan struct{})
	rejectedConn.EXPECT().closeWithTransportError(ConnectionRefused).Do(func(TransportErrorCode) {
		close(done)
	})
	server.handlePacket(
		getValidInitialPacket(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42}, randConnID(6), randConnID(8)),
	)
	select {
	case <-recorder.Args():
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	// accept one connection, freeing up one slot in the accept queue
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := server.Accept(ctx)
	require.NoError(t, err)

	// it's now possible to enqueue a new connection
	server.handlePacket(
		getValidInitialPacket(t,
			&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
			randConnID(6),
			protocol.ParseConnectionID([]byte{8, 7, 6, 5, 4, 3, 2, 1}),
		),
	)
	select {
	case args := <-recorder.Args():
		require.Equal(t, protocol.ParseConnectionID([]byte{8, 7, 6, 5, 4, 3, 2, 1}), args.origDestConnID)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestServer0RTTReordering(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	conn := NewMockQUICConn(mockCtrl)
	recorder := newConnConstructorRecorder(conn)
	server := newTestServer(t, &serverOpts{
		acceptEarly: true,
		tracer:      tracer,
		newConn:     recorder.NewConn,
	})

	connID := protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8})

	var zeroRTTPackets []receivedPacket

	for range protocol.Max0RTTQueueLen {
		p := getLongHeaderPacket(t,
			&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
			&wire.ExtendedHeader{
				Header: wire.Header{
					Type:             protocol.PacketType0RTT,
					SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
					DestConnectionID: connID,
					Length:           100,
					Version:          protocol.Version1,
				},
				PacketNumberLen: protocol.PacketNumberLen4,
			},
			make([]byte, 100),
		)
		server.handlePacket(p)
		zeroRTTPackets = append(zeroRTTPackets, p)
	}

	// send one more packet, this one should be dropped
	p := getLongHeaderPacket(t,
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketType0RTT,
				SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
				DestConnectionID: connID,
				Length:           100,
				Version:          protocol.Version1,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, 100),
	)
	mockTracer.EXPECT().DroppedPacket(p.remoteAddr, logging.PacketType0RTT, p.Size(), logging.PacketDropDOSPrevention)
	server.handlePacket(p)

	// now receive the Initial
	done := make(chan struct{})
	initial := getValidInitialPacket(t, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42}, randConnID(5), connID)
	packets := make(chan receivedPacket, protocol.Max0RTTQueueLen+1)
	conn.EXPECT().handlePacket(gomock.Any()).Do(func(p receivedPacket) { packets <- p }).AnyTimes()
	conn.EXPECT().Context().Return(context.Background())
	conn.EXPECT().earlyConnReady().Return(make(chan struct{}))
	conn.EXPECT().run().Do(func() error { close(done); return nil })
	server.handlePacket(initial)

	for i := range protocol.Max0RTTQueueLen + 1 {
		select {
		case p := <-packets:
			if i == 0 {
				require.Equal(t, initial.data, p.data)
			} else {
				require.Equal(t, zeroRTTPackets[i-1], p)
			}
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	// shutdown
	conn.EXPECT().closeWithTransportError(gomock.Any()).AnyTimes()
}

func TestServer0RTTQueueing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	tracer, mockTracer := mocklogging.NewMockTracer(mockCtrl)
	server := newTestServer(t, &serverOpts{
		acceptEarly: true,
		tracer:      tracer,
	})

	firstRcvTime := time.Now()
	otherRcvTime := firstRcvTime.Add(protocol.Max0RTTQueueingDuration / 2)
	var sizes []protocol.ByteCount
	for i := range protocol.Max0RTTQueues {
		b := make([]byte, 16)
		rand.Read(b)
		connID := protocol.ParseConnectionID(b)
		size := protocol.ByteCount(500 + i)
		p := getLongHeaderPacket(t,
			&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
			&wire.ExtendedHeader{
				Header: wire.Header{
					Type:             protocol.PacketType0RTT,
					SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
					DestConnectionID: connID,
					Length:           size,
					Version:          protocol.Version1,
				},
				PacketNumberLen: protocol.PacketNumberLen4,
			},
			make([]byte, size),
		)
		if i == 0 {
			p.rcvTime = firstRcvTime
		} else {
			p.rcvTime = otherRcvTime
		}
		sizes = append(sizes, p.Size())
		server.handlePacket(p)
	}

	// maximum number of 0-RTT queues is reached, further packets are dropped
	p := getLongHeaderPacket(t,
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketType0RTT,
				SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
				DestConnectionID: protocol.ParseConnectionID([]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				Length:           123,
				Version:          protocol.Version1,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, 123),
	)
	dropped := make(chan struct{}, protocol.Max0RTTQueues)
	mockTracer.EXPECT().DroppedPacket(p.remoteAddr, logging.PacketType0RTT, p.Size(), logging.PacketDropDOSPrevention).Do(
		func(net.Addr, logging.PacketType, logging.ByteCount, logging.PacketDropReason) { dropped <- struct{}{} },
	)
	server.handlePacket(p)
	select {
	case <-dropped:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
	require.True(t, mockCtrl.Satisfied())

	// There's no cleanup Go routine.
	// Cleanup is triggered when new packets are received.
	// 1. Receive one handshake packet, which triggers the cleanup of the first 0-RTT queue
	mockTracer.EXPECT().DroppedPacket(gomock.Any(), logging.PacketTypeHandshake, gomock.Any(), gomock.Any())
	mockTracer.EXPECT().DroppedPacket(gomock.Any(), logging.PacketType0RTT, sizes[0], gomock.Any()).Do(
		func(net.Addr, logging.PacketType, logging.ByteCount, logging.PacketDropReason) { dropped <- struct{}{} },
	)
	triggerPacket := getLongHeaderPacket(t,
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketTypeHandshake,
				SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
				DestConnectionID: protocol.ParseConnectionID([]byte{8, 7, 6, 5, 4, 3, 2, 1}),
				Length:           123,
				Version:          protocol.Version1,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, 123),
	)
	triggerPacket.rcvTime = firstRcvTime.Add(protocol.Max0RTTQueueingDuration + time.Nanosecond)
	server.handlePacket(triggerPacket)
	select {
	case <-dropped:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
	require.True(t, mockCtrl.Satisfied())

	// 2. Receive another handshake packet, which triggers the cleanup of the other 0-RTT queues
	triggerPacket = getLongHeaderPacket(t,
		&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 42},
		&wire.ExtendedHeader{
			Header: wire.Header{
				Type:             protocol.PacketTypeHandshake,
				SrcConnectionID:  protocol.ParseConnectionID([]byte{5, 4, 3, 2, 1}),
				DestConnectionID: protocol.ParseConnectionID([]byte{8, 7, 6, 5, 4, 3, 2, 1}),
				Length:           124,
				Version:          protocol.Version1,
			},
			PacketNumberLen: protocol.PacketNumberLen4,
		},
		make([]byte, 124),
	)
	triggerPacket.rcvTime = otherRcvTime.Add(protocol.Max0RTTQueueingDuration + time.Nanosecond)
	mockTracer.EXPECT().DroppedPacket(gomock.Any(), logging.PacketTypeHandshake, gomock.Any(), gomock.Any())
	for i := range sizes[1:] {
		mockTracer.EXPECT().DroppedPacket(gomock.Any(), logging.PacketType0RTT, sizes[i+1], gomock.Any()).Do(
			func(net.Addr, logging.PacketType, logging.ByteCount, logging.PacketDropReason) { dropped <- struct{}{} },
		)
	}
	server.handlePacket(triggerPacket)

	for range protocol.Max0RTTQueues - 1 {
		select {
		case <-dropped:
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
}
=======
})

var _ = Describe("default source address verification", func() {
	It("accepts a token", func() {
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1)}
		cookie := &Cookie{
			RemoteAddr: "192.168.0.1",
			SentTime:   time.Now().Add(-protocol.CookieExpiryTime).Add(time.Second), // will expire in 1 second
		}
		Expect(defaultAcceptCookie(remoteAddr, cookie)).To(BeTrue())
	})

	It("requests verification if no token is provided", func() {
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1)}
		Expect(defaultAcceptCookie(remoteAddr, nil)).To(BeFalse())
	})

	It("rejects a token if the address doesn't match", func() {
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1)}
		cookie := &Cookie{
			RemoteAddr: "127.0.0.1",
			SentTime:   time.Now(),
		}
		Expect(defaultAcceptCookie(remoteAddr, cookie)).To(BeFalse())
	})

	It("accepts a token for a remote address is not a UDP address", func() {
		remoteAddr := &net.TCPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 1337}
		cookie := &Cookie{
			RemoteAddr: "192.168.0.1:1337",
			SentTime:   time.Now(),
		}
		Expect(defaultAcceptCookie(remoteAddr, cookie)).To(BeTrue())
	})

	It("rejects an invalid token for a remote address is not a UDP address", func() {
		remoteAddr := &net.TCPAddr{IP: net.IPv4(192, 168, 0, 1), Port: 1337}
		cookie := &Cookie{
			RemoteAddr: "192.168.0.1:7331", // mismatching port
			SentTime:   time.Now(),
		}
		Expect(defaultAcceptCookie(remoteAddr, cookie)).To(BeFalse())
	})

	It("rejects an expired token", func() {
		remoteAddr := &net.UDPAddr{IP: net.IPv4(192, 168, 0, 1)}
		cookie := &Cookie{
			RemoteAddr: "192.168.0.1",
			SentTime:   time.Now().Add(-protocol.CookieExpiryTime).Add(-time.Second), // expired 1 second ago
		}
		Expect(defaultAcceptCookie(remoteAddr, cookie)).To(BeFalse())
	})
})
>>>>>>> project-faster/main
