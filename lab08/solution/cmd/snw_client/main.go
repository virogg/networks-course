package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab08/solution/internal/snw/client"
)

func main() {
	host := flag.String("host", "127.0.0.1", "адрес сервера")
	port := flag.Int("port", 9000, "порт сервера")
	timeout := flag.Duration("timeout", time.Second, "таймаут retransmit для Stop-and-Wait")
	chunk := flag.Int("chunk-size", 1024, "размер payload в одном кадре")
	loss := flag.Float64("loss", 0.3, "вероятность потери исходящего кадра [0..1]")
	corrupt := flag.Float64("corrupt-prob", 0.0, "вероятность сбить бит payload (для теста checksum)")
	sendFile := flag.String("send-file", "", "файл для отправки серверу")
	recvFile := flag.String("recv-file", "", "файл, в который клиент сохранит входящие данные")
	seed := flag.Int64("seed", 0, "seed для генератора потерь (0 = текущее время)")
	flag.Parse()

	c := client.NewClient(
		client.WithHost(*host),
		client.WithPort(*port),
		client.WithTimeout(*timeout),
		client.WithChunkSize(*chunk),
		client.WithLossProb(*loss),
		client.WithCorruptProb(*corrupt),
		client.WithSendFile(*sendFile),
		client.WithRecvFile(*recvFile),
		client.WithSeed(*seed),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := c.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
