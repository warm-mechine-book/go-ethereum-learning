package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	// "github.com/ava-labs/coreth/core/types"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

func main() {
	// 生成或加载你的私钥
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// 设置节点信息
	cfg := p2p.Config{
		PrivateKey:      privateKey,
		Name:            "my-eth-node",
		Protocols:       []p2p.Protocol{makeEthProtocol()},
		MaxPeers:        10,
		EnableMsgEvents: true,
	}

	// 创建节点
	srv := &p2p.Server{
		Config: cfg,
	}

	// 启动服务器

	if err := srv.Start(); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	} else {
		log.Println("Server started successfully")
	}
	defer srv.Stop()

	// 创建节点连接
	nodeURL := "enode://b2905e98e0a6ce81b321b33acce23a41b94f31074f8cf351457ff7c996bba2ba16539103e1cd7f258c3f093882660e32ba2b773144e60914d93af7d7d930e1fb@162.220.63.194:30303"
	remoteNode, err := enode.Parse(enode.ValidSchemes, nodeURL)
	if err != nil {
		log.Fatalf("Could not parse remote node URL: %v", err)
	}

	// 添加对等节点
	srv.AddPeer(remoteNode)

	// 处理系统中断，以优雅地停止服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
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
	log.Println("Peer connected:", peer.ID())

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
