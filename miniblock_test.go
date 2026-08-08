package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChainMineAndValidate(t *testing.T) {
	bc := newBlockchain(2)
	bc.AddTx(Transaction{From: "a", To: "b", Amount: 10})
	bc.AddTx(Transaction{From: "b", To: "c", Amount: 5})
	b := bc.Mine()
	if b.Index != 1 {
		t.Fatalf("新块 index 应为 1, 得到 %d", b.Index)
	}
	if !strings.HasPrefix(b.Hash, "00") {
		t.Fatalf("难度 2 哈希应以 00 开头, 得到 %s", b.Hash)
	}
	if len(bc.Pending()) != 0 {
		t.Fatal("挖完后待打包应清空")
	}
	if err := bc.Validate(); err != nil {
		t.Fatalf("链应当合法: %v", err)
	}
}

func TestTamperDetected(t *testing.T) {
	bc := newBlockchain(2)
	bc.AddTx(Transaction{From: "a", To: "b", Amount: 10})
	bc.Mine()
	// 篡改第二个块的前驱哈希，破坏哈希链，校验应失败。
	chain := bc.Chain()
	chain[1].PrevHash = "0x tampered"
	bc.mu.Lock()
	bc.chain = chain
	bc.mu.Unlock()
	if err := bc.Validate(); err != ErrInvalidChain {
		t.Fatalf("篡改前驱后校验应失败, 得到 %v", err)
	}

	// 篡改交易金额（不动 TxRoot），重算交易根应能发现。
	bc2 := newBlockchain(2)
	bc2.AddTx(Transaction{From: "a", To: "b", Amount: 10})
	bc2.Mine()
	c2 := bc2.Chain()
	c2[1].Txs[0].Amount = 999
	bc2.mu.Lock()
	bc2.chain = c2
	bc2.mu.Unlock()
	if err := bc2.Validate(); err != ErrInvalidChain {
		t.Fatalf("篡改交易后校验应失败, 得到 %v", err)
	}
}

func TestHTTPFlow(t *testing.T) {
	bc := newBlockchain(1)
	mux := http.NewServeMux()
	bc.Register(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 提交交易
	resp, _ := http.Post(ts.URL+"/tx", "application/json", strings.NewReader(`{"from":"a","to":"b","amount":10}`))
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("提交交易状态码 %d", resp.StatusCode)
	}

	// 挖矿
	m, _ := http.Post(ts.URL+"/mine", "application/json", nil)
	defer m.Body.Close()
	if m.StatusCode != 200 {
		t.Fatalf("挖矿状态码 %d", m.StatusCode)
	}

	// 校验
	v, _ := http.Get(ts.URL + "/validate")
	defer v.Body.Close()
	var vr struct {
		Valid string `json:"valid"`
	}
	json.NewDecoder(v.Body).Decode(&vr)
	if vr.Valid != "true" {
		t.Fatalf("链应合法, 得到 %v", vr)
	}
}
