package main

import (
	"crypto/sha256"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Block 是区块链里的一个区块。
type Block struct {
	Index     int64  `json:"index"`
	Timestamp int64  `json:"timestamp"`
	PrevHash  string `json:"prev_hash"`
	Hash      string `json:"hash"`
	Nonce     int64  `json:"nonce"`
	// TxRoot 是这批交易的默克尔根（简化版直接算所有交易拼接的哈希）。
	TxRoot string        `json:"tx_root"`
	Txs    []Transaction `json:"txs"`
}

// Transaction 是一笔交易，这里就记转账：from 给 to 转 amount。
type Transaction struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
	// 签名在真链上才有，这里用个简单字段占位，方便校验时确认不是空交易。
	Nonce int64 `json:"nonce"`
}

// calcHash 把区块头的关键字段拼起来算 sha256，PoW 就是找让这个哈希前缀够多 0 的 nonce。
func (b *Block) calcHash() string {
	var buf strings.Builder
	buf.WriteString(strconv.FormatInt(b.Index, 10))
	buf.WriteString(strconv.FormatInt(b.Timestamp, 10))
	buf.WriteString(b.PrevHash)
	buf.WriteString(b.TxRoot)
	buf.WriteString(strconv.FormatInt(b.Nonce, 10))
	sum := sha256.Sum256([]byte(buf.String()))
	return hexEncode(sum[:])
}

// txRoot 把所有交易按序拼接再哈希，作为这批交易的摘要。
func txRoot(txs []Transaction) string {
	var buf strings.Builder
	for _, t := range txs {
		b, _ := json.Marshal(t)
		buf.Write(b)
	}
	sum := sha256.Sum256([]byte(buf.String()))
	return hexEncode(sum[:])
}

// mine 做工作量证明：不断加 nonce 直到哈希前缀有 difficulty 个 0。
func mine(index int64, prevHash string, txs []Transaction, difficulty int) Block {
	root := txRoot(txs)
	block := Block{
		Index:     index,
		Timestamp: time.Now().Unix(),
		PrevHash:  prevHash,
		TxRoot:    root,
		Txs:       txs,
	}
	prefix := strings.Repeat("0", difficulty)
	for {
		h := block.calcHash()
		if strings.HasPrefix(h, prefix) {
			block.Hash = h
			return block
		}
		block.Nonce++
	}
}

// hexEncode 手写二进制转十六进制，避免再引包。
func hexEncode(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = digits[v>>4]
		out[i*2+1] = digits[v&0x0f]
	}
	return string(out)
}
