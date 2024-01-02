package utils

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	HostDefault = "127.0.0.1"
)

var (
	HostServer3 string
)

type Logger struct {
	addr      string
	sourceSrv string
	conn      net.Conn
	mu        sync.Mutex
}

func NewLogger(sourceSrv string) *Logger {
	HostServer3 = os.Getenv("HOST_SERVER_3")
	if len(HostServer3) == 0 {
		HostServer3 = HostDefault
	}
	addr := HostServer3 + ":8083"
	logger := &Logger{
		addr:      addr,
		sourceSrv: sourceSrv,
	}
	logger.restartConn()
	return logger
}

func (l *Logger) Info(v ...any) {
	l.print("Info", v)
}

func (l *Logger) Warn(v ...any) {
	l.print("Warn", v)
}

func (l *Logger) Fatal(v ...any) {
	l.print("Fatal", v)
	os.Exit(1)
}

func (l *Logger) print(level string, v ...any) {
	msg := fmt.Sprintf("%s %s from %v %s ", time.Now().Format(time.DateTime),
		level, l.sourceSrv, fmt.Sprint(v...))
	fmt.Println(msg)
	l.mu.Lock()
	for {
		if err := l.sendToServer3(msg); err != nil {
			l.restartConn()
			<-time.After(time.Second)
		} else {
			break
		}
	}
	l.mu.Unlock()
}

func (l *Logger) restartConn() {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", l.addr)
		if err != nil {
			log.Println("failed restart conn to server3 ", err)
			<-time.After(time.Second)
			continue
		}
		l.conn = conn
		return
	}
}

func (l *Logger) sendToServer3(msg string) error {
	_, err := l.conn.Write([]byte(msg + "\n"))
	return err
}
