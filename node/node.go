package node

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/CaiqueRibeiro/blocker/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type Node struct {
	logger     *zap.SugaredLogger
	version    string
	listenAddr string
	peerLock   sync.RWMutex
	peers      map[proto.NodeClient]*proto.Version
	proto.UnimplementedNodeServer
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	c, err := grpc.Dial(listenAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewNodeClient(c), nil
}

func NewNode() *Node {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	logger, _ := loggerConfig.Build()

	return &Node{
		peers:   make(map[proto.NodeClient]*proto.Version),
		version: "blocker-0.1",
		logger:  logger.Sugar(),
	}
}

func (n *Node) Start(listenAddr string, bootstrapNodes []string) error {
	n.listenAddr = listenAddr
	options := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(options...)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.listenAddr)
	if len(bootstrapNodes) > 0 { // if there are bootstrap nodes
		go n.bootstrapNetwork(bootstrapNodes) // connect with node addresses informed in startup
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
	fmt.Println("received tx from:", peer)
	return &proto.Ack{}, nil
}

// handshakes with a list of other node addresses and add it in own list of connected peers
func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) { // verify if candidate to connection is able to be connected
			continue
		}
		n.logger.Debugw("dialing remote node", "we", n.listenAddr, "remote", addr)
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
		ListenAddr: n.listenAddr,
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
	if n.listenAddr == addr {
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
		"we", n.listenAddr,
		"remoteNode", v.ListenAddr,
		"height", v.Height)
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}
