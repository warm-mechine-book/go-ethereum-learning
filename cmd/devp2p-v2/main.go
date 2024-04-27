// Copyright 2019 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
)

func parseNode(source string) (*enode.Node, error) {
	if strings.HasPrefix(source, "enode://") {
		return enode.ParseV4(source)
	}
	r, err := parseRecord(source)
	if err != nil {
		return nil, err
	}
	return enode.New(enode.ValidSchemes, r)
}

func parseRecord(source string) (*enr.Record, error) {
	bin := []byte(source)
	if d, ok := decodeRecordHex(bytes.TrimSpace(bin)); ok {
		bin = d
	} else if d, ok := decodeRecordBase64(bytes.TrimSpace(bin)); ok {
		bin = d
	}
	var r enr.Record
	err := rlp.DecodeBytes(bin, &r)
	return &r, err
}

func decodeRecordHex(b []byte) ([]byte, bool) {
	if bytes.HasPrefix(b, []byte("0x")) {
		b = b[2:]
	}
	dec := make([]byte, hex.DecodedLen(len(b)))
	_, err := hex.Decode(dec, b)
	return dec, err == nil
}

func decodeRecordBase64(b []byte) ([]byte, bool) {
	if bytes.HasPrefix(b, []byte("enr:")) {
		b = b[4:]
	}
	dec := make([]byte, base64.RawURLEncoding.DecodedLen(len(b)))
	n, err := base64.RawURLEncoding.Decode(dec, b)
	return dec[:n], err == nil
}

func main() {
	fmt.Println(os.Args)
	nodeInput := "enode://c9d9a8656916a6303e401be2e127ef6054fc3a1f74408593d9cbdb319370c5b13ee98b0d9ef6b7f22a45bec50598a696aa4770cbb9f1109e6ef82ed4d4bea26c@208.91.107.69:30303"
	node, err := parseNode(nodeInput)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(node.ID)
	fmt.Println(node.Pubkey())
	fmt.Println(node.IP())
	fmt.Println(node.TCP())
	fmt.Println(node.UDP())

	// exit(app.Run(os.Args))
}

// // commandHasFlag returns true if the current command supports the given flag.
// func commandHasFlag(ctx *cli.Context, flag cli.Flag) bool {
// 	names := flag.Names()
// 	set := make(map[string]struct{}, len(names))
// 	for _, name := range names {
// 		set[name] = struct{}{}
// 	}
// 	for _, ctx := range ctx.Lineage() {
// 		if ctx.Command != nil {
// 			for _, f := range ctx.Command.Flags {
// 				for _, name := range f.Names() {
// 					if _, ok := set[name]; ok {
// 						return true
// 					}
// 				}
// 			}
// 		}
// 	}
// 	return false
// }

// // getNodeArg handles the common case of a single node descriptor argument.
// func getNodeArg(ctx *cli.Context) *enode.Node {
// 	if ctx.NArg() < 1 {
// 		exit("missing node as command-line argument")
// 	}
// 	n, err := parseNode(ctx.Args().First())
// 	if err != nil {
// 		exit(err)
// 	}
// 	return n
// }

// func exit(err interface{}) {
// 	if err == nil {
// 		os.Exit(0)
// 	}
// 	fmt.Fprintln(os.Stderr, err)
// 	os.Exit(1)
// }
