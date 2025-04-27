package wire

<<<<<<< HEAD
import "github.com/quic-go/quic-go/internal/protocol"

// AckRange is an ACK range
type AckRange struct {
	Smallest protocol.PacketNumber
	Largest  protocol.PacketNumber
}

// Len returns the number of packets contained in this ACK range
func (r AckRange) Len() protocol.PacketNumber {
	return r.Largest - r.Smallest + 1
=======
import "github.com/project-faster/mp-quic-go/internal/protocol"

// AckRange is an ACK range
type AckRange struct {
	First protocol.PacketNumber
	Last  protocol.PacketNumber
>>>>>>> project-faster/main
}
