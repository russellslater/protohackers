package main

import (
	"log"

	"github.com/russellslater/protohackers/cmd/mob-in-the-middle/boguscoin"
	"github.com/russellslater/protohackers/cmd/mob-in-the-middle/chatproxy"
)

const (
	protohackersChatSvrAddr = "chat.protohackers.com:16963"
)

func main() {
	proxySrv := chatproxy.NewChatProxy(5000, protohackersChatSvrAddr)
	proxySrv.Rewriters = []chatproxy.Rewriter{boguscoin.NewBoguscoinAddrRewriter()}
	log.Fatal(proxySrv.Start())
}
