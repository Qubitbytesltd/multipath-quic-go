package wire

import (
<<<<<<< HEAD
	"io"
	"testing"

	"github.com/quic-go/quic-go/internal/protocol"

	"github.com/stretchr/testify/require"
)

func TestParseConnectionCloseTransportError(t *testing.T) {
	reason := "No recent network activity."
	data := encodeVarInt(0x19)
	data = append(data, encodeVarInt(0x1337)...)              // frame type
	data = append(data, encodeVarInt(uint64(len(reason)))...) // reason phrase length
	data = append(data, []byte(reason)...)
	frame, l, err := parseConnectionCloseFrame(data, connectionCloseFrameType, protocol.Version1)
	require.NoError(t, err)
	require.False(t, frame.IsApplicationError)
	require.EqualValues(t, 0x19, frame.ErrorCode)
	require.EqualValues(t, 0x1337, frame.FrameType)
	require.Equal(t, reason, frame.ReasonPhrase)
	require.Equal(t, len(data), l)
}

func TestParseConnectionCloseWithApplicationError(t *testing.T) {
	reason := "The application messed things up."
	data := encodeVarInt(0xcafe)
	data = append(data, encodeVarInt(uint64(len(reason)))...) // reason phrase length
	data = append(data, reason...)
	frame, l, err := parseConnectionCloseFrame(data, applicationCloseFrameType, protocol.Version1)
	require.NoError(t, err)
	require.True(t, frame.IsApplicationError)
	require.EqualValues(t, 0xcafe, frame.ErrorCode)
	require.Equal(t, reason, frame.ReasonPhrase)
	require.Equal(t, len(data), l)
}

func TestParseConnectionCloseLongReasonPhrase(t *testing.T) {
	data := encodeVarInt(0xcafe)
	data = append(data, encodeVarInt(0x42)...)   // frame type
	data = append(data, encodeVarInt(0xffff)...) // reason phrase length
	_, _, err := parseConnectionCloseFrame(data, connectionCloseFrameType, protocol.Version1)
	require.Equal(t, io.EOF, err)
}

func TestParseConnectionCloseErrorsOnEOFs(t *testing.T) {
	reason := "No recent network activity."
	data := encodeVarInt(0x19)
	data = append(data, encodeVarInt(0x1337)...)              // frame type
	data = append(data, encodeVarInt(uint64(len(reason)))...) // reason phrase length
	data = append(data, []byte(reason)...)
	_, l, err := parseConnectionCloseFrame(data, connectionCloseFrameType, protocol.Version1)
	require.Equal(t, len(data), l)
	require.NoError(t, err)
	for i := range data {
		_, _, err = parseConnectionCloseFrame(data[:i], connectionCloseFrameType, protocol.Version1)
		require.Equal(t, io.EOF, err)
	}
}

func TestParseConnectionCloseNoReasonPhrase(t *testing.T) {
	data := encodeVarInt(0xcafe)
	data = append(data, encodeVarInt(0x42)...) // frame type
	data = append(data, encodeVarInt(0)...)
	frame, l, err := parseConnectionCloseFrame(data, connectionCloseFrameType, protocol.Version1)
	require.NoError(t, err)
	require.Empty(t, frame.ReasonPhrase)
	require.Equal(t, len(data), l)
}

func TestWriteConnectionCloseNoReasonPhrase(t *testing.T) {
	frame := &ConnectionCloseFrame{
		ErrorCode: 0xbeef,
		FrameType: 0x12345,
	}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{connectionCloseFrameType}
	expected = append(expected, encodeVarInt(0xbeef)...)
	expected = append(expected, encodeVarInt(0x12345)...) // frame type
	expected = append(expected, encodeVarInt(0)...)       // reason phrase length
	require.Equal(t, expected, b)
}

func TestWriteConnectionCloseWithReasonPhrase(t *testing.T) {
	frame := &ConnectionCloseFrame{
		ErrorCode:    0xdead,
		ReasonPhrase: "foobar",
	}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{connectionCloseFrameType}
	expected = append(expected, encodeVarInt(0xdead)...)
	expected = append(expected, encodeVarInt(0)...) // frame type
	expected = append(expected, encodeVarInt(6)...) // reason phrase length
	expected = append(expected, []byte("foobar")...)
	require.Equal(t, expected, b)
}

func TestWriteConnectionCloseWithApplicationError(t *testing.T) {
	frame := &ConnectionCloseFrame{
		IsApplicationError: true,
		ErrorCode:          0xdead,
		ReasonPhrase:       "foobar",
	}
	b, err := frame.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{applicationCloseFrameType}
	expected = append(expected, encodeVarInt(0xdead)...)
	expected = append(expected, encodeVarInt(6)...) // reason phrase length
	expected = append(expected, []byte("foobar")...)
	require.Equal(t, expected, b)
}

func TestWriteConnectionCloseTransportError(t *testing.T) {
	f := &ConnectionCloseFrame{
		ErrorCode:    0xcafe,
		FrameType:    0xdeadbeef,
		ReasonPhrase: "foobar",
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	require.Len(t, b, int(f.Length(protocol.Version1)))
}

func TestWriteConnectionCloseLength(t *testing.T) {
	f := &ConnectionCloseFrame{
		IsApplicationError: true,
		ErrorCode:          0xcafe,
		ReasonPhrase:       "foobar",
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	require.Len(t, b, int(f.Length(protocol.Version1)))
}
=======
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/qerr"
)

var _ = Describe("ConnectionCloseFrame", func() {
	Context("when parsing", func() {
		Context("in little endian", func() {
			It("accepts sample frame", func() {
				b := bytes.NewReader([]byte{0x2,
					0x19, 0x0, 0x0, 0x0, // error code
					0x1b, 0x0, // reason phrase length
					'N', 'o', ' ', 'r', 'e', 'c', 'e', 'n', 't', ' ', 'n', 'e', 't', 'w', 'o', 'r', 'k', ' ', 'a', 'c', 't', 'i', 'v', 'i', 't', 'y', '.',
				})
				frame, err := ParseConnectionCloseFrame(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(frame.ErrorCode).To(Equal(qerr.ErrorCode(0x19)))
				Expect(frame.ReasonPhrase).To(Equal("No recent network activity."))
				Expect(b.Len()).To(BeZero())
			})

			It("rejects long reason phrases", func() {
				b := bytes.NewReader([]byte{0x2,
					0xad, 0xfb, 0xca, 0xde, // error code
					0x0, 0xff, // reason phrase length
				})
				_, err := ParseConnectionCloseFrame(b, versionLittleEndian)
				Expect(err).To(MatchError(qerr.Error(qerr.InvalidConnectionCloseData, "reason phrase too long")))
			})

			It("errors on EOFs", func() {
				data := []byte{0x2,
					0x19, 0x0, 0x0, 0x0, // error code
					0x1b, 0x0, // reason phrase length
					'N', 'o', ' ', 'r', 'e', 'c', 'e', 'n', 't', ' ', 'n', 'e', 't', 'w', 'o', 'r', 'k', ' ', 'a', 'c', 't', 'i', 'v', 'i', 't', 'y', '.',
				}
				_, err := ParseConnectionCloseFrame(bytes.NewReader(data), versionLittleEndian)
				Expect(err).NotTo(HaveOccurred())
				for i := range data {
					_, err := ParseConnectionCloseFrame(bytes.NewReader(data[0:i]), versionLittleEndian)
					Expect(err).To(HaveOccurred())
				}
			})
		})

		Context("in big endian", func() {
			It("accepts sample frame", func() {
				b := bytes.NewReader([]byte{0x2,
					0x0, 0x0, 0x0, 0x19, // error code
					0x0, 0x1b, // reason phrase length
					'N', 'o', ' ', 'r', 'e', 'c', 'e', 'n', 't', ' ', 'n', 'e', 't', 'w', 'o', 'r', 'k', ' ', 'a', 'c', 't', 'i', 'v', 'i', 't', 'y', '.',
				})
				frame, err := ParseConnectionCloseFrame(b, versionBigEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(frame.ErrorCode).To(Equal(qerr.ErrorCode(0x19)))
				Expect(frame.ReasonPhrase).To(Equal("No recent network activity."))
				Expect(b.Len()).To(BeZero())
			})

			It("rejects long reason phrases", func() {
				b := bytes.NewReader([]byte{0x2,
					0xad, 0xfb, 0xca, 0xde, // error code
					0xff, 0x0, // reason phrase length
				})
				_, err := ParseConnectionCloseFrame(b, versionBigEndian)
				Expect(err).To(MatchError(qerr.Error(qerr.InvalidConnectionCloseData, "reason phrase too long")))
			})

			It("errors on EOFs", func() {
				data := []byte{0x40,
					0x19, 0x0, 0x0, 0x0, // error code
					0x0, 0x1b, // reason phrase length
					'N', 'o', ' ', 'r', 'e', 'c', 'e', 'n', 't', ' ', 'n', 'e', 't', 'w', 'o', 'r', 'k', ' ', 'a', 'c', 't', 'i', 'v', 'i', 't', 'y', '.',
				}
				_, err := ParseConnectionCloseFrame(bytes.NewReader(data), versionBigEndian)
				Expect(err).NotTo(HaveOccurred())
				for i := range data {
					_, err := ParseConnectionCloseFrame(bytes.NewReader(data[0:i]), versionBigEndian)
					Expect(err).To(HaveOccurred())
				}
			})
		})

		It("parses a frame without a reason phrase", func() {
			b := bytes.NewReader([]byte{0x2,
				0xad, 0xfb, 0xca, 0xde, // error code
				0x0, 0x0, // reason phrase length
			})
			frame, err := ParseConnectionCloseFrame(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(frame.ReasonPhrase).To(BeEmpty())
			Expect(b.Len()).To(BeZero())
		})
	})

	Context("when writing", func() {
		Context("in little endian", func() {
			It("writes a frame without a reason phrase", func() {
				b := &bytes.Buffer{}
				frame := &ConnectionCloseFrame{
					ErrorCode: 0xdeadbeef,
				}
				err := frame.Write(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Len()).To(Equal(1 + 2 + 4))
				Expect(b.Bytes()).To(Equal([]byte{0x2,
					0xef, 0xbe, 0xad, 0xde, // error code
					0x0, 0x0, // reason phrase length
				}))
			})

			It("writes a frame with a reason phrase", func() {
				b := &bytes.Buffer{}
				frame := &ConnectionCloseFrame{
					ErrorCode:    0xdeadbeef,
					ReasonPhrase: "foobar",
				}
				err := frame.Write(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Len()).To(Equal(1 + 2 + 4 + len(frame.ReasonPhrase)))
				Expect(b.Bytes()).To(Equal([]byte{0x2,
					0xef, 0xbe, 0xad, 0xde, // error code
					0x6, 0x0, // reason phrase length
					'f', 'o', 'o', 'b', 'a', 'r',
				}))
			})
		})

		Context("in big endian", func() {
			It("writes a frame without a ReasonPhrase", func() {
				b := &bytes.Buffer{}
				frame := &ConnectionCloseFrame{
					ErrorCode: 0xdeadbeef,
				}
				err := frame.Write(b, versionBigEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Len()).To(Equal(1 + 2 + 4))
				Expect(b.Bytes()).To(Equal([]byte{0x2,
					0xde, 0xad, 0xbe, 0xef, // error code
					0x0, 0x0, // reason phrase length
				}))
			})

			It("writes a frame with a ReasonPhrase", func() {
				b := &bytes.Buffer{}
				frame := &ConnectionCloseFrame{
					ErrorCode:    0xdeadbeef,
					ReasonPhrase: "foobar",
				}
				err := frame.Write(b, versionBigEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Len()).To(Equal(1 + 2 + 4 + len(frame.ReasonPhrase)))
				Expect(b.Bytes()).To(Equal([]byte{0x2,
					0xde, 0xad, 0xbe, 0xef, // error code
					0x0, 0x6, // reason phrase length
					'f', 'o', 'o', 'b', 'a', 'r',
				}))
			})
		})

		It("rejects ReasonPhrases that are too long", func() {
			b := &bytes.Buffer{}
			reasonPhrase := strings.Repeat("a", 0xffff+0x11)
			frame := &ConnectionCloseFrame{
				ErrorCode:    0xdeadbeef,
				ReasonPhrase: reasonPhrase,
			}
			err := frame.Write(b, protocol.VersionWhatever)
			Expect(err).To(HaveOccurred())
		})

		It("has proper min length", func() {
			b := &bytes.Buffer{}
			f := &ConnectionCloseFrame{
				ErrorCode:    0xdeadbeef,
				ReasonPhrase: "foobar",
			}
			err := f.Write(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.MinLength(0)).To(Equal(protocol.ByteCount(b.Len())))
		})
	})

	It("is self-consistent", func() {
		buf := &bytes.Buffer{}
		frame := &ConnectionCloseFrame{
			ErrorCode:    0xdeadbeef,
			ReasonPhrase: "Lorem ipsum dolor sit amet.",
		}
		err := frame.Write(buf, protocol.VersionWhatever)
		Expect(err).ToNot(HaveOccurred())
		b := bytes.NewReader(buf.Bytes())
		readframe, err := ParseConnectionCloseFrame(b, protocol.VersionWhatever)
		Expect(err).ToNot(HaveOccurred())
		Expect(readframe.ErrorCode).To(Equal(frame.ErrorCode))
		Expect(readframe.ReasonPhrase).To(Equal(frame.ReasonPhrase))
		Expect(b.Len()).To(BeZero())
	})
})
>>>>>>> project-faster/main
