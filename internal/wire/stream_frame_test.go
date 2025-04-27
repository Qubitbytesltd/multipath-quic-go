package wire

import (
	"bytes"
<<<<<<< HEAD
	"io"
	"testing"

	"github.com/quic-go/quic-go/internal/protocol"

	"github.com/stretchr/testify/require"
)

func TestParseStreamFrameWithOffBit(t *testing.T) {
	data := encodeVarInt(0x12345)                    // stream ID
	data = append(data, encodeVarInt(0xdecafbad)...) // offset
	data = append(data, []byte("foobar")...)
	frame, l, err := parseStreamFrame(data, 0x8^0x4, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.StreamID(0x12345), frame.StreamID)
	require.Equal(t, []byte("foobar"), frame.Data)
	require.False(t, frame.Fin)
	require.Equal(t, protocol.ByteCount(0xdecafbad), frame.Offset)
	require.Equal(t, len(data), l)
}

func TestParseStreamFrameRespectsLEN(t *testing.T) {
	data := encodeVarInt(0x12345)           // stream ID
	data = append(data, encodeVarInt(4)...) // data length
	data = append(data, []byte("foobar")...)
	frame, l, err := parseStreamFrame(data, 0x8^0x2, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.StreamID(0x12345), frame.StreamID)
	require.Equal(t, []byte("foob"), frame.Data)
	require.False(t, frame.Fin)
	require.Zero(t, frame.Offset)
	require.Equal(t, len(data)-2, l)
}

func TestParseStreamFrameWithFINBit(t *testing.T) {
	data := encodeVarInt(9) // stream ID
	data = append(data, []byte("foobar")...)
	frame, l, err := parseStreamFrame(data, 0x8^0x1, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.StreamID(9), frame.StreamID)
	require.Equal(t, []byte("foobar"), frame.Data)
	require.True(t, frame.Fin)
	require.Zero(t, frame.Offset)
	require.Equal(t, len(data), l)
}

func TestParseStreamFrameAllowsEmpty(t *testing.T) {
	data := encodeVarInt(0x1337)                  // stream ID
	data = append(data, encodeVarInt(0x12345)...) // offset
	f, l, err := parseStreamFrame(data, 0x8^0x4, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.StreamID(0x1337), f.StreamID)
	require.Equal(t, protocol.ByteCount(0x12345), f.Offset)
	require.Nil(t, f.Data)
	require.False(t, f.Fin)
	require.Equal(t, len(data), l)
}

func TestParseStreamFrameRejectsOverflow(t *testing.T) {
	data := encodeVarInt(0x12345)                                         // stream ID
	data = append(data, encodeVarInt(uint64(protocol.MaxByteCount-5))...) // offset
	data = append(data, []byte("foobar")...)
	_, _, err := parseStreamFrame(data, 0x8^0x4, protocol.Version1)
	require.EqualError(t, err, "stream data overflows maximum offset")
}

func TestParseStreamFrameRejectsLongFrames(t *testing.T) {
	data := encodeVarInt(0x12345)                                                // stream ID
	data = append(data, encodeVarInt(uint64(protocol.MaxPacketBufferSize)+1)...) // data length
	data = append(data, make([]byte, protocol.MaxPacketBufferSize+1)...)
	_, _, err := parseStreamFrame(data, 0x8^0x2, protocol.Version1)
	require.Equal(t, io.EOF, err)
}

func TestParseStreamFrameRejectsFramesExceedingRemainingSize(t *testing.T) {
	data := encodeVarInt(0x12345)           // stream ID
	data = append(data, encodeVarInt(7)...) // data length
	data = append(data, []byte("foobar")...)
	_, _, err := parseStreamFrame(data, 0x8^0x2, protocol.Version1)
	require.Equal(t, io.EOF, err)
}

func TestParseStreamFrameErrorsOnEOFs(t *testing.T) {
	typ := uint64(0x8 ^ 0x4 ^ 0x2)
	data := encodeVarInt(0x12345)                    // stream ID
	data = append(data, encodeVarInt(0xdecafbad)...) // offset
	data = append(data, encodeVarInt(6)...)          // data length
	data = append(data, []byte("foobar")...)
	_, _, err := parseStreamFrame(data, typ, protocol.Version1)
	require.NoError(t, err)
	for i := range data {
		_, _, err = parseStreamFrame(data[:i], typ, protocol.Version1)
		require.Error(t, err)
	}
}

func TestParseStreamUsesBufferForLongFrames(t *testing.T) {
	data := encodeVarInt(0x12345) // stream ID
	data = append(data, bytes.Repeat([]byte{'f'}, protocol.MinStreamFrameBufferSize)...)
	frame, l, err := parseStreamFrame(data, 0x8, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.StreamID(0x12345), frame.StreamID)
	require.Equal(t, bytes.Repeat([]byte{'f'}, protocol.MinStreamFrameBufferSize), frame.Data)
	require.Equal(t, protocol.ByteCount(protocol.MinStreamFrameBufferSize), frame.DataLen())
	require.False(t, frame.Fin)
	require.True(t, frame.fromPool)
	require.Equal(t, len(data), l)
	require.NotPanics(t, frame.PutBack)
}

func TestParseStreamDoesNotUseBufferForShortFrames(t *testing.T) {
	data := encodeVarInt(0x12345) // stream ID
	data = append(data, bytes.Repeat([]byte{'f'}, protocol.MinStreamFrameBufferSize-1)...)
	frame, l, err := parseStreamFrame(data, 0x8, protocol.Version1)
	require.NoError(t, err)
	require.Equal(t, protocol.StreamID(0x12345), frame.StreamID)
	require.Equal(t, bytes.Repeat([]byte{'f'}, protocol.MinStreamFrameBufferSize-1), frame.Data)
	require.Equal(t, protocol.ByteCount(protocol.MinStreamFrameBufferSize-1), frame.DataLen())
	require.False(t, frame.Fin)
	require.False(t, frame.fromPool)
	require.Equal(t, len(data), l)
	require.NotPanics(t, frame.PutBack)
}

func TestWriteStreamFrameWithoutOffset(t *testing.T) {
	f := &StreamFrame{
		StreamID: 0x1337,
		Data:     []byte("foobar"),
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{0x8}
	expected = append(expected, encodeVarInt(0x1337)...) // stream ID
	expected = append(expected, []byte("foobar")...)
	require.Equal(t, expected, b)
	require.Equal(t, int(f.Length(protocol.Version1)), len(b))
}

func TestWriteStreamFrameWithOffset(t *testing.T) {
	f := &StreamFrame{
		StreamID: 0x1337,
		Offset:   0x123456,
		Data:     []byte("foobar"),
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{0x8 ^ 0x4}
	expected = append(expected, encodeVarInt(0x1337)...)   // stream ID
	expected = append(expected, encodeVarInt(0x123456)...) // offset
	expected = append(expected, []byte("foobar")...)
	require.Equal(t, expected, b)
	require.Equal(t, int(f.Length(protocol.Version1)), len(b))
}

func TestWriteStreamFrameWithFIN(t *testing.T) {
	f := &StreamFrame{
		StreamID: 0x1337,
		Offset:   0x123456,
		Fin:      true,
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{0x8 ^ 0x4 ^ 0x1}
	expected = append(expected, encodeVarInt(0x1337)...)   // stream ID
	expected = append(expected, encodeVarInt(0x123456)...) // offset
	require.Equal(t, expected, b)
	require.Equal(t, int(f.Length(protocol.Version1)), len(b))
}

func TestWriteStreamFrameWithDataLength(t *testing.T) {
	f := &StreamFrame{
		StreamID:       0x1337,
		Data:           []byte("foobar"),
		DataLenPresent: true,
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{0x8 ^ 0x2}
	expected = append(expected, encodeVarInt(0x1337)...) // stream ID
	expected = append(expected, encodeVarInt(6)...)      // data length
	expected = append(expected, []byte("foobar")...)
	require.Equal(t, expected, b)
	require.Equal(t, int(f.Length(protocol.Version1)), len(b))
}

func TestWriteStreamFrameWithDataLengthAndOffset(t *testing.T) {
	f := &StreamFrame{
		StreamID:       0x1337,
		Data:           []byte("foobar"),
		DataLenPresent: true,
		Offset:         0x123456,
	}
	b, err := f.Append(nil, protocol.Version1)
	require.NoError(t, err)
	expected := []byte{0x8 ^ 0x4 ^ 0x2}
	expected = append(expected, encodeVarInt(0x1337)...)   // stream ID
	expected = append(expected, encodeVarInt(0x123456)...) // offset
	expected = append(expected, encodeVarInt(6)...)        // data length
	expected = append(expected, []byte("foobar")...)
	require.Equal(t, expected, b)
	require.Equal(t, int(f.Length(protocol.Version1)), len(b))
}

func TestWriteStreamFrameEmptyFrameWithoutFIN(t *testing.T) {
	f := &StreamFrame{
		StreamID: 0x42,
		Offset:   0x1337,
	}
	_, err := f.Append(nil, protocol.Version1)
	require.EqualError(t, err, "StreamFrame: attempting to write empty frame without FIN")
}

func TestStreamMaxDataLength(t *testing.T) {
	const maxSize = 3000
	data := make([]byte, maxSize)
	f := &StreamFrame{
		StreamID: 0x1337,
		Offset:   0xdeadbeef,
	}
	for i := 1; i < 3000; i++ {
		f.Data = nil
		maxDataLen := f.MaxDataLen(protocol.ByteCount(i), protocol.Version1)
		if maxDataLen == 0 { // 0 means that no valid STREAM frame can be written
			// check that writing a minimal size STREAM frame (i.e. with 1 byte data) is actually larger than the desired size
			f.Data = []byte{0}
			b, err := f.Append(nil, protocol.Version1)
			require.NoError(t, err)
			require.Greater(t, len(b), i)
			continue
		}
		f.Data = data[:int(maxDataLen)]
		b, err := f.Append(nil, protocol.Version1)
		require.NoError(t, err)
		require.Equal(t, i, len(b))
	}
}

func TestStreamMaxDataLengthWithDataLenPresent(t *testing.T) {
	const maxSize = 3000
	data := make([]byte, maxSize)
	f := &StreamFrame{
		StreamID:       0x1337,
		Offset:         0xdeadbeef,
		DataLenPresent: true,
	}
	var frameOneByteTooSmallCounter int
	for i := 1; i < 3000; i++ {
		f.Data = nil
		maxDataLen := f.MaxDataLen(protocol.ByteCount(i), protocol.Version1)
		if maxDataLen == 0 { // 0 means that no valid STREAM frame can be written
			// check that writing a minimal size STREAM frame (i.e. with 1 byte data) is actually larger than the desired size
			f.Data = []byte{0}
			b, err := f.Append(nil, protocol.Version1)
			require.NoError(t, err)
			require.Greater(t, len(b), i)
			continue
		}
		f.Data = data[:int(maxDataLen)]
		b, err := f.Append(nil, protocol.Version1)
		require.NoError(t, err)
		// There's *one* pathological case, where a data length of x can be encoded into 1 byte
		// but a data lengths of x+1 needs 2 bytes
		// In that case, it's impossible to create a STREAM frame of the desired size
		if len(b) == i-1 {
			frameOneByteTooSmallCounter++
			continue
		}
		require.Equal(t, i, len(b))
	}
	require.Equal(t, 1, frameOneByteTooSmallCounter)
}

func TestStreamSplitting(t *testing.T) {
	f := &StreamFrame{
		StreamID:       0x1337,
		DataLenPresent: true,
		Offset:         0x100,
		Data:           []byte("foobar"),
	}
	frame, needsSplit := f.MaybeSplitOffFrame(f.Length(protocol.Version1)-3, protocol.Version1)
	require.True(t, needsSplit)
	require.NotNil(t, frame)
	require.True(t, f.DataLenPresent)
	require.True(t, frame.DataLenPresent)
	require.Equal(t, protocol.ByteCount(0x100), frame.Offset)
	require.Equal(t, []byte("foo"), frame.Data)
	require.Equal(t, protocol.ByteCount(0x100+3), f.Offset)
	require.Equal(t, []byte("bar"), f.Data)
}

func TestStreamSplittingNoSplitForShortFrame(t *testing.T) {
	f := &StreamFrame{
		StreamID:       0x1337,
		DataLenPresent: true,
		Offset:         0xdeadbeef,
		Data:           make([]byte, 100),
	}
	frame, needsSplit := f.MaybeSplitOffFrame(f.Length(protocol.Version1), protocol.Version1)
	require.False(t, needsSplit)
	require.Nil(t, frame)
	require.Equal(t, protocol.ByteCount(100), f.DataLen())
	frame, needsSplit = f.MaybeSplitOffFrame(f.Length(protocol.Version1)-1, protocol.Version1)
	require.True(t, needsSplit)
	require.Equal(t, protocol.ByteCount(99), frame.DataLen())
	f.PutBack()
}

func TestStreamSplittingPreservesFINBit(t *testing.T) {
	f := &StreamFrame{
		StreamID: 0x1337,
		Fin:      true,
		Offset:   0xdeadbeef,
		Data:     make([]byte, 100),
	}
	frame, needsSplit := f.MaybeSplitOffFrame(50, protocol.Version1)
	require.True(t, needsSplit)
	require.NotNil(t, frame)
	require.Less(t, frame.Offset, f.Offset)
	require.True(t, f.Fin)
	require.False(t, frame.Fin)
}

func TestStreamSplittingProducesCorrectLengthFramesWithoutDataLen(t *testing.T) {
	const size = 1000
	f := &StreamFrame{
		StreamID: 0xdecafbad,
		Offset:   0x1234,
		Data:     []byte{0},
	}
	minFrameSize := f.Length(protocol.Version1)
	for i := protocol.ByteCount(0); i < minFrameSize; i++ {
		frame, needsSplit := f.MaybeSplitOffFrame(i, protocol.Version1)
		require.True(t, needsSplit)
		require.Nil(t, frame)
	}
	for i := minFrameSize; i < size; i++ {
		f.fromPool = false
		f.Data = make([]byte, size)
		frame, needsSplit := f.MaybeSplitOffFrame(i, protocol.Version1)
		require.True(t, needsSplit)
		require.Equal(t, i, frame.Length(protocol.Version1))
	}
}

func TestStreamSplittingProducesCorrectLengthFramesWithDataLen(t *testing.T) {
	const size = 1000
	f := &StreamFrame{
		StreamID:       0xdecafbad,
		Offset:         0x1234,
		DataLenPresent: true,
		Data:           []byte{0},
	}
	minFrameSize := f.Length(protocol.Version1)
	for i := protocol.ByteCount(0); i < minFrameSize; i++ {
		frame, needsSplit := f.MaybeSplitOffFrame(i, protocol.Version1)
		require.True(t, needsSplit)
		require.Nil(t, frame)
	}
	var frameOneByteTooSmallCounter int
	for i := minFrameSize; i < size; i++ {
		f.fromPool = false
		f.Data = make([]byte, size)
		newFrame, needsSplit := f.MaybeSplitOffFrame(i, protocol.Version1)
		require.True(t, needsSplit)
		// There's *one* pathological case, where a data length of x can be encoded into 1 byte
		// but a data lengths of x+1 needs 2 bytes
		// In that case, it's impossible to create a STREAM frame of the desired size
		if newFrame.Length(protocol.Version1) == i-1 {
			frameOneByteTooSmallCounter++
			continue
		}
		require.Equal(t, i, newFrame.Length(protocol.Version1))
	}
	require.Equal(t, 1, frameOneByteTooSmallCounter)
}
=======

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/qerr"
)

var _ = Describe("StreamFrame", func() {
	Context("when parsing", func() {
		Context("in little endian", func() {
			It("accepts a sample frame", func() {
				// a STREAM frame, plus 3 additional bytes, not belonging to this frame
				b := bytes.NewReader([]byte{0x80 ^ 0x20,
					0x1,      // stream id
					0x6, 0x0, // data length
					'f', 'o', 'o', 'b', 'a', 'r',
					'f', 'o', 'o', // additional bytes
				})
				frame, err := ParseStreamFrame(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(frame.FinBit).To(BeFalse())
				Expect(frame.StreamID).To(Equal(protocol.StreamID(1)))
				Expect(frame.Offset).To(BeZero())
				Expect(frame.DataLenPresent).To(BeTrue())
				Expect(frame.Data).To(Equal([]byte("foobar")))
				Expect(b.Len()).To(Equal(3))
			})

			It("accepts frames with offsets", func() {
				b := bytes.NewReader([]byte{0x80 ^ 0x20 /* 2 byte offset */ ^ 0x4,
					0x1,       // stream id
					0x42, 0x0, // offset
					0x6, 0x0, // data length
					'f', 'o', 'o', 'b', 'a', 'r',
				})
				frame, err := ParseStreamFrame(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(frame.FinBit).To(BeFalse())
				Expect(frame.StreamID).To(Equal(protocol.StreamID(1)))
				Expect(frame.Offset).To(Equal(protocol.ByteCount(0x42)))
				Expect(frame.DataLenPresent).To(BeTrue())
				Expect(frame.Data).To(Equal([]byte("foobar")))
				Expect(b.Len()).To(BeZero())
			})

			It("errors on EOFs", func() {
				data := []byte{0x80 ^ 0x20 ^ 0x4,
					0x1,       // stream id
					0x2a, 0x0, // offset
					0x6, 0x0, // data length,
					'f', 'o', 'o', 'b', 'a', 'r',
				}
				_, err := ParseStreamFrame(bytes.NewReader(data), versionLittleEndian)
				Expect(err).NotTo(HaveOccurred())
				for i := range data {
					_, err := ParseStreamFrame(bytes.NewReader(data[0:i]), versionLittleEndian)
					Expect(err).To(HaveOccurred())
				}
			})
		})

		Context("in big endian", func() {
			It("accepts a sample frame", func() {
				// a STREAM frame, plus 3 additional bytes, not belonging to this frame
				b := bytes.NewReader([]byte{0x80 ^ 0x20,
					0x1,      // stream id
					0x0, 0x6, // data length
					'f', 'o', 'o', 'b', 'a', 'r',
					'f', 'o', 'o', // additional bytes
				})
				frame, err := ParseStreamFrame(b, versionBigEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(frame.FinBit).To(BeFalse())
				Expect(frame.StreamID).To(Equal(protocol.StreamID(1)))
				Expect(frame.Offset).To(BeZero())
				Expect(frame.DataLenPresent).To(BeTrue())
				Expect(frame.Data).To(Equal([]byte("foobar")))
				Expect(b.Len()).To(Equal(3))
			})

			It("accepts frames with offsets", func() {
				b := bytes.NewReader([]byte{0x80 ^ 0x20 /* 2 byte offset */ ^ 0x4,
					0x1,       // stream id
					0x0, 0x42, // offset
					0x0, 0x6, // data length
					'f', 'o', 'o', 'b', 'a', 'r',
				})
				frame, err := ParseStreamFrame(b, versionBigEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(frame.FinBit).To(BeFalse())
				Expect(frame.StreamID).To(Equal(protocol.StreamID(1)))
				Expect(frame.Offset).To(Equal(protocol.ByteCount(0x42)))
				Expect(frame.DataLenPresent).To(BeTrue())
				Expect(frame.Data).To(Equal([]byte("foobar")))
				Expect(b.Len()).To(BeZero())
			})

			It("errors on EOFs", func() {
				data := []byte{0x80 ^ 0x20 ^ 0x4,
					0x1,       // stream id
					0x0, 0x2a, // offset
					0x0, 0x6, // data length,
					'f', 'o', 'o', 'b', 'a', 'r',
				}
				_, err := ParseStreamFrame(bytes.NewReader(data), versionBigEndian)
				Expect(err).NotTo(HaveOccurred())
				for i := range data {
					_, err := ParseStreamFrame(bytes.NewReader(data[0:i]), versionBigEndian)
					Expect(err).To(HaveOccurred())
				}
			})
		})

		It("accepts frame without data length", func() {
			b := bytes.NewReader([]byte{0x80,
				0x1, // stream id
				'f', 'o', 'o', 'b', 'a', 'r',
			})
			frame, err := ParseStreamFrame(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(frame.FinBit).To(BeFalse())
			Expect(frame.StreamID).To(Equal(protocol.StreamID(1)))
			Expect(frame.Offset).To(BeZero())
			Expect(frame.DataLenPresent).To(BeFalse())
			Expect(frame.Data).To(Equal([]byte("foobar")))
			Expect(b.Len()).To(BeZero())
		})

		It("accepts an empty frame with FinBit set, with data length set", func() {
			// the STREAM frame, plus 3 additional bytes, not belonging to this frame
			b := bytes.NewReader([]byte{0x80 ^ 0x40 ^ 0x20,
				0x1,  // stream id
				0, 0, // data length
				'f', 'o', 'o', // additional bytes
			})
			frame, err := ParseStreamFrame(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(frame.FinBit).To(BeTrue())
			Expect(frame.DataLenPresent).To(BeTrue())
			Expect(frame.Data).To(BeEmpty())
			Expect(b.Len()).To(Equal(3))
		})

		It("accepts an empty frame with the FinBit set", func() {
			b := bytes.NewReader([]byte{0x80 ^ 0x40,
				0x1, // stream id
				'f', 'o', 'o', 'b', 'a', 'r',
			})
			frame, err := ParseStreamFrame(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(frame.FinBit).To(BeTrue())
			Expect(frame.DataLenPresent).To(BeFalse())
			Expect(frame.Data).To(Equal([]byte("foobar")))
			Expect(b.Len()).To(BeZero())
		})

		It("errors on empty stream frames that don't have the FinBit set", func() {
			b := bytes.NewReader([]byte{0x80 ^ 0x20,
				0x1,  // stream id
				0, 0, // data length
			})
			_, err := ParseStreamFrame(b, protocol.VersionWhatever)
			Expect(err).To(MatchError(qerr.EmptyStreamFrameNoFin))
		})

		It("rejects frames to too large dataLen", func() {
			b := bytes.NewReader([]byte{0xa0, 0x1, 0xff, 0xf})
			_, err := ParseStreamFrame(b, protocol.VersionWhatever)
			Expect(err).To(MatchError(qerr.Error(qerr.InvalidStreamData, "data len too large")))
		})

		It("rejects frames that overflow the offset", func() {
			// Offset + len(Data) overflows MaxByteCount
			f := &StreamFrame{
				StreamID: 1,
				Offset:   protocol.MaxByteCount,
				Data:     []byte{'f'},
			}
			b := &bytes.Buffer{}
			err := f.Write(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			_, err = ParseStreamFrame(bytes.NewReader(b.Bytes()), protocol.VersionWhatever)
			Expect(err).To(MatchError(qerr.Error(qerr.InvalidStreamData, "data overflows maximum offset")))
		})
	})

	Context("when writing", func() {
		Context("in little endian", func() {
			It("writes sample frame", func() {
				b := &bytes.Buffer{}
				err := (&StreamFrame{
					StreamID:       1,
					Data:           []byte("foobar"),
					DataLenPresent: true,
				}).Write(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()).To(Equal([]byte{0x80 ^ 0x20,
					0x1,      // stream id
					0x6, 0x0, // data length
					'f', 'o', 'o', 'b', 'a', 'r',
				}))
			})
		})

		Context("in big endian", func() {
			It("writes sample frame", func() {
				b := &bytes.Buffer{}
				err := (&StreamFrame{
					StreamID:       1,
					Data:           []byte("foobar"),
					DataLenPresent: true,
				}).Write(b, versionLittleEndian)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()).To(Equal([]byte{0x80 ^ 0x20,
					0x1,      // stream id
					0x6, 0x0, // data length
					'f', 'o', 'o', 'b', 'a', 'r',
				}))
			})
		})

		It("sets the FinBit", func() {
			b := &bytes.Buffer{}
			err := (&StreamFrame{
				StreamID: 1,
				Data:     []byte("foobar"),
				FinBit:   true,
			}).Write(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(b.Bytes()[0] & 0x40).To(Equal(byte(0x40)))
		})

		It("errors when length is zero and FIN is not set", func() {
			b := &bytes.Buffer{}
			err := (&StreamFrame{
				StreamID: 1,
			}).Write(b, protocol.VersionWhatever)
			Expect(err).To(MatchError("StreamFrame: attempting to write empty frame without FIN"))
		})

		It("has proper min length for a short StreamID and a short offset", func() {
			b := &bytes.Buffer{}
			f := &StreamFrame{
				StreamID: 1,
				Data:     []byte{},
				Offset:   0,
				FinBit:   true,
			}
			err := f.Write(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.MinLength(0)).To(Equal(protocol.ByteCount(b.Len())))
		})

		It("has proper min length for a long StreamID and a big offset", func() {
			b := &bytes.Buffer{}
			f := &StreamFrame{
				StreamID: 0xdecafbad,
				Data:     []byte{},
				Offset:   0xdeadbeefcafe,
				FinBit:   true,
			}
			err := f.Write(b, protocol.VersionWhatever)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.MinLength(0)).To(Equal(protocol.ByteCount(b.Len())))
		})

		Context("data length field", func() {
			Context("in little endian", func() {
				It("writes the data length", func() {
					dataLen := 0x1337
					b := &bytes.Buffer{}
					f := &StreamFrame{
						StreamID:       1,
						Data:           bytes.Repeat([]byte{'f'}, dataLen),
						DataLenPresent: true,
						Offset:         0,
					}
					err := f.Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					minLength, _ := f.MinLength(0)
					Expect(b.Bytes()[0] & 0x20).To(Equal(uint8(0x20)))
					Expect(b.Bytes()[minLength-2 : minLength]).To(Equal([]byte{0x37, 0x13}))
				})
			})

			Context("in big endian", func() {
				It("writes the data length", func() {
					dataLen := 0x1337
					b := &bytes.Buffer{}
					f := &StreamFrame{
						StreamID:       1,
						Data:           bytes.Repeat([]byte{'f'}, dataLen),
						DataLenPresent: true,
						Offset:         0,
					}
					err := f.Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					minLength, _ := f.MinLength(0)
					Expect(b.Bytes()[0] & 0x20).To(Equal(uint8(0x20)))
					Expect(b.Bytes()[minLength-2 : minLength]).To(Equal([]byte{0x13, 0x37}))
				})
			})

			It("omits the data length field", func() {
				dataLen := 0x1337
				b := &bytes.Buffer{}
				f := &StreamFrame{
					StreamID:       1,
					Data:           bytes.Repeat([]byte{'f'}, dataLen),
					DataLenPresent: false,
					Offset:         0,
				}
				err := f.Write(b, protocol.VersionWhatever)
				Expect(err).ToNot(HaveOccurred())
				Expect(b.Bytes()[0] & 0x20).To(Equal(uint8(0)))
				Expect(b.Bytes()[1 : b.Len()-dataLen]).ToNot(ContainSubstring(string([]byte{0x37, 0x13})))
				minLength, _ := f.MinLength(0)
				f.DataLenPresent = true
				minLengthWithoutDataLen, _ := f.MinLength(0)
				Expect(minLength).To(Equal(minLengthWithoutDataLen - 2))
			})

			It("calculates the correcct min-length", func() {
				f := &StreamFrame{
					StreamID:       0xCAFE,
					Data:           []byte("foobar"),
					DataLenPresent: false,
					Offset:         0xDEADBEEF,
				}
				minLengthWithoutDataLen, _ := f.MinLength(0)
				f.DataLenPresent = true
				Expect(f.MinLength(0)).To(Equal(minLengthWithoutDataLen + 2))
			})
		})

		Context("offset lengths", func() {
			Context("in little endian", func() {
				It("does not write an offset if the offset is 0", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x0)))
				})

				It("writes a 2-byte offset if the offset is larger than 0", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x1337,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x1 << 2)))
					Expect(b.Bytes()[2:4]).To(Equal([]byte{0x37, 0x13}))
				})

				It("writes a 3-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					(&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x13cafe,
					}).Write(b, versionLittleEndian)
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x2 << 2)))
					Expect(b.Bytes()[2:5]).To(Equal([]byte{0xfe, 0xca, 0x13}))
				})

				It("writes a 4-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0xdeadbeef,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x3 << 2)))
					Expect(b.Bytes()[2:6]).To(Equal([]byte{0xef, 0xbe, 0xad, 0xde}))
				})

				It("writes a 5-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x13deadbeef,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x4 << 2)))
					Expect(b.Bytes()[2:7]).To(Equal([]byte{0xef, 0xbe, 0xad, 0xde, 0x13}))
				})

				It("writes a 6-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0xdeadbeefcafe,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x5 << 2)))
					Expect(b.Bytes()[2:8]).To(Equal([]byte{0xfe, 0xca, 0xef, 0xbe, 0xad, 0xde}))
				})

				It("writes a 7-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x13deadbeefcafe,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x6 << 2)))
					Expect(b.Bytes()[2:9]).To(Equal([]byte{0xfe, 0xca, 0xef, 0xbe, 0xad, 0xde, 0x13}))
				})

				It("writes a 8-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x1337deadbeefcafe,
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x7 << 2)))
					Expect(b.Bytes()[2:10]).To(Equal([]byte{0xfe, 0xca, 0xef, 0xbe, 0xad, 0xde, 0x37, 0x13}))
				})
			})

			Context("in big endian", func() {
				It("does not write an offset if the offset is 0", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x0)))
				})

				It("writes a 2-byte offset if the offset is larger than 0", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x1337,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x1 << 2)))
					Expect(b.Bytes()[2:4]).To(Equal([]byte{0x13, 0x37}))
				})

				It("writes a 3-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					(&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x13cafe,
					}).Write(b, versionBigEndian)
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x2 << 2)))
					Expect(b.Bytes()[2:5]).To(Equal([]byte{0x13, 0xca, 0xfe}))
				})

				It("writes a 4-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0xdeadbeef,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x3 << 2)))
					Expect(b.Bytes()[2:6]).To(Equal([]byte{0xde, 0xad, 0xbe, 0xef}))
				})

				It("writes a 5-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x13deadbeef,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x4 << 2)))
					Expect(b.Bytes()[2:7]).To(Equal([]byte{0x13, 0xde, 0xad, 0xbe, 0xef}))
				})

				It("writes a 6-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0xdeadbeefcafe,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x5 << 2)))
					Expect(b.Bytes()[2:8]).To(Equal([]byte{0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe}))
				})

				It("writes a 7-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x13deadbeefcafe,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x6 << 2)))
					Expect(b.Bytes()[2:9]).To(Equal([]byte{0x13, 0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe}))
				})

				It("writes a 8-byte offset if the offset", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 1,
						Data:     []byte("foobar"),
						Offset:   0x1337deadbeefcafe,
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x1c).To(Equal(uint8(0x7 << 2)))
					Expect(b.Bytes()[2:10]).To(Equal([]byte{0x13, 0x37, 0xde, 0xad, 0xbe, 0xef, 0xca, 0xfe}))
				})
			})
		})

		Context("lengths of StreamIDs", func() {
			Context("in little endian", func() {
				It("writes a 1 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 13,
						Data:     []byte("foobar"),
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x0)))
					Expect(b.Bytes()[1]).To(Equal(uint8(13)))
				})

				It("writes a 2 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 0xcafe,
						Data:     []byte("foobar"),
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x1)))
					Expect(b.Bytes()[1:3]).To(Equal([]byte{0xfe, 0xca}))
				})

				It("writes a 3 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 0x13beef,
						Data:     []byte("foobar"),
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x2)))
					Expect(b.Bytes()[1:4]).To(Equal([]byte{0xef, 0xbe, 0x13}))
				})

				It("writes a 4 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 0xdecafbad,
						Data:     []byte("foobar"),
					}).Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x3)))
					Expect(b.Bytes()[1:5]).To(Equal([]byte{0xad, 0xfb, 0xca, 0xde}))
				})

				It("writes a multiple byte StreamID, after the Stream length was already determined by MinLenght()", func() {
					b := &bytes.Buffer{}
					frame := &StreamFrame{
						StreamID: 0xdecafbad,
						Data:     []byte("foobar"),
					}
					frame.MinLength(0)
					err := frame.Write(b, versionLittleEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x3)))
					Expect(b.Bytes()[1:5]).To(Equal([]byte{0xad, 0xfb, 0xca, 0xde}))
				})
			})

			Context("in big endian", func() {
				It("writes a 1 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 13,
						Data:     []byte("foobar"),
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x0)))
					Expect(b.Bytes()[1]).To(Equal(uint8(13)))
				})

				It("writes a 2 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 0xcafe,
						Data:     []byte("foobar"),
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x1)))
					Expect(b.Bytes()[1:3]).To(Equal([]byte{0xca, 0xfe}))
				})

				It("writes a 3 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 0x13beef,
						Data:     []byte("foobar"),
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x2)))
					Expect(b.Bytes()[1:4]).To(Equal([]byte{0x13, 0xbe, 0xef}))
				})

				It("writes a 4 byte StreamID", func() {
					b := &bytes.Buffer{}
					err := (&StreamFrame{
						StreamID: 0xdecafbad,
						Data:     []byte("foobar"),
					}).Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x3)))
					Expect(b.Bytes()[1:5]).To(Equal([]byte{0xde, 0xca, 0xfb, 0xad}))
				})

				It("writes a multiple byte StreamID, after the Stream length was already determined by MinLenght()", func() {
					b := &bytes.Buffer{}
					frame := &StreamFrame{
						StreamID: 0xdecafbad,
						Data:     []byte("foobar"),
					}
					frame.MinLength(0)
					err := frame.Write(b, versionBigEndian)
					Expect(err).ToNot(HaveOccurred())
					Expect(b.Bytes()[0] & 0x3).To(Equal(uint8(0x3)))
					Expect(b.Bytes()[1:5]).To(Equal([]byte{0xde, 0xca, 0xfb, 0xad}))
				})
			})
		})
	})

	Context("shortening of StreamIDs", func() {
		It("determines the length of a 1 byte StreamID", func() {
			f := &StreamFrame{StreamID: 0xFF}
			Expect(f.calculateStreamIDLength()).To(Equal(uint8(1)))
		})

		It("determines the length of a 2 byte StreamID", func() {
			f := &StreamFrame{StreamID: 0xFFFF}
			Expect(f.calculateStreamIDLength()).To(Equal(uint8(2)))
		})

		It("determines the length of a 3 byte StreamID", func() {
			f := &StreamFrame{StreamID: 0xFFFFFF}
			Expect(f.calculateStreamIDLength()).To(Equal(uint8(3)))
		})

		It("determines the length of a 4 byte StreamID", func() {
			f := &StreamFrame{StreamID: 0xFFFFFFFF}
			Expect(f.calculateStreamIDLength()).To(Equal(uint8(4)))
		})
	})

	Context("shortening of Offsets", func() {
		It("determines length 0 of offset 0", func() {
			f := &StreamFrame{Offset: 0}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(0)))
		})

		It("determines the length of a 2 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(2)))
		})

		It("determines the length of a 2 byte offset, even if it would fit into 1 byte", func() {
			f := &StreamFrame{Offset: 0x1}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(2)))
		})

		It("determines the length of a 3 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(3)))
		})

		It("determines the length of a 4 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFFFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(4)))
		})

		It("determines the length of a 5 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFFFFFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(5)))
		})

		It("determines the length of a 6 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFFFFFFFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(6)))
		})

		It("determines the length of a 7 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFFFFFFFFFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(7)))
		})

		It("determines the length of an 8 byte offset", func() {
			f := &StreamFrame{Offset: 0xFFFFFFFFFFFFFFFF}
			Expect(f.getOffsetLength()).To(Equal(protocol.ByteCount(8)))
		})
	})

	Context("DataLen", func() {
		It("determines the length of the data", func() {
			frame := StreamFrame{
				Data: []byte("foobar"),
			}
			Expect(frame.DataLen()).To(Equal(protocol.ByteCount(6)))
		})
	})
})
>>>>>>> project-faster/main
