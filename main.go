package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":9200", "监听地址")
	difficulty := flag.Int("diff", 3, "PoW 难度（哈希前缀 0 的个数）")
	flag.Parse()

	bc := newBlockchain(*difficulty)
	bc.Register(http.DefaultServeMux)

	log.Printf("miniblock 启动 %s，PoW 难度 %d", *addr, *difficulty)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
