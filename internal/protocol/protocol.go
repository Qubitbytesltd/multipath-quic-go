package protocol

<<<<<<< HEAD
import (
	"fmt"
	"time"
)

// The PacketType is the Long Header Type
type PacketType uint8

const (
	// PacketTypeInitial is the packet type of an Initial packet
	PacketTypeInitial PacketType = 1 + iota
	// PacketTypeRetry is the packet type of a Retry packet
	PacketTypeRetry
	// PacketTypeHandshake is the packet type of a Handshake packet
	PacketTypeHandshake
	// PacketType0RTT is the packet type of a 0-RTT packet
	PacketType0RTT
)

func (t PacketType) String() string {
	switch t {
	case PacketTypeInitial:
		return "Initial"
	case PacketTypeRetry:
		return "Retry"
	case PacketTypeHandshake:
		return "Handshake"
	case PacketType0RTT:
		return "0-RTT Protected"
	default:
		return fmt.Sprintf("unknown packet type: %d", t)
	}
}

type ECN uint8

const (
	ECNUnsupported ECN = iota
	ECNNon             // 00
	ECT1               // 01
	ECT0               // 10
	ECNCE              // 11
)

func ParseECNHeaderBits(bits byte) ECN {
	switch bits {
	case 0:
		return ECNNon
	case 0b00000010:
		return ECT0
	case 0b00000001:
		return ECT1
	case 0b00000011:
		return ECNCE
	default:
		panic("invalid ECN bits")
	}
}

func (e ECN) ToHeaderBits() byte {
	//nolint:exhaustive // There are only 4 values.
	switch e {
	case ECNNon:
		return 0
	case ECT0:
		return 0b00000010
	case ECT1:
		return 0b00000001
	case ECNCE:
		return 0b00000011
	default:
		panic("ECN unsupported")
	}
}

func (e ECN) String() string {
	switch e {
	case ECNUnsupported:
		return "ECN unsupported"
	case ECNNon:
		return "Not-ECT"
	case ECT1:
		return "ECT(1)"
	case ECT0:
		return "ECT(0)"
	case ECNCE:
		return "CE"
	default:
		return fmt.Sprintf("invalid ECN value: %d", e)
	}
}

// A ByteCount in QUIC
type ByteCount int64

// MaxByteCount is the maximum value of a ByteCount
const MaxByteCount = ByteCount(1<<62 - 1)

// InvalidByteCount is an invalid byte count
const InvalidByteCount ByteCount = -1

// A StatelessResetToken is a stateless reset token.
type StatelessResetToken [16]byte

// MaxPacketBufferSize maximum packet size of any QUIC packet, based on
// ethernet's max size, minus the IP and UDP headers. IPv6 has a 40 byte header,
// UDP adds an additional 8 bytes.  This is a total overhead of 48 bytes.
// Ethernet's max packet size is 1500 bytes,  1500 - 48 = 1452.
const MaxPacketBufferSize = 1452

// MaxLargePacketBufferSize is used when using GSO
const MaxLargePacketBufferSize = 20 * 1024

// MinInitialPacketSize is the minimum size an Initial packet is required to have.
const MinInitialPacketSize = 1200

// MinUnknownVersionPacketSize is the minimum size a packet with an unknown version
// needs to have in order to trigger a Version Negotiation packet.
const MinUnknownVersionPacketSize = MinInitialPacketSize

// MinStatelessResetSize is the minimum size of a stateless reset packet that we send
const MinStatelessResetSize = 1 /* first byte */ + 20 /* max. conn ID length */ + 4 /* max. packet number length */ + 1 /* min. payload length */ + 16 /* token */

// MinReceivedStatelessResetSize is the minimum size of a received stateless reset,
// as specified in section 10.3 of RFC 9000.
const MinReceivedStatelessResetSize = 5 + 16

// MinConnectionIDLenInitial is the minimum length of the destination connection ID on an Initial packet.
const MinConnectionIDLenInitial = 8

// DefaultAckDelayExponent is the default ack delay exponent
const DefaultAckDelayExponent = 3

// DefaultActiveConnectionIDLimit is the default active connection ID limit
const DefaultActiveConnectionIDLimit = 2

// MaxAckDelayExponent is the maximum ack delay exponent
const MaxAckDelayExponent = 20

// DefaultMaxAckDelay is the default max_ack_delay
const DefaultMaxAckDelay = 25 * time.Millisecond

// MaxMaxAckDelay is the maximum max_ack_delay
const MaxMaxAckDelay = (1<<14 - 1) * time.Millisecond

// MaxConnIDLen is the maximum length of the connection ID
const MaxConnIDLen = 20

// InvalidPacketLimitAES is the maximum number of packets that we can fail to decrypt when using
// AEAD_AES_128_GCM or AEAD_AES_265_GCM.
const InvalidPacketLimitAES = 1 << 52

// InvalidPacketLimitChaCha is the maximum number of packets that we can fail to decrypt when using AEAD_CHACHA20_POLY1305.
const InvalidPacketLimitChaCha = 1 << 36
=======
import "math"

// A PacketNumber in QUIC
type PacketNumber uint64

// PacketNumberLen is the length of the packet number in bytes
type PacketNumberLen uint8

const (
	// PacketNumberLenInvalid is the default value and not a valid length for a packet number
	PacketNumberLenInvalid PacketNumberLen = 0
	// PacketNumberLen1 is a packet number length of 1 byte
	PacketNumberLen1 PacketNumberLen = 1
	// PacketNumberLen2 is a packet number length of 2 bytes
	PacketNumberLen2 PacketNumberLen = 2
	// PacketNumberLen4 is a packet number length of 4 bytes
	PacketNumberLen4 PacketNumberLen = 4
	// PacketNumberLen6 is a packet number length of 6 bytes
	PacketNumberLen6 PacketNumberLen = 6
)

// A ConnectionID in QUIC
type ConnectionID uint64

// A StreamID in QUIC
type StreamID uint32

// A PathID in QUIC
type PathID uint8

// Initial PathID
const InitialPathID = 0

// A ByteCount in QUIC
type ByteCount uint64

// MaxByteCount is the maximum value of a ByteCount
const MaxByteCount = ByteCount(math.MaxUint64)

// MaxReceivePacketSize maximum packet size of any QUIC packet, based on
// ethernet's max size, minus the IP and UDP headers. IPv6 has a 40 byte header,
// UDP adds an additional 8 bytes.  This is a total overhead of 48 bytes.
// Ethernet's max packet size is 1500 bytes,  1500 - 48 = 1452.
const MaxReceivePacketSize ByteCount = 1452

// DefaultTCPMSS is the default maximum packet size used in the Linux TCP implementation.
// Used in QUIC for congestion window computations in bytes.
const DefaultTCPMSS ByteCount = 1460

// InitialStreamFlowControlWindow is the initial stream-level flow control window for sending
const InitialStreamFlowControlWindow ByteCount = (1 << 14) // 16 kB

// InitialConnectionFlowControlWindow is the initial connection-level flow control window for sending
const InitialConnectionFlowControlWindow ByteCount = (1 << 14) // 16 kB

// ClientHelloMinimumSize is the minimum size the server expects an inchoate CHLO to have.
const ClientHelloMinimumSize = 1024

// MaxClientHellos is the maximum number of times we'll send a client hello
// The value 3 accounts for:
// * one failure due to an incorrect or missing source-address token
// * one failure due the server's certificate chain being unavailible and the server being unwilling to send it without a valid source-address token
const MaxClientHellos = 3
>>>>>>> project-faster/main
