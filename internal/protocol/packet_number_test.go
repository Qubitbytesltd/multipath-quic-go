package protocol

import (
<<<<<<< HEAD
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInvalidPacketNumberIsSmallerThanAllValidPacketNumbers(t *testing.T) {
	require.Less(t, InvalidPacketNumber, PacketNumber(0))
}

func TestPacketNumberLenHasCorrectValue(t *testing.T) {
	require.EqualValues(t, 1, PacketNumberLen1)
	require.EqualValues(t, 2, PacketNumberLen2)
	require.EqualValues(t, 3, PacketNumberLen3)
	require.EqualValues(t, 4, PacketNumberLen4)
}

func TestDecodePacketNumber(t *testing.T) {
	require.Equal(t, PacketNumber(255), DecodePacketNumber(PacketNumberLen1, 10, 255))
	require.Equal(t, PacketNumber(0), DecodePacketNumber(PacketNumberLen1, 10, 0))
	require.Equal(t, PacketNumber(256), DecodePacketNumber(PacketNumberLen1, 127, 0))
	require.Equal(t, PacketNumber(256), DecodePacketNumber(PacketNumberLen1, 128, 0))
	require.Equal(t, PacketNumber(256), DecodePacketNumber(PacketNumberLen1, 256+126, 0))
	require.Equal(t, PacketNumber(512), DecodePacketNumber(PacketNumberLen1, 256+127, 0))
	require.Equal(t, PacketNumber(0xffff), DecodePacketNumber(PacketNumberLen2, 0xffff, 0xffff))
	require.Equal(t, PacketNumber(0xffff), DecodePacketNumber(PacketNumberLen2, 0xffff+1, 0xffff))

	// example from https://www.rfc-editor.org/rfc/rfc9000.html#section-a.3
	require.Equal(t, PacketNumber(0xa82f9b32), DecodePacketNumber(PacketNumberLen2, 0xa82f30ea, 0x9b32))
}

func TestPacketNumberLengthForHeader(t *testing.T) {
	require.Equal(t, PacketNumberLen2, PacketNumberLengthForHeader(1, InvalidPacketNumber))
	require.Equal(t, PacketNumberLen2, PacketNumberLengthForHeader(1<<15-2, InvalidPacketNumber))
	require.Equal(t, PacketNumberLen3, PacketNumberLengthForHeader(1<<15-1, InvalidPacketNumber))
	require.Equal(t, PacketNumberLen3, PacketNumberLengthForHeader(1<<23-2, InvalidPacketNumber))
	require.Equal(t, PacketNumberLen4, PacketNumberLengthForHeader(1<<23-1, InvalidPacketNumber))
	require.Equal(t, PacketNumberLen2, PacketNumberLengthForHeader(1<<15+9, 10))
	require.Equal(t, PacketNumberLen3, PacketNumberLengthForHeader(1<<15+10, 10))
	require.Equal(t, PacketNumberLen3, PacketNumberLengthForHeader(1<<23+99, 100))
	require.Equal(t, PacketNumberLen4, PacketNumberLengthForHeader(1<<23+100, 100))
	// examples from https://www.rfc-editor.org/rfc/rfc9000.html#section-a.2
	require.Equal(t, PacketNumberLen2, PacketNumberLengthForHeader(0xac5c02, 0xabe8b3))
	require.Equal(t, PacketNumberLen3, PacketNumberLengthForHeader(0xace8fe, 0xabe8b3))
}
=======
	"fmt"
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Tests taken and extended from chrome
var _ = Describe("packet number calculation", func() {
	Context("infering a packet number", func() {
		check := func(length PacketNumberLen, expected, last uint64) {
			epoch := uint64(1) << (length * 8)
			epochMask := epoch - 1
			wirePacketNumber := expected & epochMask
			Expect(InferPacketNumber(length, PacketNumber(last), PacketNumber(wirePacketNumber))).To(Equal(PacketNumber(expected)))
		}
		for _, length := range []PacketNumberLen{PacketNumberLen1, PacketNumberLen2, PacketNumberLen4, PacketNumberLen6} {
			Context(fmt.Sprintf("with %d bytes", length), func() {
				epoch := uint64(1) << (length * 8)
				epochMask := epoch - 1

				It("works near epoch start", func() {
					// A few quick manual sanity check
					check(length, 1, 0)
					check(length, epoch+1, epochMask)
					check(length, epoch, epochMask)

					// Cases where the last number was close to the start of the range.
					for last := uint64(0); last < 10; last++ {
						// Small numbers should not wrap (even if they're out of order).
						for j := uint64(0); j < 10; j++ {
							check(length, j, last)
						}

						// Large numbers should not wrap either (because we're near 0 already).
						for j := uint64(0); j < 10; j++ {
							check(length, epoch-1-j, last)
						}
					}
				})

				It("works near epoch end", func() {
					// Cases where the last number was close to the end of the range
					for i := uint64(0); i < 10; i++ {
						last := epoch - i

						// Small numbers should wrap.
						for j := uint64(0); j < 10; j++ {
							check(length, epoch+j, last)
						}

						// Large numbers should not (even if they're out of order).
						for j := uint64(0); j < 10; j++ {
							check(length, epoch-1-j, last)
						}
					}
				})

				// Next check where we're in a non-zero epoch to verify we handle
				// reverse wrapping, too.
				It("works near previous epoch", func() {
					prevEpoch := 1 * epoch
					curEpoch := 2 * epoch
					// Cases where the last number was close to the start of the range
					for i := uint64(0); i < 10; i++ {
						last := curEpoch + i
						// Small number should not wrap (even if they're out of order).
						for j := uint64(0); j < 10; j++ {
							check(length, curEpoch+j, last)
						}

						// But large numbers should reverse wrap.
						for j := uint64(0); j < 10; j++ {
							num := epoch - 1 - j
							check(length, prevEpoch+num, last)
						}
					}
				})

				It("works near next epoch", func() {
					curEpoch := 2 * epoch
					nextEpoch := 3 * epoch
					// Cases where the last number was close to the end of the range
					for i := uint64(0); i < 10; i++ {
						last := nextEpoch - 1 - i

						// Small numbers should wrap.
						for j := uint64(0); j < 10; j++ {
							check(length, nextEpoch+j, last)
						}

						// but large numbers should not (even if they're out of order).
						for j := uint64(0); j < 10; j++ {
							num := epoch - 1 - j
							check(length, curEpoch+num, last)
						}
					}
				})

				It("works near next max", func() {
					maxNumber := uint64(math.MaxUint64)
					maxEpoch := maxNumber & ^epochMask

					// Cases where the last number was close to the end of the range
					for i := uint64(0); i < 10; i++ {
						// Subtract 1, because the expected next packet number is 1 more than the
						// last packet number.
						last := maxNumber - i - 1

						// Small numbers should not wrap, because they have nowhere to go.
						for j := uint64(0); j < 10; j++ {
							check(length, maxEpoch+j, last)
						}

						// Large numbers should not wrap either.
						for j := uint64(0); j < 10; j++ {
							num := epoch - 1 - j
							check(length, maxEpoch+num, last)
						}
					}
				})
			})
		}
	})

	Context("shortening a packet number for the publicHeader", func() {
		Context("shortening", func() {
			It("sends out low packet numbers as 2 byte", func() {
				length := GetPacketNumberLengthForPublicHeader(4, 2)
				Expect(length).To(Equal(PacketNumberLen2))
			})

			It("sends out high packet numbers as 2 byte, if all ACKs are received", func() {
				length := GetPacketNumberLengthForPublicHeader(0xDEADBEEF, 0xDEADBEEF-1)
				Expect(length).To(Equal(PacketNumberLen2))
			})

			It("sends out higher packet numbers as 4 bytes, if a lot of ACKs are missing", func() {
				length := GetPacketNumberLengthForPublicHeader(40000, 2)
				Expect(length).To(Equal(PacketNumberLen4))
			})
		})

		Context("self-consistency", func() {
			It("works for small packet numbers", func() {
				for i := uint64(1); i < 10000; i++ {
					packetNumber := PacketNumber(i)
					leastUnacked := PacketNumber(1)
					length := GetPacketNumberLengthForPublicHeader(packetNumber, leastUnacked)
					wirePacketNumber := (uint64(packetNumber) << (64 - length*8)) >> (64 - length*8)

					inferedPacketNumber := InferPacketNumber(length, leastUnacked, PacketNumber(wirePacketNumber))
					Expect(inferedPacketNumber).To(Equal(packetNumber))
				}
			})

			It("works for small packet numbers and increasing ACKed packets", func() {
				for i := uint64(1); i < 10000; i++ {
					packetNumber := PacketNumber(i)
					leastUnacked := PacketNumber(i / 2)
					length := GetPacketNumberLengthForPublicHeader(packetNumber, leastUnacked)
					wirePacketNumber := (uint64(packetNumber) << (64 - length*8)) >> (64 - length*8)

					inferedPacketNumber := InferPacketNumber(length, leastUnacked, PacketNumber(wirePacketNumber))
					Expect(inferedPacketNumber).To(Equal(packetNumber))
				}
			})

			It("also works for larger packet numbers", func() {
				increment := uint64(1 << (8 - 3))
				for i := uint64(1); i < (2 << 46); i += increment {
					packetNumber := PacketNumber(i)
					leastUnacked := PacketNumber(1)
					length := GetPacketNumberLengthForPublicHeader(packetNumber, leastUnacked)
					wirePacketNumber := (uint64(packetNumber) << (64 - length*8)) >> (64 - length*8)

					inferedPacketNumber := InferPacketNumber(length, leastUnacked, PacketNumber(wirePacketNumber))
					Expect(inferedPacketNumber).To(Equal(packetNumber))

					switch length {
					case PacketNumberLen2:
						increment = 1 << (2*8 - 3)
					case PacketNumberLen4:
						increment = 1 << (4*8 - 3)
					case PacketNumberLen6:
						increment = 1 << (6*8 - 3)
					}
				}
			})

			It("works for packet numbers larger than 2^48", func() {
				for i := (uint64(1) << 48); i < ((uint64(1) << 63) - 1); i += (uint64(1) << 48) {
					packetNumber := PacketNumber(i)
					leastUnacked := PacketNumber(i - 1000)
					length := GetPacketNumberLengthForPublicHeader(packetNumber, leastUnacked)
					wirePacketNumber := (uint64(packetNumber) << (64 - length*8)) >> (64 - length*8)

					inferedPacketNumber := InferPacketNumber(length, leastUnacked, PacketNumber(wirePacketNumber))
					Expect(inferedPacketNumber).To(Equal(packetNumber))
				}
			})
		})
	})

	Context("determining the minimum length of a packet number", func() {
		It("1 byte", func() {
			Expect(GetPacketNumberLength(0xFF)).To(Equal(PacketNumberLen1))
		})

		It("2 byte", func() {
			Expect(GetPacketNumberLength(0xFFFF)).To(Equal(PacketNumberLen2))
		})

		It("4 byte", func() {
			Expect(GetPacketNumberLength(0xFFFFFFFF)).To(Equal(PacketNumberLen4))
		})

		It("6 byte", func() {
			Expect(GetPacketNumberLength(0xFFFFFFFFFFFF)).To(Equal(PacketNumberLen6))
		})
	})
})
>>>>>>> project-faster/main
