package main

import (
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/p2p/enode"
)

func main() {
	// 这里替换为你的ENR字符串
	enrString := "enr:-KO4QB3lLMTcKsb_Y2yrupMRpGHxnE1SFsCJmZu7uUsAKxGhJWCIo0p5GCzGquohFkYxGT5wmycsL_wzXkkHWwoy8O2GAY1bD2Rhg2V0aMfGhJ89IlSAgmlkgnY0gmlwhKLcP8KJc2VjcDI1NmsxoQOykF6Y4KbOgbMhszrM4jpBuU8xB0-M81FFf_fJlruiuoRzbmFwwIN0Y3CCdl-DdWRwgnZf"

	// 解析ENR字符串
	node, err := enode.Parse(enode.ValidSchemes, enrString)
	if err != nil {
		log.Fatalf("Failed to parse ENR: %v", err)
	}

	// 打印出解析后的信息
	fmt.Println("Node ID:", node.ID())
	fmt.Println("IP Address:", node.IP())
	fmt.Println("UDP Port:", node.UDP())
	fmt.Println("TCP Port:", node.TCP())

}
