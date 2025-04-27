package wire

import (
<<<<<<< HEAD
	"github.com/quic-go/quic-go/internal/protocol"
=======
	"bytes"

	"github.com/project-faster/mp-quic-go/internal/protocol"
>>>>>>> project-faster/main
)

// A Frame in QUIC
type Frame interface {
<<<<<<< HEAD
	Append(b []byte, version protocol.Version) ([]byte, error)
	Length(version protocol.Version) protocol.ByteCount
}

// IsProbingFrame returns true if the frame is a probing frame.
// See section 9.1 of RFC 9000.
func IsProbingFrame(f Frame) bool {
	switch f.(type) {
	case *PathChallengeFrame, *PathResponseFrame, *NewConnectionIDFrame:
		return true
	}
	return false
=======
	Write(b *bytes.Buffer, version protocol.VersionNumber) error
	MinLength(version protocol.VersionNumber) (protocol.ByteCount, error)
>>>>>>> project-faster/main
}
