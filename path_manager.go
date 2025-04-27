package quic

import (
<<<<<<< HEAD
	"crypto/rand"
	"net"
	"slices"
	"time"

	"github.com/quic-go/quic-go/internal/ackhandler"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/utils"
	"github.com/quic-go/quic-go/internal/wire"
)

type pathID int64

const invalidPathID pathID = -1

// Maximum number of paths to keep track of.
// If the peer probes another path (before the pathTimeout of an existing path expires),
// this probing attempt is ignored.
const maxPaths = 3

// If no packet is received for a path for pathTimeout,
// the path can be evicted when the peer probes another path.
// This prevents an attacker from churning through paths by duplicating packets and
// sending them with spoofed source addresses.
const pathTimeout = 5 * time.Second

type path struct {
	id             pathID
	addr           net.Addr
	lastPacketTime time.Time
	pathChallenge  [8]byte
	validated      bool
	rcvdNonProbing bool
}

type pathManager struct {
	nextPathID pathID
	// ordered by lastPacketTime, with the most recently used path at the end
	paths []*path

	getConnID    func(pathID) (_ protocol.ConnectionID, ok bool)
	retireConnID func(pathID)

	logger utils.Logger
}

func newPathManager(
	getConnID func(pathID) (_ protocol.ConnectionID, ok bool),
	retireConnID func(pathID),
	logger utils.Logger,
) *pathManager {
	return &pathManager{
		paths:        make([]*path, 0, maxPaths+1),
		getConnID:    getConnID,
		retireConnID: retireConnID,
		logger:       logger,
	}
}

// Returns a path challenge frame if one should be sent.
// May return nil.
func (pm *pathManager) HandlePacket(
	remoteAddr net.Addr,
	t time.Time,
	pathChallenge *wire.PathChallengeFrame, // may be nil if the packet didn't contain a PATH_CHALLENGE
	isNonProbing bool,
) (_ protocol.ConnectionID, _ []ackhandler.Frame, shouldSwitch bool) {
	var p *path
	for i, path := range pm.paths {
		if addrsEqual(path.addr, remoteAddr) {
			p = path
			p.lastPacketTime = t
			// already sent a PATH_CHALLENGE for this path
			if isNonProbing {
				path.rcvdNonProbing = true
			}
			if pm.logger.Debug() {
				pm.logger.Debugf("received packet for path %s that was already probed, validated: %t", remoteAddr, path.validated)
			}
			shouldSwitch = path.validated && path.rcvdNonProbing
			if i != len(pm.paths)-1 {
				// move the path to the end of the list
				pm.paths = slices.Delete(pm.paths, i, i+1)
				pm.paths = append(pm.paths, p)
			}
			if pathChallenge == nil {
				return protocol.ConnectionID{}, nil, shouldSwitch
			}
		}
	}

	if len(pm.paths) >= maxPaths {
		if pm.paths[0].lastPacketTime.Add(pathTimeout).After(t) {
			if pm.logger.Debug() {
				pm.logger.Debugf("received packet for previously unseen path %s, but already have %d paths", remoteAddr, len(pm.paths))
			}
			return protocol.ConnectionID{}, nil, shouldSwitch
		}
		// evict the oldest path, if the last packet was received more than pathTimeout ago
		pm.retireConnID(pm.paths[0].id)
		pm.paths = pm.paths[1:]
	}

	var pathID pathID
	if p != nil {
		pathID = p.id
	} else {
		pathID = pm.nextPathID
	}

	// previously unseen path, initiate path validation by sending a PATH_CHALLENGE
	connID, ok := pm.getConnID(pathID)
	if !ok {
		pm.logger.Debugf("skipping validation of new path %s since no connection ID is available", remoteAddr)
		return protocol.ConnectionID{}, nil, shouldSwitch
	}

	frames := make([]ackhandler.Frame, 0, 2)
	if p == nil {
		var pathChallengeData [8]byte
		rand.Read(pathChallengeData[:])
		p = &path{
			id:             pm.nextPathID,
			addr:           remoteAddr,
			lastPacketTime: t,
			rcvdNonProbing: isNonProbing,
			pathChallenge:  pathChallengeData,
		}
		pm.nextPathID++
		pm.paths = append(pm.paths, p)
		frames = append(frames, ackhandler.Frame{
			Frame:   &wire.PathChallengeFrame{Data: p.pathChallenge},
			Handler: (*pathManagerAckHandler)(pm),
		})
		pm.logger.Debugf("enqueueing PATH_CHALLENGE for new path %s", remoteAddr)
	}
	if pathChallenge != nil {
		frames = append(frames, ackhandler.Frame{
			Frame:   &wire.PathResponseFrame{Data: pathChallenge.Data},
			Handler: (*pathManagerAckHandler)(pm),
		})
	}
	return connID, frames, shouldSwitch
}

func (pm *pathManager) HandlePathResponseFrame(f *wire.PathResponseFrame) {
	for _, p := range pm.paths {
		if f.Data == p.pathChallenge {
			// path validated
			p.validated = true
			pm.logger.Debugf("path %s validated", p.addr)
			break
		}
	}
}

// SwitchToPath is called when the connection switches to a new path
func (pm *pathManager) SwitchToPath(addr net.Addr) {
	// retire all other paths
	for _, path := range pm.paths {
		if addrsEqual(path.addr, addr) {
			pm.logger.Debugf("switching to path %d (%s)", path.id, addr)
			continue
		}
		pm.retireConnID(path.id)
	}
	clear(pm.paths)
	pm.paths = pm.paths[:0]
}

type pathManagerAckHandler pathManager

var _ ackhandler.FrameHandler = &pathManagerAckHandler{}

// Acknowledging the frame doesn't validate the path, only receiving the PATH_RESPONSE does.
func (pm *pathManagerAckHandler) OnAcked(f wire.Frame) {}

func (pm *pathManagerAckHandler) OnLost(f wire.Frame) {
	pc, ok := f.(*wire.PathChallengeFrame)
	if !ok {
		return
	}
	for i, path := range pm.paths {
		if path.pathChallenge == pc.Data {
			pm.paths = slices.Delete(pm.paths, i, i+1)
			pm.retireConnID(path.id)
			break
=======
	"errors"
	"net"
	"time"

	"github.com/project-faster/mp-quic-go/congestion"
	"github.com/project-faster/mp-quic-go/internal/protocol"
	"github.com/project-faster/mp-quic-go/internal/utils"
	"github.com/project-faster/mp-quic-go/internal/wire"
)

type pathManager struct {
	pconnMgr  *pconnManager
	sess      *session
	nxtPathID protocol.PathID
	// Number of paths, excluding the initial one
	nbPaths uint8

	remoteAddrs4 []net.UDPAddr
	remoteAddrs6 []net.UDPAddr

	advertisedLocAddrs map[string]bool

	// TODO (QDC): find a cleaner way
	oliaSenders map[protocol.PathID]*congestion.OliaSender

	handshakeCompleted chan struct{}
	runClosed          chan struct{}
	timer              *time.Timer
}

func (pm *pathManager) setup(conn connection) {
	// Initial PathID is 0
	// PathIDs of client-initiated paths are even
	// those of server-initiated paths odd
	if pm.sess.perspective == protocol.PerspectiveClient {
		pm.nxtPathID = 1
	} else {
		pm.nxtPathID = 2
	}

	pm.remoteAddrs4 = make([]net.UDPAddr, 0)
	pm.remoteAddrs6 = make([]net.UDPAddr, 0)
	pm.advertisedLocAddrs = make(map[string]bool)
	pm.handshakeCompleted = make(chan struct{}, 1)
	pm.runClosed = make(chan struct{}, 1)
	pm.timer = time.NewTimer(0)
	pm.nbPaths = 0

	pm.oliaSenders = make(map[protocol.PathID]*congestion.OliaSender)

	// Setup the first path of the connection
	pm.sess.paths[protocol.InitialPathID] = &path{
		pathID: protocol.InitialPathID,
		sess:   pm.sess,
		conn:   conn,
	}

	// Setup this first path
	pm.sess.paths[protocol.InitialPathID].setup(pm.oliaSenders)

	// With the initial path, get the remoteAddr to create paths accordingly
	if conn.RemoteAddr() != nil {
		remAddr, err := net.ResolveUDPAddr("udp", conn.RemoteAddr().String())
		if err != nil {
			utils.Errorf("path manager: encountered error while parsing remote addr: %v", remAddr)
		}

		if remAddr.IP.To4() != nil {
			pm.remoteAddrs4 = append(pm.remoteAddrs4, *remAddr)
		} else {
			pm.remoteAddrs6 = append(pm.remoteAddrs6, *remAddr)
		}
	}

	// Launch the path manager
	go pm.run()
}

func (pm *pathManager) run() {
	// Close immediately if requested
	select {
	case <-pm.runClosed:
		return
	case <-pm.handshakeCompleted:
		if pm.sess.createPaths {
			err := pm.createPaths()
			if err != nil {
				pm.closePaths()
				return
			}
		}
	}

runLoop:
	for {
		select {
		case <-pm.runClosed:
			break runLoop
		case <-pm.pconnMgr.changePaths:
			if pm.sess.createPaths {
				pm.createPaths()
			}
		}
	}
	// Close paths
	pm.closePaths()
}

func getIPVersion(ip net.IP) int {
	if ip.To4() != nil {
		return 4
	}
	return 6
}

func (pm *pathManager) advertiseAddresses() {
	pm.pconnMgr.mutex.Lock()
	defer pm.pconnMgr.mutex.Unlock()
	for _, locAddr := range pm.pconnMgr.localConns {
		_, sent := pm.advertisedLocAddrs[locAddr.String()]
		if !sent {
			version := getIPVersion(locAddr.IP)
			pm.sess.streamFramer.AddAddressForTransmission(uint8(version), locAddr)
			pm.advertisedLocAddrs[locAddr.String()] = true
>>>>>>> project-faster/main
		}
	}
}

<<<<<<< HEAD
func addrsEqual(addr1, addr2 net.Addr) bool {
	if addr1 == nil || addr2 == nil {
		return false
	}
	a1, ok1 := addr1.(*net.UDPAddr)
	a2, ok2 := addr2.(*net.UDPAddr)
	if ok1 && ok2 {
		return a1.IP.Equal(a2.IP) && a1.Port == a2.Port
	}
	return addr1.String() == addr2.String()
=======
func (pm *pathManager) createPath(locAddr net.UDPAddr, remAddr net.UDPAddr) error {
	// First check that the path does not exist yet
	pm.sess.pathsLock.Lock()
	defer pm.sess.pathsLock.Unlock()
	paths := pm.sess.paths
	for _, pth := range paths {
		locAddrPath := pth.conn.LocalAddr().String()
		remAddrPath := pth.conn.RemoteAddr().String()
		if locAddr.String() == locAddrPath && remAddr.String() == remAddrPath {
			// Path already exists, so don't create it again
			return nil
		}
	}
	// No matching path, so create it
	pth := &path{
		pathID: pm.nxtPathID,
		sess:   pm.sess,
		conn:   &conn{pconn: pm.pconnMgr.pconns[locAddr.String()], currentAddr: &remAddr},
	}
	pth.setup(pm.oliaSenders)
	pm.sess.paths[pm.nxtPathID] = pth
	if utils.Debug() {
		utils.Debugf("Created path %x on %s to %s", pm.nxtPathID, locAddr.String(), remAddr.String())
	}
	pm.nxtPathID += 2
	// Send a PING frame to get latency info about the new path and informing the
	// peer of its existence
	// Because we hold pathsLock, it is safe to send packet now
	return pm.sess.sendPing(pth)
}

func (pm *pathManager) createPaths() error {
	if utils.Debug() {
		utils.Debugf("Path manager tries to create paths")
	}

	// XXX (QDC): don't let the server create paths for now
	if pm.sess.perspective == protocol.PerspectiveServer {
		pm.advertiseAddresses()
		return nil
	}
	// TODO (QDC): clearly not optimali
	pm.pconnMgr.mutex.Lock()
	defer pm.pconnMgr.mutex.Unlock()
	for _, locAddr := range pm.pconnMgr.localConns {
		version := getIPVersion(locAddr.IP)
		if version == 4 {
			for _, remAddr := range pm.remoteAddrs4 {
				err := pm.createPath(locAddr, remAddr)
				if err != nil {
					return err
				}
			}
		} else {
			for _, remAddr := range pm.remoteAddrs6 {
				err := pm.createPath(locAddr, remAddr)
				if err != nil {
					return err
				}
			}
		}
	}
	pm.sess.schedulePathsFrame()
	return nil
}

func (pm *pathManager) createPathFromRemote(p *receivedPacket) (*path, error) {
	pm.sess.pathsLock.Lock()
	defer pm.sess.pathsLock.Unlock()
	localPconn := p.rcvPconn
	remoteAddr := p.remoteAddr
	pathID := p.publicHeader.PathID

	// Sanity check: pathID should not exist yet
	_, ko := pm.sess.paths[pathID]
	if ko {
		return nil, errors.New("trying to create already existing path")
	}

	// Sanity check: odd is client initiated, even for server initiated
	if pm.sess.perspective == protocol.PerspectiveClient && pathID%2 != 0 {
		return nil, errors.New("server tries to create odd pathID")
	}
	if pm.sess.perspective == protocol.PerspectiveServer && pathID%2 == 0 {
		return nil, errors.New("client tries to create even pathID")
	}

	pth := &path{
		pathID: pathID,
		sess:   pm.sess,
		conn:   &conn{pconn: localPconn, currentAddr: remoteAddr},
	}

	pth.setup(pm.oliaSenders)
	pm.sess.paths[pathID] = pth

	if utils.Debug() {
		utils.Debugf("Created remote path %x on %s to %s", pathID, localPconn.LocalAddr().String(), remoteAddr.String())
	}

	return pth, nil
}

func (pm *pathManager) handleAddAddressFrame(f *wire.AddAddressFrame) error {
	switch f.IPVersion {
	case 4:
		pm.remoteAddrs4 = append(pm.remoteAddrs4, f.Addr)
	case 6:
		pm.remoteAddrs6 = append(pm.remoteAddrs6, f.Addr)
	default:
		return wire.ErrUnknownIPVersion
	}
	if pm.sess.createPaths {
		return pm.createPaths()
	}
	return nil
}

func (pm *pathManager) closePath(pthID protocol.PathID) error {
	pm.sess.pathsLock.RLock()
	defer pm.sess.pathsLock.RUnlock()

	pth, ok := pm.sess.paths[pthID]
	if !ok {
		// XXX (QDC) Unknown path, what should we do?
		return nil
	}

	if pth.open.Get() {
		pth.closeChan <- nil
	}

	return nil
}

func (pm *pathManager) closePaths() {
	pm.sess.pathsLock.RLock()
	paths := pm.sess.paths
	for _, pth := range paths {
		if pth.open.Get() {
			select {
			case pth.closeChan <- nil:
			default:
				// Don't remain stuck here!
			}
		}
	}
	pm.sess.pathsLock.RUnlock()
>>>>>>> project-faster/main
}
