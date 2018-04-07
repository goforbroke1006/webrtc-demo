package main

import (
	"net/http"
	"fmt"
	"os"
	"io/ioutil"
	"io"
	"log"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebRtcStreamMeta struct {
	Type string `json:"type"`
	SDP  string `json:"sdp"`
}

type Peer struct {
	SDP *WebRtcStreamMeta `json:"sdp"`
	ICE *WebRtcStreamMeta `json:"ice"`
}

var connections = []*websocket.Conn{}

func main() {
	logger := log.New(os.Stdout, "server: ", log.Lshortfile)

	router := mux.NewRouter()
	router.HandleFunc("/{filepath:[0-9a-zA-Z/-]+}.{extension:html|css|js|png|jpg}", staticHandler).Methods("GET")
	router.HandleFunc("/stream", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		connections = append(connections, conn)

		peer := Peer{}
		err = conn.ReadJSON(peer)
		if err != nil {
			log.Println(err)
			return
		}
		//fmt.Println("data", string(data))

		for _, c := range connections {
			/*if c == conn {
				continue
			}*/
			c.WriteJSON(peer)
		}
	})
	http.Handle("/", router)
	if err := http.ListenAndServe(":8036", nil); nil != err {
		logger.Fatal(err.Error())
	}
}

func staticHandler(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	mimeSufix := "plain"
	switch params["extension"] {
	case "js":
		mimeSufix = "javascript"
	default:
		mimeSufix = params["extension"]
	}

	w.Header().Set("Content-Type", "text/"+mimeSufix)
	w.WriteHeader(http.StatusOK)

	fi, err := os.Open(fmt.Sprintf("public/%s.%s",
		params["filepath"],
		params["extension"],
	))
	if err != nil {
		panic(err)
	}

	n, err := ioutil.ReadAll(fi)
	if err != nil && err != io.EOF {
		panic(err)
	}

	w.Write(n)
}
