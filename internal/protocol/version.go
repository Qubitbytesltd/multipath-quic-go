package protocol

<<<<<<< HEAD
import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	mrand "math/rand/v2"
	"sync"
)

// Version is a version number as int
type Version uint32

// gQUIC version range as defined in the wiki: https://github.com/quicwg/base-drafts/wiki/QUIC-Versions
const (
	gquicVersion0   = 0x51303030
	maxGquicVersion = 0x51303439
)

// The version numbers, making grepping easier
const (
	VersionUnknown Version = math.MaxUint32
	versionDraft29 Version = 0xff00001d // draft-29 used to be a widely deployed version
	Version1       Version = 0x1
	Version2       Version = 0x6b3343cf
=======
import "fmt"

// VersionNumber is a version number as int
type VersionNumber int

// The version numbers, making grepping easier
const (
	Version37 VersionNumber = 37 + iota
	Version38
	Version39
	VersionTLS         VersionNumber = 101
	VersionWhatever    VersionNumber = 0 // for when the version doesn't matter
	VersionUnsupported VersionNumber = -1
	VersionUnknown     VersionNumber = -2
	VersionMP          VersionNumber = 512
>>>>>>> project-faster/main
)

// SupportedVersions lists the versions that the server supports
// must be in sorted descending order
<<<<<<< HEAD
var SupportedVersions = []Version{Version1, Version2}

// IsValidVersion says if the version is known to quic-go
func IsValidVersion(v Version) bool {
	return v == Version1 || IsSupportedVersion(SupportedVersions, v)
}

func (vn Version) String() string {
	//nolint:exhaustive
	switch vn {
	case VersionUnknown:
		return "unknown"
	case versionDraft29:
		return "draft-29"
	case Version1:
		return "v1"
	case Version2:
		return "v2"
	default:
		if vn.isGQUIC() {
			return fmt.Sprintf("gQUIC %d", vn.toGQUICVersion())
		}
		return fmt.Sprintf("%#x", uint32(vn))
	}
}

func (vn Version) isGQUIC() bool {
	return vn > gquicVersion0 && vn <= maxGquicVersion
}

func (vn Version) toGQUICVersion() int {
	return int(10*(vn-gquicVersion0)/0x100) + int(vn%0x10)
}

// IsSupportedVersion returns true if the server supports this version
func IsSupportedVersion(supported []Version, v Version) bool {
=======
var SupportedVersions = []VersionNumber{
	VersionMP,
	Version39,
	Version38,
	Version37,
}

// UsesTLS says if this QUIC version uses TLS 1.3 for the handshake
func (vn VersionNumber) UsesTLS() bool {
	return vn == VersionTLS
}

func (vn VersionNumber) String() string {
	switch vn {
	case VersionWhatever:
		return "whatever"
	case VersionUnsupported:
		return "unsupported"
	case VersionUnknown:
		return "unknown"
	case VersionTLS:
		return "TLS dev version (WIP)"
	default:
		return fmt.Sprintf("%d", vn)
	}
}

// VersionNumberToTag maps version numbers ('32') to tags ('Q032')
func VersionNumberToTag(vn VersionNumber) uint32 {
	v := uint32(vn)
	return 'Q' + ((v/100%10)+'0')<<8 + ((v/10%10)+'0')<<16 + ((v%10)+'0')<<24
}

// VersionTagToNumber is built from VersionNumberToTag in init()
func VersionTagToNumber(v uint32) VersionNumber {
	return VersionNumber(((v>>8)&0xff-'0')*100 + ((v>>16)&0xff-'0')*10 + ((v>>24)&0xff - '0'))
}

// IsSupportedVersion returns true if the server supports this version
func IsSupportedVersion(supported []VersionNumber, v VersionNumber) bool {
>>>>>>> project-faster/main
	for _, t := range supported {
		if t == v {
			return true
		}
	}
	return false
}

// ChooseSupportedVersion finds the best version in the overlap of ours and theirs
// ours is a slice of versions that we support, sorted by our preference (descending)
<<<<<<< HEAD
// theirs is a slice of versions offered by the peer. The order does not matter.
// The bool returned indicates if a matching version was found.
func ChooseSupportedVersion(ours, theirs []Version) (Version, bool) {
	for _, ourVer := range ours {
		for _, theirVer := range theirs {
			if ourVer == theirVer {
				return ourVer, true
			}
		}
	}
	return 0, false
}

var (
	versionNegotiationMx   sync.Mutex
	versionNegotiationRand mrand.Rand
)

func init() {
	var seed [16]byte
	rand.Read(seed[:])
	versionNegotiationRand = *mrand.New(mrand.NewPCG(
		binary.BigEndian.Uint64(seed[:8]),
		binary.BigEndian.Uint64(seed[8:]),
	))
}

// generateReservedVersion generates a reserved version (v & 0x0f0f0f0f == 0x0a0a0a0a)
func generateReservedVersion() Version {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], versionNegotiationRand.Uint32())
	return Version((binary.BigEndian.Uint32(b[:]) | 0x0a0a0a0a) & 0xfafafafa)
}

// GetGreasedVersions adds one reserved version number to a slice of version numbers, at a random position.
// It doesn't modify the supported slice.
func GetGreasedVersions(supported []Version) []Version {
	versionNegotiationMx.Lock()
	defer versionNegotiationMx.Unlock()
	randPos := versionNegotiationRand.IntN(len(supported) + 1)
	greased := make([]Version, len(supported)+1)
	copy(greased, supported[:randPos])
	greased[randPos] = generateReservedVersion()
	copy(greased[randPos+1:], supported[randPos:])
	return greased
=======
// theirs is a slice of versions offered by the peer. The order does not matter
// if no suitable version is found, it returns VersionUnsupported
func ChooseSupportedVersion(ours, theirs []VersionNumber) VersionNumber {
	for _, ourVer := range ours {
		for _, theirVer := range theirs {
			if ourVer == theirVer {
				return ourVer
			}
		}
	}
	return VersionUnsupported
>>>>>>> project-faster/main
}
