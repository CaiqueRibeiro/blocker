package node

import (
	"context"
	"encoding/hex"
	"net"
	"sync"
	"time"

	"github.com/CaiqueRibeiro/blocker/crypto"
	"github.com/CaiqueRibeiro/blocker/proto"
	"github.com/CaiqueRibeiro/blocker/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

const BLOCK_TIME = 5 * time.Second

type Mempool struct {
	txx map[string]*proto.Transaction
}

func NewMemPool() *Mempool {
	return &Mempool{txx: make(map[string]*proto.Transaction)}
}

func (m *Mempool) Has(tx *proto.Transaction) bool {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	_, ok := m.txx[hash]
	return ok
}

func (m *Mempool) Add(tx *proto.Transaction) bool {
	if m.Has(tx) {
		return false
	}
	hash := hex.EncodeToString(types.HashTransaction(tx))
	m.txx[hash] = tx
	return true
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	ServerConfig
	logger   *zap.SugaredLogger
	peerLock sync.RWMutex
	peers    map[proto.NodeClient]*proto.Version
	mempool  *Mempool

	proto.UnimplementedNodeServer
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(c), nil
}

func NewNode(cfg ServerConfig) *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()

	return &Node{
		peers:        make(map[proto.NodeClient]*proto.Version),
		logger:       logger.Sugar(),
		mempool:      NewMemPool(),
		ServerConfig: cfg,
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.ListenAddr = listenAddr
	options := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(options...)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.ListenAddr)
	if len(bootstrapNodes) > 0 { // if there are bootstrap nodes
		go n.bootstrapNetwork(bootstrapNodes) // connect with node addresses informed in startup
	}
	if n.PrivateKey != nil {
		go n.validatorLoop()
	}
	return grpcServer.Serve(ln)
}

// receives a connection from an external node, returns own version and add the node to peer list
func (n *Node) Handshake(ctx context.Context, v *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(v.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, v)            // add the receiving node to the list of connected peers (two-way connection)
	return n.getVersion(), nil // returns own version to receiving node to be added in its list of connected peers
}

func (n *Node) HandleTransaction(ctx context.Context, tx *proto.Transaction) (*proto.Ack, error) {
	peer, _ := peer.FromContext(ctx)
	hash := hex.EncodeToString(types.HashTransaction(tx))
	if n.mempool.Add(tx) {
		n.logger.Debugw("received tx", "from", peer.Addr, "hash", hash, "we", n.ListenAddr)
		go func() {
			if err := n.broadcast(tx); err != nil {
				n.logger.Errorw("broadcast error", "err", err)
			}
		}()
	}
	return &proto.Ack{}, nil
}

func (n *Node) validatorLoop() {
	n.logger.Infow("starting validator loop", "pubkey", n.PrivateKey.Public(), "blockTime", BLOCK_TIME)
	ticket := time.NewTicker(BLOCK_TIME)
	for {
		<-ticket.C
		n.logger.Debugw("time to create a new block", "lenTx", len(n.mempool.txx))
		for hash := range n.mempool.txx {
			delete(n.mempool.txx, hash)
		}
	}
}

// Loop through all connected peers and broadcast the message to each one
func (n *Node) broadcast(msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			_, err := peer.HandleTransaction(context.Background(), v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// handshakes with a list of other node addresses and add it in own list of connected peers
func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) { // verify if candidate to connection is able to be connected
			continue
		}
		n.logger.Debugw("dialing remote node", "we", n.ListenAddr, "remote", addr)
		c, v, err := n.dialRemoteWork(addr)
		if err != nil {
			return err
		}
		n.addPeer(c, v) // adds the node to the list of connected peers
	}
	return nil
}

// makes handshake with a single address and returns client/version to be added in node peer
func (n *Node) dialRemoteWork(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr) // connects to an external node address
	if err != nil {
		return nil, nil, err
	}
	v, err := c.Handshake(context.Background(), n.getVersion()) // sends own version to another node an receives its version from it
	if err != nil {
		return nil, nil, err
	}
	return c, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "blocker-0.1",
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

/*
verify if candidate to connection is able to be connected

There are two cases in which a connection is not possible:

  - if the candidate address is the same as the node's own address
  - if the candidate address is already in the list of connected peers
*/
func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}
	connectedPeers := n.getPeerList()
	for _, connectedAddr := range connectedPeers {
		if addr == connectedAddr {
			return false
		}
	}
	return true
}

func (n *Node) getPeerList() []string {
	n.peerLock.RLock()
	defer n.peerLock.RUnlock()
	peers := []string{}
	for _, version := range n.peers {
		peers = append(peers, version.ListenAddr)
	}
	return peers
}

/*
Gets the client and version of an external node and add it to the list of connected peers.

A node cannot connect with itself nor with a node that is already in the list of connected peers
*/
func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	n.peers[c] = v
	// connect to all peers in the received list of peer from other node
	if len(v.PeerList) > 0 {
		go n.bootstrapNetwork(v.PeerList)
	}
	n.logger.Debugw("new peer connected",
		"we", n.ListenAddr,
		"remoteNode", v.ListenAddr,
		"height", v.Height)
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}
