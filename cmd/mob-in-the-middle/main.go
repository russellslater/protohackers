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
	proxySvr := chatproxy.NewChatProxy(5000, protohackersChatSvrAddr)
	proxySvr.Rewriters = []chatproxy.Rewriter{boguscoin.NewBoguscoinAddrRewriter()}
	log.Fatal(proxySvr.Start())
}
