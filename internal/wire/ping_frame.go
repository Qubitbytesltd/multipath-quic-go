package wire

import (
<<<<<<< HEAD
	"github.com/quic-go/quic-go/internal/protocol"
)

// A PingFrame is a PING frame
type PingFrame struct{}

func (f *PingFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	return append(b, pingFrameType), nil
}

// Length of a written frame
func (f *PingFrame) Length(_ protocol.Version) protocol.ByteCount {
	return 1
=======
	"bytes"

	"github.com/project-faster/mp-quic-go/internal/protocol"
)

// A PingFrame is a ping frame
type PingFrame struct{}

// ParsePingFrame parses a Ping frame
func ParsePingFrame(r *bytes.Reader, version protocol.VersionNumber) (*PingFrame, error) {
	frame := &PingFrame{}

	_, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	return frame, nil
}

func (f *PingFrame) Write(b *bytes.Buffer, version protocol.VersionNumber) error {
	typeByte := uint8(0x07)
	b.WriteByte(typeByte)
	return nil
}

// MinLength of a written frame
func (f *PingFrame) MinLength(version protocol.VersionNumber) (protocol.ByteCount, error) {
	return 1, nil
>>>>>>> project-faster/main
}
