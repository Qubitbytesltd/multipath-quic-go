package wire

import "github.com/project-faster/mp-quic-go/internal/protocol"

// AckRange is an ACK range
type AckRange struct {
	First protocol.PacketNumber
	Last  protocol.PacketNumber
}
