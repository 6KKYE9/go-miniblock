package main

import (
	"errors"
	"strings"
	"sync"
)

// ErrInvalidChain 表示链在校验时发现某个区块的哈希或前驱对不上。
var ErrInvalidChain = errors.New("区块链校验失败")

// Blockchain 管理整条链和等待打包的交易池。
type Blockchain struct {
	mu         sync.Mutex
	chain      []Block
	pending    []Transaction
	difficulty int
}

func newBlockchain(difficulty int) *Blockchain {
	if difficulty <= 0 {
		difficulty = 3
	}
	bc := &Blockchain{difficulty: difficulty}
	// 创世区块：没有前驱，prev 填空，带一笔 coinbase 奖励交易。
	genesis := mine(0, "", []Transaction{{From: "coinbase", To: "miner", Amount: 50}}, difficulty)
	bc.chain = append(bc.chain, genesis)
	return bc
}

// AddTx 把一笔交易放进待打包池。做最基本的字段校验。
func (bc *Blockchain) AddTx(t Transaction) error {
	if t.To == "" || t.Amount <= 0 {
		return errors.New("收款人不能为空，金额必须为正")
	}
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.pending = append(bc.pending, t)
	return nil
}

// Mine 把待打包的交易挖成一个新区块，接到链上，返回新块。
func (bc *Blockchain) Mine() Block {
	bc.mu.Lock()
	txs := bc.pending
	prev := bc.chain[len(bc.chain)-1]
	bc.mu.Unlock()

	// 真链这里会校验每笔交易的余额和签名，练手版跳过，只打包。
	newBlock := mine(prev.Index+1, prev.Hash, txs, bc.difficulty)

	bc.mu.Lock()
	bc.chain = append(bc.chain, newBlock)
	bc.pending = nil
	bc.mu.Unlock()
	return newBlock
}

func (bc *Blockchain) Chain() []Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	out := make([]Block, len(bc.chain))
	copy(out, bc.chain)
	return out
}

func (bc *Blockchain) Pending() []Transaction {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	out := make([]Transaction, len(bc.pending))
	copy(out, bc.pending)
	return out
}

// Validate 顺着链检查：每个块的哈希自洽，且 prev_hash 指向前一个块的哈希。
func (bc *Blockchain) Validate() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	var prevHash string
	for i, b := range bc.chain {
		if b.calcHash() != b.Hash {
			return ErrInvalidChain
		}
		// 重算交易根和存的对比，防止只改交易内容不改 TxRoot 的篡改。
		if txRoot(b.Txs) != b.TxRoot {
			return ErrInvalidChain
		}
		if i > 0 && b.PrevHash != prevHash {
			return ErrInvalidChain
		}
		// PoW 难度检查：哈希前缀应有 difficulty 个 0
		if !strings.HasPrefix(b.Hash, strings.Repeat("0", bc.difficulty)) {
			return ErrInvalidChain
		}
		prevHash = b.Hash
	}
	return nil
}
