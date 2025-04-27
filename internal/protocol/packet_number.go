package protocol

<<<<<<< HEAD
// A PacketNumber in QUIC
type PacketNumber int64

// InvalidPacketNumber is a packet number that is never sent.
// In QUIC, 0 is a valid packet number.
const InvalidPacketNumber PacketNumber = -1

// PacketNumberLen is the length of the packet number in bytes
type PacketNumberLen uint8

const (
	// PacketNumberLen1 is a packet number length of 1 byte
	PacketNumberLen1 PacketNumberLen = 1
	// PacketNumberLen2 is a packet number length of 2 bytes
	PacketNumberLen2 PacketNumberLen = 2
	// PacketNumberLen3 is a packet number length of 3 bytes
	PacketNumberLen3 PacketNumberLen = 3
	// PacketNumberLen4 is a packet number length of 4 bytes
	PacketNumberLen4 PacketNumberLen = 4
)

// DecodePacketNumber calculates the packet number based its length and the last seen packet number
// This function is taken from https://www.rfc-editor.org/rfc/rfc9000.html#section-a.3.
func DecodePacketNumber(length PacketNumberLen, largest PacketNumber, truncated PacketNumber) PacketNumber {
	expected := largest + 1
	win := PacketNumber(1 << (length * 8))
	hwin := win / 2
	mask := win - 1
	candidate := (expected & ^mask) | truncated
	if candidate <= expected-hwin && candidate < 1<<62-win {
		return candidate + win
	}
	if candidate > expected+hwin && candidate >= win {
		return candidate - win
	}
	return candidate
}

// PacketNumberLengthForHeader gets the length of the packet number for the public header
// it never chooses a PacketNumberLen of 1 byte, since this is too short under certain circumstances
func PacketNumberLengthForHeader(pn, largestAcked PacketNumber) PacketNumberLen {
	var numUnacked PacketNumber
	if largestAcked == InvalidPacketNumber {
		numUnacked = pn + 1
	} else {
		numUnacked = pn - largestAcked
	}
	if numUnacked < 1<<(16-1) {
		return PacketNumberLen2
	}
	if numUnacked < 1<<(24-1) {
		return PacketNumberLen3
	}
	return PacketNumberLen4
=======
// InferPacketNumber calculates the packet number based on the received packet number, its length and the last seen packet number
func InferPacketNumber(packetNumberLength PacketNumberLen, lastPacketNumber PacketNumber, wirePacketNumber PacketNumber) PacketNumber {
	epochDelta := PacketNumber(1) << (uint8(packetNumberLength) * 8)
	epoch := lastPacketNumber & ^(epochDelta - 1)
	prevEpochBegin := epoch - epochDelta
	nextEpochBegin := epoch + epochDelta
	return closestTo(
		lastPacketNumber+1,
		epoch+wirePacketNumber,
		closestTo(lastPacketNumber+1, prevEpochBegin+wirePacketNumber, nextEpochBegin+wirePacketNumber),
	)
}

func closestTo(target, a, b PacketNumber) PacketNumber {
	if delta(target, a) < delta(target, b) {
		return a
	}
	return b
}

func delta(a, b PacketNumber) PacketNumber {
	if a < b {
		return b - a
	}
	return a - b
}

// GetPacketNumberLengthForPublicHeader gets the length of the packet number for the public header
// it never chooses a PacketNumberLen of 1 byte, since this is too short under certain circumstances
func GetPacketNumberLengthForPublicHeader(packetNumber PacketNumber, leastUnacked PacketNumber) PacketNumberLen {
	diff := uint64(packetNumber - leastUnacked)
	if diff < (2 << (uint8(PacketNumberLen2)*8 - 2)) {
		return PacketNumberLen2
	}
	if diff < (2 << (uint8(PacketNumberLen4)*8 - 2)) {
		return PacketNumberLen4
	}
	// we do not check if there are less than 2^46 packets in flight, since flow control and congestion control will limit this number *a lot* sooner
	return PacketNumberLen6
}

// GetPacketNumberLength gets the minimum length needed to fully represent the packet number
func GetPacketNumberLength(packetNumber PacketNumber) PacketNumberLen {
	if packetNumber < (1 << (uint8(PacketNumberLen1) * 8)) {
		return PacketNumberLen1
	}
	if packetNumber < (1 << (uint8(PacketNumberLen2) * 8)) {
		return PacketNumberLen2
	}
	if packetNumber < (1 << (uint8(PacketNumberLen4) * 8)) {
		return PacketNumberLen4
	}
	return PacketNumberLen6
>>>>>>> project-faster/main
}
