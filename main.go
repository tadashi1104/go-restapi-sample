package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // 接続されるクライアント
var broadcast = make(chan Message)           // メッセージブロードキャストチャネル

// upgrader
var upgrader = websocket.Upgrader{}

// Message struct
type Message struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/countries", GetAllCountries),
		rest.Post("/countries", PostCountry),
		rest.Get("/countries/:code", GetCountry),
		rest.Delete("/countries/:code", DeleteCountry),
		rest.Get("/100", Get101),
		rest.Get("/200", Get200),
		rest.Get("/300", Get302),
		rest.Get("/400", Get400),
		rest.Get("/500", Get500),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)

	// ファイルサーバーを立ち上げる
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
	// websockerへのルーティングを紐づけ
	http.HandleFunc("/ws", handleConnections)
	go handleMessages()
	go func() {
		// サーバーをlocalhostのポート8000で立ち上げる
		log.Println("http server started on :8000")
		err := http.ListenAndServe(":8000", nil)
		// エラーがあった場合ロギングする
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 送られてきたGETリクエストをwebsocketにアップグレード
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// 関数が終わった際に必ずwebsocketnのコネクションを閉じる
	defer ws.Close()

	// クライアントを新しく登録
	clients[ws] = true

	for {
		var msg Message
		// 新しいメッセージをJSONとして読み込みMessageオブジェクトにマッピングする
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// 新しく受信されたメッセージをブロードキャストチャネルに送る
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// ブロードキャストチャネルから次のメッセージを受け取る
		msg := <-broadcast
		// 現在接続しているクライアント全てにメッセージを送信する
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// Thing for Thing
type Thing struct {
	Name string
}

var b bool

// Get101 for to get 101
func Get101(w rest.ResponseWriter, r *rest.Request) {

	if b == false {
		b = true
		w.WriteHeader(http.StatusSwitchingProtocols)
	} else {
		b = false
		w.WriteHeader(http.StatusOK)
	}

	// cpt := 0
	// for {
	// 	cpt++
	// 	w.WriteJson(
	// 		&Thing{
	// 			Name: fmt.Sprintf("thing #%d", cpt),
	// 		},
	// 	)
	// 	w.WriteHeader(http.StatusSwitchingProtocols)
	// 	w.(http.ResponseWriter).Write([]byte("\n"))
	// 	// Flush the buffer to client
	// 	w.(http.Flusher).Flush()
	// 	// wait 1 seconds
	// 	time.Sleep(time.Duration(1) * time.Second)
	// 	if cpt == 10 {
	// 		return
	// 	}
	// }
}

// Get200 for to get 200
func Get200(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

// Get302 for to get 302
func Get302(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusFound)
	return
}

// Get400 for to get 400
func Get400(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusBadRequest)
	return
}

// Get500 for to get 500
func Get500(w rest.ResponseWriter, r *rest.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	return
}

// Country model
type Country struct {
	Code string
	Name string
}

var store = map[string]*Country{}

var lock = sync.RWMutex{}

// GetCountry for to get country
func GetCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")

	lock.RLock()
	var country *Country
	if store[code] != nil {
		country = &Country{}
		*country = *store[code]
	}
	lock.RUnlock()

	if country == nil {
		rest.NotFound(w, r)
		return
	}
	w.WriteJson(country)
}

// GetAllCountries for to get all country
func GetAllCountries(w rest.ResponseWriter, r *rest.Request) {
	lock.RLock()
	countries := make([]Country, len(store))
	i := 0
	for _, country := range store {
		countries[i] = *country
		i++
	}
	lock.RUnlock()
	w.WriteJson(&countries)
}

// PostCountry for to register country
func PostCountry(w rest.ResponseWriter, r *rest.Request) {
	country := Country{}
	err := r.DecodeJsonPayload(&country)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if country.Code == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if country.Name == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	lock.Lock()
	store[country.Code] = &country
	lock.Unlock()
	w.WriteJson(&country)
}

// DeleteCountry for to delete country
func DeleteCountry(w rest.ResponseWriter, r *rest.Request) {
	code := r.PathParam("code")
	lock.Lock()
	delete(store, code)
	lock.Unlock()
	w.WriteHeader(http.StatusOK)
}
