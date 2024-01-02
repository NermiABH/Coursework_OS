package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

var (
	file *os.File
	mu   sync.Mutex
)

func main() {
	var err error
	file, err = os.OpenFile("logfile.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Ошибка открытия файла логов:", err)
	}
	defer file.Close()

	listener, err := net.Listen("tcp", ":8083")
	if err != nil {
		fmt.Println("Ошибка прослушивания:", err)
		return
	}
	defer listener.Close()

	log.Println("Сервер 3 запущен. Ожидание подключений...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка приема соединения:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Ошибка при чтении данных:", err)
		return
	}
	receivedData := buffer[:n]
	writeToLog(string(receivedData))
}

func writeToLog(msg string) {
	mu.Lock()
	_, err := file.WriteString(msg + "\n")
	if err != nil {
		fmt.Println("Ошибка записи в файл:", err)
	}
	mu.Unlock()
}
