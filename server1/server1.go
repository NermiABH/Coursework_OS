package main

import (
	"Coursework_OS/utils"
	"fmt"
	"net"
	"syscall"
	"time"
)

var (
	logger *utils.Logger
)

func main() {
	logger = utils.NewLogger("Server1")

	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		logger.Warn("Ошибка прослушивания:", err)
		return
	}
	defer listener.Close()

	logger.Info("Сервер запущен. Ожидание подключений...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				logger.Fatal("Сервер отключился:", err)
				return
			}
			logger.Warn("Ошибка приема соединения:", err)
			continue
		}
		go handlerConnection(conn)
	}
}

func handlerConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 32)
	n, err := conn.Read(buffer)
	if err != nil {
		logger.Warn("Ошибка при чтении данных:", err)
		return
	}
	receivedData := string(buffer[:n])
	logger.Info("Произошел запрос: ", receivedData)
	switch receivedData {
	case "/v1":
		handleV1Connection(conn)
	case "/v2":
		handleV2Connection(conn)
	}
}

func handleV1Connection(conn net.Conn) {
	var si syscall.Sysinfo_t
	var response string
	if err := syscall.Sysinfo(&si); err == nil {
		swapFileSize := int64(si.Totalswap) * int64(si.Unit)
		freeSwap := int64(si.Freeswap) * int64(si.Unit)
		response = fmt.Sprintf("Размер файла подкачки: %d байт\n "+
			"Свободные байты в файле подкачки: %d байт",
			swapFileSize, freeSwap)
	} else {
		response = "Не удалось получить информацию о системе."
	}
	_, err := conn.Write([]byte(response))
	if err != nil {
		logger.Warn("Ошибка при записи данных:", err)
		return
	}
}

func handleV2Connection(conn net.Conn) {
	var swapFileSizeOld int64
	var freeSwapOld int64
	for {
		var si syscall.Sysinfo_t
		var response string
		if err := syscall.Sysinfo(&si); err == nil {
			swapFileSize, freeSwap := int64(si.Totalswap)*int64(si.Unit), int64(si.Freeswap)*int64(si.Unit)
			if swapFileSizeOld != swapFileSize || freeSwapOld != freeSwap {
				response = fmt.Sprintf("Размер файла подкачки: %d байт\n "+
					"Свободные байты в файле подкачки: %d байт",
					swapFileSize, freeSwap)
				swapFileSizeOld, freeSwapOld = swapFileSize, freeSwap
				logger.Warn("Данные поменялись")
			}
		} else {
			response = "Не удалось получить информацию о системе."
			logger.Warn("Не удалось получить информцию о системе")
		}
		if len(response) != 0 {
			_, err := conn.Write([]byte(response))
			if err != nil {
				logger.Warn("Ошибка при записи данных:", err)
				return
			}
		}
		<-time.After(time.Second * 5)
	}
}
