package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"nhooyr.io/websocket"
	"os"
)

//ip addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}'

const (
	HostDefault = "127.0.0.1"
)

var (
	HostServer1 string
	HostServer2 string
	HostClient  = "localhost"
)

func init() {
	HostServer1 = os.Getenv("HOST_SERVER_1")
	fmt.Println(HostServer1)
	if len(HostServer1) == 0 {
		HostServer1 = HostDefault
	}
	HostServer2 = os.Getenv("HOST_SERVER_2")
	if len(HostServer2) == 0 {
		HostServer2 = HostDefault
	}
	hostClient := os.Getenv("HOST_CLIENT")
	if len(hostClient) != 0 {
		HostClient = hostClient
	}
}

func main() {
	srv := http.Server{
		Addr:    ":8080",
		Handler: handlerHTTP(),
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func handlerHTTP() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/v1/server1", V1Handler(HostServer1, "8081"))
	mux.Handle("/v1/server2", V1Handler(HostServer2, "8082"))
	mux.Handle("/v2/server1", V2Handler(HostClient, "8080", "server1"))
	mux.Handle("/v2/server2", V2Handler(HostClient, "8080", "server2"))
	mux.Handle("/ws/server1", V2HandlerWS(HostServer1, "8081"))
	mux.Handle("/ws/server2", V2HandlerWS(HostServer2, "8082"))
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	return mux
}

func V1Handler(host string, port string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := net.Dial("tcp", host+":"+port)
		if err != nil {
			fmt.Println("Ошибка подключения к серверу:", err)
			return
		}
		defer conn.Close()

		message := "/v1"
		_, err = conn.Write([]byte(message))
		if err != nil {
			log.Println("Ошибка записи:", err)
		}

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			log.Println("Ошибка чтения ответа от сервера:", err)
			return
		}

		serverResponse := string(buffer[:n])
		ts, err := template.ParseFiles("static/index.html")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		data := ViewData{
			Order: serverResponse,
		}
		if err = ts.Execute(w, data); err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
		}
	})
}

func V2Handler(host, port, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts, err := template.ParseFiles("static/index_v2.html")
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		data := ViewData{
			Order: fmt.Sprintf("ws://%s:%s/ws/%s", host, port, path),
		}
		if err = ts.Execute(w, data); err != nil {
			log.Println(err.Error())
			http.Error(w, "Internal Server Error", 500)
		}
	})
}

func V2HandlerWS(host string, port string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ws server")
		conn, err := net.Dial("tcp", host+":"+port)
		if err != nil {
			fmt.Println("Ошибка подключения к серверу:", err)
			return
		}
		message := "/v2"
		conn.Write([]byte(message))

		response := make(chan []byte, 5)

		go func(conn net.Conn, response chan []byte) {
			buffer := make([]byte, 1024)
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					response <- []byte("Ошибка чтения ответа от сервера:")
					close(response)
					return
				}
				response <- buffer[:n]
			}
		}(conn, response)
		handleWebSocket(w, r, response)
	})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, respTcp chan []byte) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		fmt.Println("Ошибки установки соединения websockets", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "Internal Server Error")

	ctx := r.Context()

	for {
		msg, ok := <-respTcp
		if !ok {
			fmt.Println("Канал respTcp закрыт")
			return
		}
		fmt.Println(string(msg))
		err := conn.Write(ctx, websocket.MessageText, msg)
		if err != nil {
			fmt.Println("Ошибка записи по websockets")
			return
		}
	}
}

type ViewData struct {
	Order string
}
