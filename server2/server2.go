package main

import (
	"Coursework_OS/utils"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"time"
	"unsafe"
)

var (
	logger *utils.Logger
)

func main() {
	logger = utils.NewLogger("Server1")

	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		logger.Fatal("Ошибка прослушивания:", err)
	}
	defer listener.Close()

	logger.Info("Сервер запущен. Ожидание подключений...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				logger.Fatal("Сервер отключился:", err)
			}
			logger.Warn("Ошибка приема соединения:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	buffer := make([]byte, 32)
	n, err := conn.Read(buffer)
	if err != nil {
		logger.Warn("Ошибка при чтении данных:", err)
		return
	}
	receivedData := string(buffer[:n])
	switch receivedData {
	case "/v1":
		handleV1Connection(conn)
	case "/v2":
		handleV2Connection(conn)
	}
}

func handleV1Connection(conn net.Conn) {
	var response string
	startTime := time.Now()
	width, height, err := getTerminalSize()
	if err != nil {
		response += fmt.Sprintf("Не удалось получить информацию о размере экрана")
		logger.Warn("Не удалось получить информацию о размере экрана \n")
	} else {
		response += fmt.Sprintf("Ширина экрана: %d символов.\nВысота экрана: %d строк \n", width, height)
	}
	endTime := time.Now()
	timeDiff := "Время рабочего процесса " + endTime.Sub(startTime).String()
	response += timeDiff + "\n"
	_, err = conn.Write([]byte(response))
	if err != nil {
		logger.Warn("Ошибка при записи данных:", err)
		return
	}
}

func handleV2Connection(conn net.Conn) {
	var widthOld, heightOld int
	var timeDiffOld string
	for {
		var response string
		startTime := time.Now()
		width, height, err := getTerminalSize()
		if err != nil {
			response += fmt.Sprintf("Не удалось получить информацию о размере экрана \n")
			logger.Warn("Не удалось получить информацию о размере экрана \n")
		} else {
			if widthOld != width || heightOld != height {
				response += fmt.Sprintf("Ширина экрана: %d символов.\nВысота экрана: %d строк \n ", width, height)
			}
		}
		endTime := time.Now()
		timeDiff := endTime.Sub(startTime).String()
		if timeDiffOld != timeDiff {
			response += "Время рабочего процесса " + timeDiff + "\n"
		}
		if len(response) != 0 {
			logger.Warn("Информация изменилась")
			_, err = conn.Write([]byte(response))
			if err != nil {
				logger.Warn("Ошибка при записи данных:", err)
				return
			}
		}
	}
}

func getTerminalSize() (int, int, error) {
	var w unix.Winsize
	if _, _, err := unix.Syscall(unix.SYS_IOCTL, 1, unix.TIOCGWINSZ, uintptr(unsafe.Pointer(&w))); err != 0 {
		return 0, 0, err
	}
	return int(w.Col), int(w.Row), nil
}
