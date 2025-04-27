package quic

import (
<<<<<<< HEAD
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/quic-go/quic-go/internal/flowcontrol"
	"github.com/quic-go/quic-go/internal/mocks"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/qerr"
	"github.com/quic-go/quic-go/internal/wire"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	firstIncomingBidiStreamServer protocol.StreamID = 0
	firstOutgoingBidiStreamServer protocol.StreamID = 1
	firstIncomingUniStreamServer  protocol.StreamID = 2
	firstOutgoingUniStreamServer  protocol.StreamID = 3
)

const (
	firstIncomingBidiStreamClient protocol.StreamID = 1
	firstOutgoingBidiStreamClient protocol.StreamID = 0
	firstIncomingUniStreamClient  protocol.StreamID = 3
	firstOutgoingUniStreamClient  protocol.StreamID = 2
)

func (e streamError) TestError() error {
	nums := make([]interface{}, len(e.nums))
	for i, num := range e.nums {
		nums[i] = num
	}
	return fmt.Errorf(e.message, nums...)
}

func TestStreamsMapCreatingStreams(t *testing.T) {
	t.Run("client", func(t *testing.T) {
		testStreamsMapCreatingAndDeletingStreams(t, protocol.PerspectiveClient,
			firstIncomingBidiStreamClient,
			firstOutgoingBidiStreamClient,
			firstIncomingUniStreamClient,
			firstOutgoingUniStreamClient,
		)
	})
	t.Run("server", func(t *testing.T) {
		testStreamsMapCreatingAndDeletingStreams(t, protocol.PerspectiveServer,
			firstIncomingBidiStreamServer,
			firstOutgoingBidiStreamServer,
			firstIncomingUniStreamServer,
			firstOutgoingUniStreamServer,
		)
	})
}

func testStreamsMapCreatingAndDeletingStreams(t *testing.T,
	perspective protocol.Perspective,
	firstIncomingBidiStream protocol.StreamID,
	firstOutgoingBidiStream protocol.StreamID,
	firstIncomingUniStream protocol.StreamID,
	firstOutgoingUniStream protocol.StreamID,
) {
	mockCtrl := gomock.NewController(t)
	mockSender := NewMockStreamSender(mockCtrl)
	m := newStreamsMap(
		context.Background(),
		mockSender,
		func(wire.Frame) {},
		func(protocol.StreamID) flowcontrol.StreamFlowController {
			return mocks.NewMockStreamFlowController(mockCtrl)
		},
		1,
		1,
		perspective,
	)
	m.UpdateLimits(&wire.TransportParameters{
		MaxBidiStreamNum: protocol.MaxStreamCount,
		MaxUniStreamNum:  protocol.MaxStreamCount,
	})

	// opening streams
	str1, err := m.OpenStream()
	require.NoError(t, err)
	str2, err := m.OpenStream()
	require.NoError(t, err)
	ustr1, err := m.OpenUniStream()
	require.NoError(t, err)
	ustr2, err := m.OpenUniStream()
	require.NoError(t, err)

	assert.Equal(t, str1.StreamID(), firstOutgoingBidiStream)
	assert.Equal(t, str2.StreamID(), firstOutgoingBidiStream+4)
	assert.Equal(t, ustr1.StreamID(), firstOutgoingUniStream)
	assert.Equal(t, ustr2.StreamID(), firstOutgoingUniStream+4)

	// accepting streams:
	// This function is called when a frame referencing this stream is received.
	// The peer may open a peer-initiated stream...
	_, err = m.GetOrOpenReceiveStream(firstIncomingBidiStream)
	require.NoError(t, err)
	_, err = m.GetOrOpenReceiveStream(firstIncomingUniStream)
	require.NoError(t, err)

	// ... but not a stream that is initiated by us.
	_, err = m.GetOrOpenSendStream(firstOutgoingBidiStream + 8)
	require.ErrorIs(t, err, &qerr.TransportError{
		ErrorCode:    qerr.StreamStateError,
		ErrorMessage: fmt.Sprintf("peer attempted to open stream %d", firstOutgoingBidiStream+8),
	})
	_, err = m.GetOrOpenSendStream(firstOutgoingUniStream + 8)
	require.ErrorIs(t, err, &qerr.TransportError{
		ErrorCode:    qerr.StreamStateError,
		ErrorMessage: fmt.Sprintf("peer attempted to open stream %d", firstOutgoingUniStream+8),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	str, err := m.AcceptStream(ctx)
	require.NoError(t, err)
	ustr, err := m.AcceptUniStream(ctx)
	require.NoError(t, err)

	assert.Equal(t, str.StreamID(), firstIncomingBidiStream)
	assert.Equal(t, ustr.StreamID(), firstIncomingUniStream)
}

func TestStreamsMapDeletingStreams(t *testing.T) {
	t.Run("client", func(t *testing.T) {
		testStreamsMapDeletingStreams(t, protocol.PerspectiveClient,
			firstIncomingBidiStreamClient,
			firstOutgoingBidiStreamClient,
			firstIncomingUniStreamClient,
			firstOutgoingUniStreamClient,
		)
	})
	t.Run("server", func(t *testing.T) {
		testStreamsMapDeletingStreams(t, protocol.PerspectiveServer,
			firstIncomingBidiStreamServer,
			firstOutgoingBidiStreamServer,
			firstIncomingUniStreamServer,
			firstOutgoingUniStreamServer,
		)
	})
}

func testStreamsMapDeletingStreams(t *testing.T,
	perspective protocol.Perspective,
	firstIncomingBidiStream protocol.StreamID,
	firstOutgoingBidiStream protocol.StreamID,
	firstIncomingUniStream protocol.StreamID,
	firstOutgoingUniStream protocol.StreamID,
) {
	mockCtrl := gomock.NewController(t)
	mockSender := NewMockStreamSender(mockCtrl)
	var frameQueue []wire.Frame
	m := newStreamsMap(
		context.Background(),
		mockSender,
		func(frame wire.Frame) { frameQueue = append(frameQueue, frame) },
		func(protocol.StreamID) flowcontrol.StreamFlowController {
			return mocks.NewMockStreamFlowController(mockCtrl)
		},
		100,
		100,
		perspective,
	)
	m.UpdateLimits(&wire.TransportParameters{
		MaxBidiStreamNum: 10,
		MaxUniStreamNum:  10,
	})

	_, err := m.OpenStream()
	require.NoError(t, err)
	require.NoError(t, m.DeleteStream(firstOutgoingBidiStream))
	sstr, err := m.GetOrOpenSendStream(firstOutgoingBidiStream)
	require.NoError(t, err)
	require.Nil(t, sstr)
	require.ErrorContains(t,
		m.DeleteStream(firstOutgoingBidiStream+400),
		fmt.Sprintf("tried to delete unknown outgoing stream %d", firstOutgoingBidiStream+400),
	)

	_, err = m.OpenUniStream()
	require.NoError(t, err)
	require.NoError(t, m.DeleteStream(firstOutgoingUniStream))
	sstr, err = m.GetOrOpenSendStream(firstOutgoingUniStream)
	require.NoError(t, err)
	require.Nil(t, sstr)
	require.ErrorContains(t,
		m.DeleteStream(firstOutgoingUniStream+400),
		fmt.Sprintf("tried to delete unknown outgoing stream %d", firstOutgoingUniStream+400),
	)

	require.Empty(t, frameQueue)
	// deleting incoming bidirectional streams
	_, err = m.GetOrOpenReceiveStream(firstIncomingBidiStream)
	require.NoError(t, err)
	require.NoError(t, m.DeleteStream(firstIncomingBidiStream))
	sstr, err = m.GetOrOpenSendStream(firstIncomingBidiStream)
	require.NoError(t, err)
	require.Nil(t, sstr)
	require.ErrorContains(t,
		m.DeleteStream(firstIncomingBidiStream+400),
		fmt.Sprintf("tried to delete unknown incoming stream %d", firstIncomingBidiStream+400),
	)
	// the MAX_STREAMS frame is only queued once the stream is accepted
	require.Empty(t, frameQueue)
	_, err = m.AcceptStream(context.Background())
	require.NoError(t, err)

	require.Equal(t, frameQueue, []wire.Frame{
		&wire.MaxStreamsFrame{
			Type:         protocol.StreamTypeBidi,
			MaxStreamNum: 101,
		},
	})
	frameQueue = frameQueue[:0]

	// deleting incoming unidirectional streams
	_, err = m.GetOrOpenReceiveStream(firstIncomingUniStream)
	require.NoError(t, err)
	require.NoError(t, m.DeleteStream(firstIncomingUniStream))
	rstr, err := m.GetOrOpenReceiveStream(firstIncomingUniStream)
	require.NoError(t, err)
	require.Nil(t, rstr)
	require.ErrorContains(t,
		m.DeleteStream(firstIncomingUniStream+400),
		fmt.Sprintf("tried to delete unknown incoming stream %d", firstIncomingUniStream+400),
	)
	// the MAX_STREAMS frame is only queued once the stream is accepted
	require.Empty(t, frameQueue)
	_, err = m.AcceptUniStream(context.Background())
	require.NoError(t, err)

	require.Equal(t, frameQueue, []wire.Frame{
		&wire.MaxStreamsFrame{
			Type:         protocol.StreamTypeUni,
			MaxStreamNum: 101,
		},
=======
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/project-faster/mp-quic-go/internal/mocks"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/qerr"
)

var _ = Describe("Streams Map", func() {
	const (
		maxIncomingStreams = 75
		maxOutgoingStreams = 60
	)

	var (
		m       *streamsMap
		mockCpm *mocks.MockConnectionParametersManager
	)

	setNewStreamsMap := func(p protocol.Perspective) {
		mockCpm = mocks.NewMockConnectionParametersManager(mockCtrl)

		mockCpm.EXPECT().GetMaxOutgoingStreams().AnyTimes().Return(uint32(maxOutgoingStreams))
		mockCpm.EXPECT().GetMaxIncomingStreams().AnyTimes().Return(uint32(maxIncomingStreams))

		m = newStreamsMap(nil, p, mockCpm)
		m.newStream = func(id protocol.StreamID) *stream {
			return newStream(id, nil, nil, nil)
		}
	}

	AfterEach(func() {
		Expect(m.openStreams).To(HaveLen(len(m.streams)))
>>>>>>> project-faster/main
	})
	frameQueue = frameQueue[:0]
}

<<<<<<< HEAD
func TestStreamsMapStreamLimits(t *testing.T) {
	t.Run("client", func(t *testing.T) {
		testStreamsMapStreamLimits(t, protocol.PerspectiveClient)
=======
	Context("getting and creating streams", func() {
		Context("as a server", func() {
			BeforeEach(func() {
				setNewStreamsMap(protocol.PerspectiveServer)
			})

			Context("client-side streams", func() {
				It("gets new streams", func() {
					s, err := m.GetOrOpenStream(1)
					Expect(err).NotTo(HaveOccurred())
					Expect(s.StreamID()).To(Equal(protocol.StreamID(1)))
					Expect(m.numIncomingStreams).To(BeEquivalentTo(1))
					Expect(m.numOutgoingStreams).To(BeZero())
				})

				It("rejects streams with even IDs", func() {
					_, err := m.GetOrOpenStream(6)
					Expect(err).To(MatchError("InvalidStreamID: attempted to open stream 6 from client-side"))
				})

				It("rejects streams with even IDs, which are lower thatn the highest client-side stream", func() {
					_, err := m.GetOrOpenStream(5)
					Expect(err).NotTo(HaveOccurred())
					_, err = m.GetOrOpenStream(4)
					Expect(err).To(MatchError("InvalidStreamID: attempted to open stream 4 from client-side"))
				})

				It("gets existing streams", func() {
					s, err := m.GetOrOpenStream(5)
					Expect(err).NotTo(HaveOccurred())
					numStreams := m.numIncomingStreams
					s, err = m.GetOrOpenStream(5)
					Expect(err).NotTo(HaveOccurred())
					Expect(s.StreamID()).To(Equal(protocol.StreamID(5)))
					Expect(m.numIncomingStreams).To(Equal(numStreams))
				})

				It("returns nil for closed streams", func() {
					_, err := m.GetOrOpenStream(5)
					Expect(err).NotTo(HaveOccurred())
					err = m.RemoveStream(5)
					Expect(err).NotTo(HaveOccurred())
					s, err := m.GetOrOpenStream(5)
					Expect(err).NotTo(HaveOccurred())
					Expect(s).To(BeNil())
				})

				It("opens skipped streams", func() {
					_, err := m.GetOrOpenStream(5)
					Expect(err).NotTo(HaveOccurred())
					Expect(m.streams).To(HaveKey(protocol.StreamID(1)))
					Expect(m.streams).To(HaveKey(protocol.StreamID(3)))
					Expect(m.streams).To(HaveKey(protocol.StreamID(5)))
				})

				It("doesn't reopen an already closed stream", func() {
					_, err := m.GetOrOpenStream(5)
					Expect(err).ToNot(HaveOccurred())
					err = m.RemoveStream(5)
					Expect(err).ToNot(HaveOccurred())
					str, err := m.GetOrOpenStream(5)
					Expect(err).ToNot(HaveOccurred())
					Expect(str).To(BeNil())
				})

				Context("counting streams", func() {
					It("errors when too many streams are opened", func() {
						for i := 0; i < maxIncomingStreams; i++ {
							_, err := m.GetOrOpenStream(protocol.StreamID(i*2 + 1))
							Expect(err).NotTo(HaveOccurred())
						}
						_, err := m.GetOrOpenStream(protocol.StreamID(2*maxIncomingStreams + 3))
						Expect(err).To(MatchError(qerr.TooManyOpenStreams))
					})

					It("errors when too many streams are opened implicitely", func() {
						_, err := m.GetOrOpenStream(protocol.StreamID(maxIncomingStreams*2 + 1))
						Expect(err).To(MatchError(qerr.TooManyOpenStreams))
					})

					It("does not error when many streams are opened and closed", func() {
						for i := 2; i < 10*maxIncomingStreams; i++ {
							_, err := m.GetOrOpenStream(protocol.StreamID(i*2 + 1))
							Expect(err).NotTo(HaveOccurred())
							m.RemoveStream(protocol.StreamID(i*2 + 1))
						}
					})
				})
			})

			Context("server-side streams", func() {
				It("opens a stream 2 first", func() {
					s, err := m.OpenStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(s).ToNot(BeNil())
					Expect(s.StreamID()).To(Equal(protocol.StreamID(2)))
					Expect(m.numIncomingStreams).To(BeZero())
					Expect(m.numOutgoingStreams).To(BeEquivalentTo(1))
				})

				It("returns the error when the streamsMap was closed", func() {
					testErr := errors.New("test error")
					m.CloseWithError(testErr)
					_, err := m.OpenStream()
					Expect(err).To(MatchError(testErr))
				})

				It("doesn't reopen an already closed stream", func() {
					str, err := m.OpenStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(2)))
					err = m.RemoveStream(2)
					Expect(err).ToNot(HaveOccurred())
					str, err = m.GetOrOpenStream(2)
					Expect(err).ToNot(HaveOccurred())
					Expect(str).To(BeNil())
				})

				Context("counting streams", func() {
					It("errors when too many streams are opened", func() {
						for i := 1; i <= maxOutgoingStreams; i++ {
							_, err := m.OpenStream()
							Expect(err).NotTo(HaveOccurred())
						}
						_, err := m.OpenStream()
						Expect(err).To(MatchError(qerr.TooManyOpenStreams))
					})

					It("does not error when many streams are opened and closed", func() {
						for i := 2; i < 10*maxOutgoingStreams; i++ {
							str, err := m.OpenStream()
							Expect(err).NotTo(HaveOccurred())
							m.RemoveStream(str.StreamID())
						}
					})

					It("allows many server- and client-side streams at the same time", func() {
						for i := 1; i < maxOutgoingStreams; i++ {
							_, err := m.OpenStream()
							Expect(err).ToNot(HaveOccurred())
						}
						for i := 0; i < maxOutgoingStreams; i++ {
							_, err := m.GetOrOpenStream(protocol.StreamID(2*i + 1))
							Expect(err).ToNot(HaveOccurred())
						}
					})
				})

				Context("opening streams synchronously", func() {
					openMaxNumStreams := func() {
						for i := 1; i <= maxOutgoingStreams; i++ {
							_, err := m.OpenStream()
							Expect(err).NotTo(HaveOccurred())
						}
						_, err := m.OpenStream()
						Expect(err).To(MatchError(qerr.TooManyOpenStreams))
					}

					It("waits until another stream is closed", func() {
						openMaxNumStreams()
						var returned bool
						var str *stream
						go func() {
							defer GinkgoRecover()
							var err error
							str, err = m.OpenStreamSync()
							Expect(err).ToNot(HaveOccurred())
							returned = true
						}()

						Consistently(func() bool { return returned }).Should(BeFalse())
						err := m.RemoveStream(6)
						Expect(err).ToNot(HaveOccurred())
						Eventually(func() bool { return returned }).Should(BeTrue())
						Expect(str.StreamID()).To(Equal(protocol.StreamID(2*maxOutgoingStreams + 2)))
					})

					It("stops waiting when an error is registered", func() {
						openMaxNumStreams()
						testErr := errors.New("test error")
						var err error
						var returned bool
						go func() {
							_, err = m.OpenStreamSync()
							returned = true
						}()

						Consistently(func() bool { return returned }).Should(BeFalse())
						m.CloseWithError(testErr)
						Eventually(func() bool { return returned }).Should(BeTrue())
						Expect(err).To(MatchError(testErr))
					})

					It("immediately returns when OpenStreamSync is called after an error was registered", func() {
						testErr := errors.New("test error")
						m.CloseWithError(testErr)
						_, err := m.OpenStreamSync()
						Expect(err).To(MatchError(testErr))
					})
				})
			})

			Context("accepting streams", func() {
				It("does nothing if no stream is opened", func() {
					var accepted bool
					go func() {
						_, _ = m.AcceptStream()
						accepted = true
					}()
					Consistently(func() bool { return accepted }).Should(BeFalse())
				})

				It("accepts stream 1 first", func() {
					var str *stream
					go func() {
						defer GinkgoRecover()
						var err error
						str, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					_, err := m.GetOrOpenStream(1)
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() Stream { return str }).ShouldNot(BeNil())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(1)))
				})

				It("returns an implicitly opened stream, if a stream number is skipped", func() {
					var str *stream
					go func() {
						defer GinkgoRecover()
						var err error
						str, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					_, err := m.GetOrOpenStream(5)
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() Stream { return str }).ShouldNot(BeNil())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(1)))
				})

				It("returns to multiple accepts", func() {
					var str1, str2 *stream
					go func() {
						defer GinkgoRecover()
						var err error
						str1, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					go func() {
						defer GinkgoRecover()
						var err error
						str2, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					_, err := m.GetOrOpenStream(3) // opens stream 1 and 3
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() *stream { return str1 }).ShouldNot(BeNil())
					Eventually(func() *stream { return str2 }).ShouldNot(BeNil())
					Expect(str1.StreamID()).ToNot(Equal(str2.StreamID()))
					Expect(str1.StreamID() + str2.StreamID()).To(BeEquivalentTo(1 + 3))
				})

				It("waits a new stream is available", func() {
					var str *stream
					go func() {
						defer GinkgoRecover()
						var err error
						str, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					Consistently(func() *stream { return str }).Should(BeNil())
					_, err := m.GetOrOpenStream(1)
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() *stream { return str }).ShouldNot(BeNil())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(1)))
				})

				It("returns multiple streams on subsequent Accept calls, if available", func() {
					var str *stream
					go func() {
						defer GinkgoRecover()
						var err error
						str, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					_, err := m.GetOrOpenStream(3)
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() *stream { return str }).ShouldNot(BeNil())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(1)))
					str, err = m.AcceptStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(3)))
				})

				It("blocks after accepting a stream", func() {
					var accepted bool
					_, err := m.GetOrOpenStream(1)
					Expect(err).ToNot(HaveOccurred())
					str, err := m.AcceptStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(1)))
					go func() {
						defer GinkgoRecover()
						_, _ = m.AcceptStream()
						accepted = true
					}()
					Consistently(func() bool { return accepted }).Should(BeFalse())
				})

				It("stops waiting when an error is registered", func() {
					testErr := errors.New("testErr")
					var acceptErr error
					go func() {
						_, acceptErr = m.AcceptStream()
					}()
					Consistently(func() error { return acceptErr }).ShouldNot(HaveOccurred())
					m.CloseWithError(testErr)
					Eventually(func() error { return acceptErr }).Should(MatchError(testErr))
				})

				It("immediately returns when Accept is called after an error was registered", func() {
					testErr := errors.New("testErr")
					m.CloseWithError(testErr)
					_, err := m.AcceptStream()
					Expect(err).To(MatchError(testErr))
				})
			})
		})

		Context("as a client", func() {
			BeforeEach(func() {
				setNewStreamsMap(protocol.PerspectiveClient)
			})

			Context("client-side streams", func() {
				It("rejects streams with odd IDs", func() {
					_, err := m.GetOrOpenStream(5)
					Expect(err).To(MatchError("InvalidStreamID: attempted to open stream 5 from server-side"))
				})

				It("rejects streams with odds IDs, which are lower thatn the highest server-side stream", func() {
					_, err := m.GetOrOpenStream(6)
					Expect(err).NotTo(HaveOccurred())
					_, err = m.GetOrOpenStream(5)
					Expect(err).To(MatchError("InvalidStreamID: attempted to open stream 5 from server-side"))
				})

				It("gets new streams", func() {
					s, err := m.GetOrOpenStream(2)
					Expect(err).NotTo(HaveOccurred())
					Expect(s.StreamID()).To(Equal(protocol.StreamID(2)))
					Expect(m.numOutgoingStreams).To(BeEquivalentTo(1))
					Expect(m.numIncomingStreams).To(BeZero())
				})

				It("opens skipped streams", func() {
					_, err := m.GetOrOpenStream(6)
					Expect(err).NotTo(HaveOccurred())
					Expect(m.streams).To(HaveKey(protocol.StreamID(2)))
					Expect(m.streams).To(HaveKey(protocol.StreamID(4)))
					Expect(m.streams).To(HaveKey(protocol.StreamID(6)))
				})

				It("doesn't reopen an already closed stream", func() {
					str, err := m.OpenStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(1)))
					err = m.RemoveStream(1)
					Expect(err).ToNot(HaveOccurred())
					str, err = m.GetOrOpenStream(1)
					Expect(err).ToNot(HaveOccurred())
					Expect(str).To(BeNil())
				})
			})

			Context("server-side streams", func() {
				It("opens stream 1 first", func() {
					s, err := m.OpenStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(s).ToNot(BeNil())
					Expect(s.StreamID()).To(BeEquivalentTo(1))
					Expect(m.numOutgoingStreams).To(BeZero())
					Expect(m.numIncomingStreams).To(BeEquivalentTo(1))
				})

				It("opens multiple streams", func() {
					s1, err := m.OpenStream()
					Expect(err).ToNot(HaveOccurred())
					s2, err := m.OpenStream()
					Expect(err).ToNot(HaveOccurred())
					Expect(s2.StreamID()).To(Equal(s1.StreamID() + 2))
				})

				It("doesn't reopen an already closed stream", func() {
					_, err := m.GetOrOpenStream(4)
					Expect(err).ToNot(HaveOccurred())
					err = m.RemoveStream(4)
					Expect(err).ToNot(HaveOccurred())
					str, err := m.GetOrOpenStream(4)
					Expect(err).ToNot(HaveOccurred())
					Expect(str).To(BeNil())
				})
			})

			Context("accepting streams", func() {
				It("accepts stream 2 first", func() {
					var str *stream
					go func() {
						defer GinkgoRecover()
						var err error
						str, err = m.AcceptStream()
						Expect(err).ToNot(HaveOccurred())
					}()
					_, err := m.GetOrOpenStream(2)
					Expect(err).ToNot(HaveOccurred())
					Eventually(func() *stream { return str }).ShouldNot(BeNil())
					Expect(str.StreamID()).To(Equal(protocol.StreamID(2)))
				})
			})
		})
>>>>>>> project-faster/main
	})
	t.Run("server", func(t *testing.T) {
		testStreamsMapStreamLimits(t, protocol.PerspectiveServer)
	})
}

func testStreamsMapStreamLimits(t *testing.T, perspective protocol.Perspective) {
	mockCtrl := gomock.NewController(t)
	mockSender := NewMockStreamSender(mockCtrl)
	var frameQueue []wire.Frame
	m := newStreamsMap(
		context.Background(),
		mockSender,
		func(frame wire.Frame) { frameQueue = append(frameQueue, frame) },
		func(protocol.StreamID) flowcontrol.StreamFlowController {
			fc := mocks.NewMockStreamFlowController(mockCtrl)
			fc.EXPECT().UpdateSendWindow(gomock.Any()).AnyTimes()
			return fc
		},
		100,
		100,
		perspective,
	)

	// increase via transport parameters
	_, err := m.OpenStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})
	m.UpdateLimits(&wire.TransportParameters{MaxBidiStreamNum: 1})
	_, err = m.OpenStream()
	require.NoError(t, err)
	_, err = m.OpenStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})

	_, err = m.OpenUniStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})
	m.UpdateLimits(&wire.TransportParameters{MaxUniStreamNum: 1})
	_, err = m.OpenUniStream()
	require.NoError(t, err)
	_, err = m.OpenUniStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})

	// increase via MAX_STREAMS frames
	m.HandleMaxStreamsFrame(&wire.MaxStreamsFrame{
		Type:         protocol.StreamTypeBidi,
		MaxStreamNum: 2,
	})
	_, err = m.OpenStream()
	require.NoError(t, err)
	_, err = m.OpenStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})

	m.HandleMaxStreamsFrame(&wire.MaxStreamsFrame{
		Type:         protocol.StreamTypeUni,
		MaxStreamNum: 2,
	})
	_, err = m.OpenUniStream()
	require.NoError(t, err)
	_, err = m.OpenUniStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})

	// decrease via transport parameters
	m.UpdateLimits(&wire.TransportParameters{MaxBidiStreamNum: 0})
	_, err = m.OpenStream()
	require.ErrorIs(t, err, &StreamLimitReachedError{})
}

func TestStreamsMapClosing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockSender := NewMockStreamSender(mockCtrl)
	m := newStreamsMap(
		context.Background(),
		mockSender,
		func(wire.Frame) {},
		func(protocol.StreamID) flowcontrol.StreamFlowController {
			return mocks.NewMockStreamFlowController(mockCtrl)
		},
		1,
		1,
		protocol.PerspectiveClient,
	)
	m.CloseWithError(assert.AnError)
	_, err := m.OpenStream()
	require.ErrorIs(t, err, assert.AnError)
	_, err = m.OpenUniStream()
	require.ErrorIs(t, err, assert.AnError)
	_, err = m.AcceptStream(context.Background())
	require.ErrorIs(t, err, assert.AnError)
	_, err = m.AcceptUniStream(context.Background())
	require.ErrorIs(t, err, assert.AnError)
}

func TestStreamsMap0RTT(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockSender := NewMockStreamSender(mockCtrl)
	fcBidi := mocks.NewMockStreamFlowController(mockCtrl)
	fcUni := mocks.NewMockStreamFlowController(mockCtrl)
	fcs := []flowcontrol.StreamFlowController{fcBidi, fcUni}
	m := newStreamsMap(
		context.Background(),
		mockSender,
		func(wire.Frame) {},
		func(protocol.StreamID) flowcontrol.StreamFlowController {
			fc := fcs[0]
			fcs = fcs[1:]
			return fc
		},
		1,
		1,
		protocol.PerspectiveClient,
	)
	// restored transport parameters
	m.UpdateLimits(&wire.TransportParameters{
		MaxBidiStreamNum: 1,
		MaxUniStreamNum:  1,
	})
	_, err := m.OpenStream()
	require.NoError(t, err)
	_, err = m.OpenUniStream()
	require.NoError(t, err)

	fcBidi.EXPECT().UpdateSendWindow(protocol.ByteCount(1234))
	fcUni.EXPECT().UpdateSendWindow(protocol.ByteCount(4321))
	// new transport parameters
	m.UpdateLimits(&wire.TransportParameters{
		MaxBidiStreamNum:               1000,
		InitialMaxStreamDataBidiRemote: 1234,
		MaxUniStreamNum:                1000,
		InitialMaxStreamDataUni:        4321,
	})
}

func TestStreamsMap0RTTRejection(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockSender := NewMockStreamSender(mockCtrl)
	m := newStreamsMap(
		context.Background(),
		mockSender,
		func(wire.Frame) {},
		func(protocol.StreamID) flowcontrol.StreamFlowController {
			return mocks.NewMockStreamFlowController(mockCtrl)
		},
		1,
		1,
		protocol.PerspectiveClient,
	)

	m.ResetFor0RTT()
	_, err := m.OpenStream()
	require.ErrorIs(t, err, Err0RTTRejected)
	_, err = m.OpenUniStream()
	require.ErrorIs(t, err, Err0RTTRejected)
	_, err = m.AcceptStream(context.Background())
	require.ErrorIs(t, err, Err0RTTRejected)
	// make sure that we can still get new streams, as the server might be sending us data
	str, err := m.GetOrOpenReceiveStream(3)
	require.NoError(t, err)
	require.NotNil(t, str)

	// now switch to using the new streams map
	m.UseResetMaps()
	_, err = m.OpenStream()
	require.Error(t, err)
	require.ErrorIs(t, err, &StreamLimitReachedError{})
}
