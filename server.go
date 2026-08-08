package main

import (
	"encoding/json"
	"net/http"
)

// Register 挂上区块链的 HTTP 接口。
func (bc *Blockchain) Register(mux *http.ServeMux) {
	// 查看整条链
	mux.HandleFunc("/chain", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, bc.Chain())
	})

	// 查看待打包交易
	mux.HandleFunc("/pending", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, bc.Pending())
	})

	// 提交一笔交易
	mux.HandleFunc("/tx", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支持 POST", http.StatusMethodNotAllowed)
			return
		}
		var t Transaction
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			writeErr(w, "body 要像 {\"from\":\"a\",\"to\":\"b\",\"amount\":10}", http.StatusBadRequest)
			return
		}
		if err := bc.AddTx(t); err != nil {
			writeErr(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]string{"status": "queued"})
	})

	// 挖矿：把待打包交易打包成新块
	mux.HandleFunc("/mine", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支持 POST", http.StatusMethodNotAllowed)
			return
		}
		b := bc.Mine()
		writeJSON(w, b)
	})

	// 校验整条链
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		if err := bc.Validate(); err != nil {
			writeJSON(w, map[string]string{"valid": "false", "error": err.Error()})
			return
		}
		writeJSON(w, map[string]string{"valid": "true"})
	})
}
