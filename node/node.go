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

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    "blocker-0.1",
		Height:     0,
		ListenAddr: n.listenAddr,
	}
}

func (n *Node) addPeer(c proto.NodeClient, v *proto.Version) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	n.logger.Debugw("new peer connected", "addr", v.ListenAddr, "height", v.Height)
	n.peers[c] = v
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.peerLock.Lock()
	defer n.peerLock.Unlock()
	delete(n.peers, c)
}

func (n *Node) BootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		c, err := makeNodeClient(addr) // connects to an external node address
		if err != nil {
			return err
		}
		v, err := c.Handshake(context.Background(), n.getVersion()) // sends own version to another node an receives its version from it
		if err != nil {
			n.logger.Error("handshake error", err)
			continue
		}
		n.addPeer(c, v) // adds the node to the list of connected peers
	}
	return nil
}

func (n *Node) Start(listenAddr string) error {
	n.listenAddr = listenAddr
	options := []grpc.ServerOption{}
	grpcServer := grpc.NewServer(options...)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	proto.RegisterNodeServer(grpcServer, n)
	n.logger.Infow("node started...", "port", n.listenAddr)
	return grpcServer.Serve(ln)
}

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
