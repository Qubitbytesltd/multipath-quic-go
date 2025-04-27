package wire

import (
<<<<<<< HEAD
	"io"

	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/quicvarint"
)

// A ConnectionCloseFrame is a CONNECTION_CLOSE frame
type ConnectionCloseFrame struct {
	IsApplicationError bool
	ErrorCode          uint64
	FrameType          uint64
	ReasonPhrase       string
}

func parseConnectionCloseFrame(b []byte, typ uint64, _ protocol.Version) (*ConnectionCloseFrame, int, error) {
	startLen := len(b)
	f := &ConnectionCloseFrame{IsApplicationError: typ == applicationCloseFrameType}
	ec, l, err := quicvarint.Parse(b)
	if err != nil {
		return nil, 0, replaceUnexpectedEOF(err)
	}
	b = b[l:]
	f.ErrorCode = ec
	// read the Frame Type, if this is not an application error
	if !f.IsApplicationError {
		ft, l, err := quicvarint.Parse(b)
		if err != nil {
			return nil, 0, replaceUnexpectedEOF(err)
		}
		b = b[l:]
		f.FrameType = ft
	}
	var reasonPhraseLen uint64
	reasonPhraseLen, l, err = quicvarint.Parse(b)
	if err != nil {
		return nil, 0, replaceUnexpectedEOF(err)
	}
	b = b[l:]
	if int(reasonPhraseLen) > len(b) {
		return nil, 0, io.EOF
	}

	reasonPhrase := make([]byte, reasonPhraseLen)
	copy(reasonPhrase, b)
	f.ReasonPhrase = string(reasonPhrase)
	return f, startLen - len(b) + int(reasonPhraseLen), nil
}

// Length of a written frame
func (f *ConnectionCloseFrame) Length(protocol.Version) protocol.ByteCount {
	length := 1 + protocol.ByteCount(quicvarint.Len(f.ErrorCode)+quicvarint.Len(uint64(len(f.ReasonPhrase)))) + protocol.ByteCount(len(f.ReasonPhrase))
	if !f.IsApplicationError {
		length += protocol.ByteCount(quicvarint.Len(f.FrameType)) // for the frame type
	}
	return length
}

func (f *ConnectionCloseFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	if f.IsApplicationError {
		b = append(b, applicationCloseFrameType)
	} else {
		b = append(b, connectionCloseFrameType)
	}

	b = quicvarint.Append(b, f.ErrorCode)
	if !f.IsApplicationError {
		b = quicvarint.Append(b, f.FrameType)
	}
	b = quicvarint.Append(b, uint64(len(f.ReasonPhrase)))
	b = append(b, []byte(f.ReasonPhrase)...)
	return b, nil
=======
	"bytes"
	"errors"
	"io"
	"math"

	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/internal/utils"
	"github.com/project-faster/mp-quic-go/qerr"
)

// A ConnectionCloseFrame in QUIC
type ConnectionCloseFrame struct {
	ErrorCode    qerr.ErrorCode
	ReasonPhrase string
}

// ParseConnectionCloseFrame reads a CONNECTION_CLOSE frame
func ParseConnectionCloseFrame(r *bytes.Reader, version protocol.VersionNumber) (*ConnectionCloseFrame, error) {
	frame := &ConnectionCloseFrame{}

	// read the TypeByte
	_, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	errorCode, err := utils.GetByteOrder(version).ReadUint32(r)
	if err != nil {
		return nil, err
	}
	frame.ErrorCode = qerr.ErrorCode(errorCode)

	reasonPhraseLen, err := utils.GetByteOrder(version).ReadUint16(r)
	if err != nil {
		return nil, err
	}

	if reasonPhraseLen > uint16(protocol.MaxPacketSize) {
		return nil, qerr.Error(qerr.InvalidConnectionCloseData, "reason phrase too long")
	}

	reasonPhrase := make([]byte, reasonPhraseLen)
	if _, err := io.ReadFull(r, reasonPhrase); err != nil {
		return nil, err
	}
	frame.ReasonPhrase = string(reasonPhrase)

	return frame, nil
}

// MinLength of a written frame
func (f *ConnectionCloseFrame) MinLength(version protocol.VersionNumber) (protocol.ByteCount, error) {
	return 1 + 4 + 2 + protocol.ByteCount(len(f.ReasonPhrase)), nil
}

// Write writes an CONNECTION_CLOSE frame.
func (f *ConnectionCloseFrame) Write(b *bytes.Buffer, version protocol.VersionNumber) error {
	b.WriteByte(0x02)
	utils.GetByteOrder(version).WriteUint32(b, uint32(f.ErrorCode))

	if len(f.ReasonPhrase) > math.MaxUint16 {
		return errors.New("ConnectionFrame: ReasonPhrase too long")
	}

	reasonPhraseLen := uint16(len(f.ReasonPhrase))
	utils.GetByteOrder(version).WriteUint16(b, reasonPhraseLen)
	b.WriteString(f.ReasonPhrase)

	return nil
>>>>>>> project-faster/main
}
