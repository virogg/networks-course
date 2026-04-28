package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/virogg/networks-course/lab08/solution/internal/snw/server"
)

func main() {
	addr := flag.String("addr", ":9000", "UDP-адрес для прослушивания")
	timeout := flag.Duration("timeout", time.Second, "таймаут retransmit для Stop-and-Wait")
	chunk := flag.Int("chunk-size", 1024, "размер payload в одном кадре")
	loss := flag.Float64("loss", 0.3, "вероятность потери исходящего кадра [0..1]")
	corrupt := flag.Float64("corrupt-prob", 0.0, "вероятность сбить бит payload (для теста checksum)")
	sendFile := flag.String("send-file", "", "файл, который сервер отправит клиенту")
	recvFile := flag.String("recv-file", "", "файл, в который сервер сохранит входящие данные")
	seed := flag.Int64("seed", 0, "seed для генератора потерь (0 = текущее время)")
	flag.Parse()

	srv := server.NewServer(
		server.WithAddr(*addr),
		server.WithTimeout(*timeout),
		server.WithChunkSize(*chunk),
		server.WithLossProb(*loss),
		server.WithCorruptProb(*corrupt),
		server.WithSendFile(*sendFile),
		server.WithRecvFile(*recvFile),
		server.WithSeed(*seed),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatal(err)
	}
}
