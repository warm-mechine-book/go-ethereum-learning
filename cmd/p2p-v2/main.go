package main

import (
	"fmt"
	// "log"
	"os"
	"os/signal"
	"syscall"

	// "github.com/ava-labs/coreth/core/types"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/nat"
)

func main() {
	logger := log.New()

	// 创建事件通道
	peerEventCh := make(chan *p2p.PeerEvent, 10)

	// 生成或加载你的私钥
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		logger.Error("Failed to generate private key: %v", err)
	}

	// 设置节点信息
	cfg := p2p.Config{
		PrivateKey:      privateKey,
		Name:            "my-eth-node",
		Protocols:       []p2p.Protocol{makeEthProtocol()},
		MaxPeers:        10,
		NAT:             nat.Any(),
		ListenAddr:      ":30303",
		EnableMsgEvents: true,
		Logger:          logger,
	}

	// 创建节点
	srv := &p2p.Server{
		Config: cfg,
	}

	sub := srv.SubscribeEvents(peerEventCh)
	defer sub.Unsubscribe()

	// 启动服务器

	if err := srv.Start(); err != nil {
		logger.Error("Unable to start server: %v", err)
	} else {
		logger.Info("Server started successfully")
	}
	defer srv.Stop()

	// 创建节点连接
	nodeURL := "enode://b2905e98e0a6ce81b321b33acce23a41b94f31074f8cf351457ff7c996bba2ba16539103e1cd7f258c3f093882660e32ba2b773144e60914d93af7d7d930e1fb@162.220.63.194:30303"
	remoteNode, err := enode.Parse(enode.ValidSchemes, nodeURL)
	if err != nil {
		logger.Error("Could not parse remote node URL: %v", err)
	}

	// 添加对等节点
	srv.AddPeer(remoteNode)

	// 处理系统中断，以优雅地停止服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 事件处理循环
	go func() {
		for {
			select {
			case event := <-peerEventCh:
				handlePeerEvent(event)
			case <-quit:
				return
			}
		}
	}()

	<-quit
	fmt.Println("Received shutdown signal")
}

func makeEthProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    "eth",
		Version: 66,
		Length:  17,
		Run:     handlePeer,
	}
}

func handlePeer(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
	fmt.Println("Peer connected:", peer.ID())

	request := &eth.GetBlockHeadersRequest{
		Origin:  eth.HashOrNumber{Number: 1920000}, // Specify either Number or Hash
		Amount:  10,
		Skip:    0,
		Reverse: false,
	}

	// 发送 GetBlockHeaders 消息
	packet := eth.GetBlockHeadersPacket{
		RequestId:              1, // Unique request ID
		GetBlockHeadersRequest: request,
	}
	err := p2p.Send(rw, eth.GetBlockHeadersMsg, packet)
	if err != nil {
		return fmt.Errorf("could not send block header request: %v", err)
	}

	// 接收响应
	msg, err := rw.ReadMsg()
	fmt.Println("Received block headers message", msg)
	if err != nil {
		return fmt.Errorf("failed to read message: %v", err)
	}

	if msg.Code == eth.BlockHeadersMsg {
		fmt.Println("BlockHeadersMsg")
		// var headers []*eth.Header
		// if err := msg.Decode(&headers); err != nil {
		// 	return fmt.Errorf("failed to decode headers: %v", err)
		// }
		// for _, header := range headers {
		// 	fmt.Println("Block number:", header.Number)
		// }
	}

	return nil
}

// handlePeerEvent 处理节点事件
func handlePeerEvent(event *p2p.PeerEvent) {
	switch event.Type {
	case p2p.PeerEventTypeAdd:
		fmt.Printf("Peer added: %s\n", event.Peer)
	case p2p.PeerEventTypeDrop:
		fmt.Printf("Peer dropped: %s\n", event.Peer)
	case p2p.PeerEventTypeMsgRecv:
		fmt.Printf("Message received from: %s\n", event.Peer)
	case p2p.PeerEventTypeMsgSend:
		fmt.Printf("Message sent to: %s\n", event.Peer)
	}
}
