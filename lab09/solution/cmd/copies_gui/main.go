package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/virogg/networks-course/lab09/solution/internal/copies"
)

func main() {
	port := flag.Int("port", 9999, "broadcast port")
	interval := flag.Duration("interval", 2*time.Second, "broadcast interval")
	dead := flag.Int("dead-multiplier", 3, "drop peer after dead-multiplier * interval")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	core := copies.NewApp(
		copies.WithPort(*port),
		copies.WithInterval(*interval),
		copies.WithDeadMultiplier(*dead),
	)

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := core.Run(ctx); err != nil {
			log.Printf("core: %v", err)
		}
	})

	a := app.New()
	w := a.NewWindow("Анализ запущенных копий")
	w.Resize(fyne.NewSize(280, 360))

	countLbl := widget.NewLabel("Копий запущено: 0")
	intervalEntry := widget.NewEntry()
	intervalEntry.SetText(fmt.Sprintf("%d", interval.Milliseconds()))
	intervalEntry.Disable()
	intervalRow := container.NewBorder(nil, nil, widget.NewLabel("Ожидание, мс"), nil, intervalEntry)

	peerList := widget.NewLabel("")
	peerList.Wrapping = fyne.TextWrapWord
	listScroll := container.NewVScroll(peerList)

	var fyneDone atomic.Bool
	shutdown := func() {
		fyneDone.Store(true)
		stop()
		a.Quit()
	}

	closeBtn := widget.NewButton("Закрыть", shutdown)

	w.SetContent(container.NewBorder(
		container.NewVBox(countLbl, intervalRow),
		closeBtn,
		nil, nil,
		listScroll,
	))

	w.SetCloseIntercept(shutdown)

	wg.Go(func() { runRefresher(ctx, core, *interval, countLbl, peerList) })

	wg.Go(func() {
		<-ctx.Done()
		if fyneDone.Load() {
			return
		}
		fyne.Do(func() { a.Quit() })
	})

	w.ShowAndRun()
	fyneDone.Store(true)
	stop()
	wg.Wait()
}

func runRefresher(ctx context.Context, core *copies.App, interval time.Duration, count, list *widget.Label) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			snap := core.Snapshot()
			var sb strings.Builder
			sb.WriteString(snap.Self)
			sb.WriteString("  (self)\n")
			for _, p := range snap.Peers {
				sb.WriteString(p.ID)
				sb.WriteByte('\n')
			}
			text := sb.String()
			fyne.Do(func() {
				count.SetText(fmt.Sprintf("Копий запущено: %d", len(snap.Peers)+1))
				list.SetText(text)
			})
		}
	}
}
