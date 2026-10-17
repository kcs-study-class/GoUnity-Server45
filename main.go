package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// Itemはやり取りする１件分のデータ
// 構造体フィールドの後ろの `json:"..."` はJsonにしたときのキー名
type Item struct {
	ID   int    `json:"id"`   // サーバー側で自動採番するID
	Name string `json:"name"` //名前
}

// データの保管場所（OnMemory）
var (
	mu     sync.Mutex // itemsを同時に読み書きしても壊れないようにするロック
	items  = []Item{} //登録されたItem一覧
	nextID = 1        //次に採番するID
)

func main() {
	http.HandleFunc("/items", itemsHandler)

	addr := ":8080"
	log.Printf("サーバー起動するよー: http://localhost%s", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}

}

// itemsHandler は　/items へのアクセスをHTTPメソッドごとに処理する
// GET ... 登録済の一覧をJsonで返却
// Post... Jsonを受け取って１件追加して、追加結果をJsonで返すもの。
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		// GET もしくは　POST以外は許可しない。　405 を返却[そのメソッドは不可].
		http.Error(w, "GET もしくは POST を利用してください。", http.StatusMethodNotAllowed)
	}

}

// handleGet 現在のアイテム一覧をJsonで返却
func handleGet(w http.ResponseWriter, r *http.Request) {
	// mu.Lock(): 一覧を読む間はロックして安全に扱うようにする
	mu.Lock()
	// defer mu.Unlock(): 関数を抜けるときに自動でロックを解除する.
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	// items をjsonに変換.
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handlePost は受け取ったJsonを１件追加して追加後のItemを返却します
func handlePost(w http.ResponseWriter, r *http.Request) {
	// 受信用の変数　クライアントは　{"name": "..."} を送って来る想定
	var in Item

	//リクエストの本文のJsonをinに読み込む
	// 書き換えたいので &in (ポインター)を渡します
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w,
			"JSON  の読み取りに失敗しました。"+err.Error(),
			http.StatusBadRequest)
	}

	if in.Name == "" {
		http.Error(w, "name は必須です。", http.StatusBadRequest)
		return
	}

	if in.Name == " " || in.Name == "　" {
		http.Error(w,
			"name に不正な文字があります。name:"+in.Name,
			http.StatusBadRequest)
		return
	}

	mu.Lock()
	in.ID = nextID
	nextID++
	//配列追加.
	items = append(items, in)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	// items をjsonに変換.
	if err := json.NewEncoder(w).Encode(in); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
