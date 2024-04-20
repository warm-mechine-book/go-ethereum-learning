package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// go run cmd/p2p/main.go
func main() {
	nodekey, _ := crypto.GenerateKey()
	srv := p2p.Server{
		Config: p2p.Config{
			PrivateKey: nodekey,
			Name:       "myethclient",
			ListenAddr: ":30303",
			MaxPeers:   10,
		},
	}

	ethProtocol := p2p.Protocol{
		Name:    "eth",
		Version: 63,
		Length:  17,
		Run:     handleEthProtocol,
	}

	srv.Protocols = append(srv.Protocols, ethProtocol)

	if err := srv.Start(); err != nil {
		fmt.Println("Failed to start server:", err)
		return
	}

	defer srv.Stop()
	fmt.Println("P2P Server started. Listening on", srv.ListenAddr)

	// 已知节点的 enode URL 列表
	enodeURLs := []string{
		"enode://b2905e98e0a6ce81b321b33acce23a41b94f31074f8cf351457ff7c996bba2ba16539103e1cd7f258c3f093882660e32ba2b773144e60914d93af7d7d930e1fb@162.220.63.194:30303",
		"enode://4c46bd49d0e9ea5ad9d76aa07bce469847f57c7ee39509d6d7699f0af01208b2d881a0956521ffff42b30f9c5156fe85fef5093e4d0694647190791f9783419f@119.131.14.103:30306",
		"enode://565ea37d856b122b8b82f15126d78cfb7b4b7113d0e613ac8fab2be5de2be558933b7fafafac49f8d22e2e1630364928e8243cdacfe9aad70e9249255b163675@124.79.157.6:40303",
		"enode://55dd85d21e76225e5cd7f49fe93e6aab3bb4557bc9631c0b481aec716a7b27f56eeda1b391a731a5b13372cf431037ad783f259e7fb43373547239f87412be57@27.25.122.202:30404",
	}

	// 解析每个 enode URL 并添加到服务器的对等体列表
	for _, url := range enodeURLs {
		node, err := enode.ParseV4(url)
		if err != nil {
			log.Printf("Failed to parse enode URL %s: %v", url, err)
			continue
		}
		srv.AddPeer(node)
		fmt.Println("Added node:", node)
	}

	fmt.Println("Node is running and connected to known nodes.")

	select {}
}

func handleEthProtocol(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	log.Println("Eth protocol handler running with peer:", peer)
	for {
		msg, err := ws.ReadMsg()
		if err != nil {
			return err
		}
		fmt.Println("msg:", msg)

		switch msg.Code {
		case eth.NewBlockHashesMsg:
			var mydata []eth.NewBlockHashesPacket
			if err := msg.Decode(&mydata); err != nil {
				return err
			}
			fmt.Println("New Block Hashes:", mydata)

		case eth.NewBlockMsg:
			var mydata eth.NewBlockPacket
			if err := msg.Decode(&mydata); err != nil {
				return err
			}
			fmt.Println("New Block:", mydata.Block.Number(), mydata.Block.Hash().Hex())
		}
	}
}
