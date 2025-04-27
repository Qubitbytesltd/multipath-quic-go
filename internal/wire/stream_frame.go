package wire

import (
<<<<<<< HEAD
	"errors"
	"io"

	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
=======
	"bytes"
	"errors"
	"io"

	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/internal/utils"
	"github.com/project-faster/mp-quic-go/qerr"
>>>>>>> project-faster/main
)

// A StreamFrame of QUIC
type StreamFrame struct {
	StreamID       protocol.StreamID
<<<<<<< HEAD
	Offset         protocol.ByteCount
	Data           []byte
	Fin            bool
	DataLenPresent bool

	fromPool bool
}

func parseStreamFrame(b []byte, typ uint64, _ protocol.Version) (*StreamFrame, int, error) {
	startLen := len(b)
	hasOffset := typ&0b100 > 0
	fin := typ&0b1 > 0
	hasDataLen := typ&0b10 > 0

	streamID, l, err := quicvarint.Parse(b)
	if err != nil {
		return nil, 0, replaceUnexpectedEOF(err)
	}
	b = b[l:]
	var offset uint64
	if hasOffset {
		offset, l, err = quicvarint.Parse(b)
		if err != nil {
			return nil, 0, replaceUnexpectedEOF(err)
		}
		b = b[l:]
	}

	var dataLen uint64
	if hasDataLen {
		var err error
		var l int
		dataLen, l, err = quicvarint.Parse(b)
		if err != nil {
			return nil, 0, replaceUnexpectedEOF(err)
		}
		b = b[l:]
		if dataLen > uint64(len(b)) {
			return nil, 0, io.EOF
		}
	} else {
		// The rest of the packet is data
		dataLen = uint64(len(b))
	}

	var frame *StreamFrame
	if dataLen < protocol.MinStreamFrameBufferSize {
		frame = &StreamFrame{}
		if dataLen > 0 {
			frame.Data = make([]byte, dataLen)
		}
	} else {
		frame = GetStreamFrame()
		// The STREAM frame can't be larger than the StreamFrame we obtained from the buffer,
		// since those StreamFrames have a buffer length of the maximum packet size.
		if dataLen > uint64(cap(frame.Data)) {
			return nil, 0, io.EOF
		}
		frame.Data = frame.Data[:dataLen]
	}

	frame.StreamID = protocol.StreamID(streamID)
	frame.Offset = protocol.ByteCount(offset)
	frame.Fin = fin
	frame.DataLenPresent = hasDataLen

	if dataLen > 0 {
		copy(frame.Data, b)
	}
	if frame.Offset+frame.DataLen() > protocol.MaxByteCount {
		return nil, 0, errors.New("stream data overflows maximum offset")
	}
	return frame, startLen - len(b) + int(dataLen), nil
}

func (f *StreamFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	if len(f.Data) == 0 && !f.Fin {
		return nil, errors.New("StreamFrame: attempting to write empty frame without FIN")
	}

	typ := byte(0x8)
	if f.Fin {
		typ ^= 0b1
	}
	hasOffset := f.Offset != 0
	if f.DataLenPresent {
		typ ^= 0b10
	}
	if hasOffset {
		typ ^= 0b100
	}
	b = append(b, typ)
	b = quicvarint.Append(b, uint64(f.StreamID))
	if hasOffset {
		b = quicvarint.Append(b, uint64(f.Offset))
	}
	if f.DataLenPresent {
		b = quicvarint.Append(b, uint64(f.DataLen()))
	}
	b = append(b, f.Data...)
	return b, nil
}

// Length returns the total length of the STREAM frame
func (f *StreamFrame) Length(protocol.Version) protocol.ByteCount {
	length := 1 + quicvarint.Len(uint64(f.StreamID))
	if f.Offset != 0 {
		length += quicvarint.Len(uint64(f.Offset))
	}
	if f.DataLenPresent {
		length += quicvarint.Len(uint64(f.DataLen()))
	}
	return protocol.ByteCount(length) + f.DataLen()
=======
	FinBit         bool
	DataLenPresent bool
	Offset         protocol.ByteCount
	Data           []byte
}

var (
	errInvalidStreamIDLen = errors.New("StreamFrame: Invalid StreamID length")
	errInvalidOffsetLen   = errors.New("StreamFrame: Invalid offset length")
)

// ParseStreamFrame reads a stream frame. The type byte must not have been read yet.
func ParseStreamFrame(r *bytes.Reader, version protocol.VersionNumber) (*StreamFrame, error) {
	frame := &StreamFrame{}

	typeByte, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	frame.FinBit = typeByte&0x40 > 0
	frame.DataLenPresent = typeByte&0x20 > 0
	offsetLen := typeByte & 0x1c >> 2
	if offsetLen != 0 {
		offsetLen++
	}
	streamIDLen := typeByte&0x3 + 1

	sid, err := utils.GetByteOrder(version).ReadUintN(r, streamIDLen)
	if err != nil {
		return nil, err
	}
	frame.StreamID = protocol.StreamID(sid)

	offset, err := utils.GetByteOrder(version).ReadUintN(r, offsetLen)
	if err != nil {
		return nil, err
	}
	frame.Offset = protocol.ByteCount(offset)

	var dataLen uint16
	if frame.DataLenPresent {
		dataLen, err = utils.GetByteOrder(version).ReadUint16(r)
		if err != nil {
			return nil, err
		}
	}

	if dataLen > uint16(protocol.MaxPacketSize) {
		return nil, qerr.Error(qerr.InvalidStreamData, "data len too large")
	}

	if !frame.DataLenPresent {
		// The rest of the packet is data
		dataLen = uint16(r.Len())
	}
	if dataLen != 0 {
		frame.Data = make([]byte, dataLen)
		if _, err := io.ReadFull(r, frame.Data); err != nil {
			return nil, err
		}
	}

	if frame.Offset+frame.DataLen() < frame.Offset {
		return nil, qerr.Error(qerr.InvalidStreamData, "data overflows maximum offset")
	}
	if !frame.FinBit && frame.DataLen() == 0 {
		return nil, qerr.EmptyStreamFrameNoFin
	}
	return frame, nil
}

// WriteStreamFrame writes a stream frame.
func (f *StreamFrame) Write(b *bytes.Buffer, version protocol.VersionNumber) error {
	if len(f.Data) == 0 && !f.FinBit {
		return errors.New("StreamFrame: attempting to write empty frame without FIN")
	}

	typeByte := uint8(0x80) // sets the leftmost bit to 1
	if f.FinBit {
		typeByte ^= 0x40
	}
	if f.DataLenPresent {
		typeByte ^= 0x20
	}

	offsetLength := f.getOffsetLength()
	if offsetLength > 0 {
		typeByte ^= (uint8(offsetLength) - 1) << 2
	}

	streamIDLen := f.calculateStreamIDLength()
	typeByte ^= streamIDLen - 1

	b.WriteByte(typeByte)

	switch streamIDLen {
	case 1:
		b.WriteByte(uint8(f.StreamID))
	case 2:
		utils.GetByteOrder(version).WriteUint16(b, uint16(f.StreamID))
	case 3:
		utils.GetByteOrder(version).WriteUint24(b, uint32(f.StreamID))
	case 4:
		utils.GetByteOrder(version).WriteUint32(b, uint32(f.StreamID))
	default:
		return errInvalidStreamIDLen
	}

	switch offsetLength {
	case 0:
	case 2:
		utils.GetByteOrder(version).WriteUint16(b, uint16(f.Offset))
	case 3:
		utils.GetByteOrder(version).WriteUint24(b, uint32(f.Offset))
	case 4:
		utils.GetByteOrder(version).WriteUint32(b, uint32(f.Offset))
	case 5:
		utils.GetByteOrder(version).WriteUint40(b, uint64(f.Offset))
	case 6:
		utils.GetByteOrder(version).WriteUint48(b, uint64(f.Offset))
	case 7:
		utils.GetByteOrder(version).WriteUint56(b, uint64(f.Offset))
	case 8:
		utils.GetByteOrder(version).WriteUint64(b, uint64(f.Offset))
	default:
		return errInvalidOffsetLen
	}

	if f.DataLenPresent {
		utils.GetByteOrder(version).WriteUint16(b, uint16(len(f.Data)))
	}

	b.Write(f.Data)
	return nil
}

func (f *StreamFrame) calculateStreamIDLength() uint8 {
	if f.StreamID < (1 << 8) {
		return 1
	} else if f.StreamID < (1 << 16) {
		return 2
	} else if f.StreamID < (1 << 24) {
		return 3
	}
	return 4
}

func (f *StreamFrame) getOffsetLength() protocol.ByteCount {
	if f.Offset == 0 {
		return 0
	}
	if f.Offset < (1 << 16) {
		return 2
	}
	if f.Offset < (1 << 24) {
		return 3
	}
	if f.Offset < (1 << 32) {
		return 4
	}
	if f.Offset < (1 << 40) {
		return 5
	}
	if f.Offset < (1 << 48) {
		return 6
	}
	if f.Offset < (1 << 56) {
		return 7
	}
	return 8
}

// MinLength returns the length of the header of a StreamFrame
// the total length of the StreamFrame is frame.MinLength() + frame.DataLen()
func (f *StreamFrame) MinLength(protocol.VersionNumber) (protocol.ByteCount, error) {
	length := protocol.ByteCount(1) + protocol.ByteCount(f.calculateStreamIDLength()) + f.getOffsetLength()
	if f.DataLenPresent {
		length += 2
	}
	return length, nil
>>>>>>> project-faster/main
}

// DataLen gives the length of data in bytes
func (f *StreamFrame) DataLen() protocol.ByteCount {
	return protocol.ByteCount(len(f.Data))
}
<<<<<<< HEAD

// MaxDataLen returns the maximum data length
// If 0 is returned, writing will fail (a STREAM frame must contain at least 1 byte of data).
func (f *StreamFrame) MaxDataLen(maxSize protocol.ByteCount, _ protocol.Version) protocol.ByteCount {
	headerLen := 1 + protocol.ByteCount(quicvarint.Len(uint64(f.StreamID)))
	if f.Offset != 0 {
		headerLen += protocol.ByteCount(quicvarint.Len(uint64(f.Offset)))
	}
	if f.DataLenPresent {
		// Pretend that the data size will be 1 byte.
		// If it turns out that varint encoding the length will consume 2 bytes, we need to adjust the data length afterward
		headerLen++
	}
	if headerLen > maxSize {
		return 0
	}
	maxDataLen := maxSize - headerLen
	if f.DataLenPresent && quicvarint.Len(uint64(maxDataLen)) != 1 {
		maxDataLen--
	}
	return maxDataLen
}

// MaybeSplitOffFrame splits a frame such that it is not bigger than n bytes.
// It returns if the frame was actually split.
// The frame might not be split if:
// * the size is large enough to fit the whole frame
// * the size is too small to fit even a 1-byte frame. In that case, the frame returned is nil.
func (f *StreamFrame) MaybeSplitOffFrame(maxSize protocol.ByteCount, version protocol.Version) (*StreamFrame, bool /* was splitting required */) {
	if maxSize >= f.Length(version) {
		return nil, false
	}

	n := f.MaxDataLen(maxSize, version)
	if n == 0 {
		return nil, true
	}

	new := GetStreamFrame()
	new.StreamID = f.StreamID
	new.Offset = f.Offset
	new.Fin = false
	new.DataLenPresent = f.DataLenPresent

	// swap the data slices
	new.Data, f.Data = f.Data, new.Data
	new.fromPool, f.fromPool = f.fromPool, new.fromPool

	f.Data = f.Data[:protocol.ByteCount(len(new.Data))-n]
	copy(f.Data, new.Data[n:])
	new.Data = new.Data[:n]
	f.Offset += n

	return new, true
}

func (f *StreamFrame) PutBack() {
	putStreamFrame(f)
}
=======
>>>>>>> project-faster/main
