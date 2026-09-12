package main

import (
	"unsafe"

	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/event"
	"github.com/anton2920/gofa/event/event_"
	"github.com/anton2920/gofa/fmt/fmt_"
	"github.com/anton2920/gofa/log/log_"
	"github.com/anton2920/gofa/mem/mem_"
	"github.com/anton2920/gofa/net/net_"
	"github.com/anton2920/gofa/os"
	"github.com/anton2920/gofa/time"
	"github.com/anton2920/gofa/time/time_"
)

func EchoServer(ctx *context.Context) {
	var buf mem_.CircularBuffer
	context_.Must(ctx, buf.Init(ctx, os.PageSize), "Failed to initialize circular buffer")

	var q event_.Queue
	context_.Must(ctx, q.Init(ctx), "Failed to initialize new event queue")

	const addr = "0.0.0.0:8080"
	l, ok := net_.Listen(ctx, "tcp", addr)
	if !ok {
		context_.Fatal(ctx, "Failed to listen for incoming connections")
	}
	fmt_.Println(ctx, ctx.Fmt.Reset().S("Listening on ").S(addr).S("..."))

	_ = q.AddFile(ctx, l, event.RequestRead, event.TriggerEdge, nil)
	_ = q.AddPeriodicTimer(ctx, 1, 1, time.Second, nil)
	_ = q.AddTimerAt(ctx, 2, time_.NowInSeconds()+5, time.Second, nil)
	_ = q.AddAndIgnoreTerminateSignals(ctx)

	events := make([]os.Event, 64)
	now := time_.NowInNanoseconds()

	//log.Print("Test")
	log_.Println(ctx.Log.Reset(now).S("Test"))

	//slog.Info("Test")
	log_.Println(ctx.Log.Info(now).S("Test"))

	var quit bool
	for !quit {
		n, ok := q.GetEvents(ctx, events)
		if !ok {
			fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed to get events from queue: ").S(ctx.Error()))
			continue
		}

		for i := 0; i < n; i++ {
			e := &events[i]

			switch e.EventType {
			case os.EventTypeRead:
				if e.Identifier == uintptr(l) {
					var addr os.InternetAddress
					var addrLen uint32

					c, ok := os.AcceptIncomingConnection(ctx, l, addr.AsNetworkAddress(), &addrLen)
					if !ok {
						fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed to accept incoming connection: ").S(ctx.Error()))
					}
					q.AddFile(ctx, c, event.RequestRead, event.TriggerEdge, nil)
				} else {
					if e.EndOfFile() {
						continue
					}
					c := os.Handle(e.Identifier)

					n, ok := os.ReadFromFile(ctx, c, buf.RemainingSlice())
					if n == 0 {
						continue
					}
					if !ok {
						fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed to read data from client: ").S(ctx.Error()))
					}
					buf.Produce(n)

					m, ok := os.WriteToFile(ctx, c, buf.UnconsumedSlice())
					if m == 0 {
						continue
					}
					if !ok {
						fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed to write data to client: ").S(ctx.Error()))
					}
					buf.Consume(m)
				}
			case os.EventTypeSignal:
				fmt_.Println(ctx, ctx.Fmt.Reset().S("Received signal ").D(int(e.Identifier)).S(" (").S(os.Signal(e.Identifier).String()).S("). Exitting..."))
				quit = true
			case os.EventTypeTimer:
				//now := time_.NowInNanoseconds()
				save := ctx.Arena

				t := (*os.SecondsWithNanoseconds)(ctx.Arena.PushSizeWithAlignment(2*unsafe.Sizeof(0), unsafe.Alignof(0)))
				_ = os.GetCurrentTime(ctx, t)
				now := int64(t.Seconds)*time.Second + int64(t.Nanoseconds)
				if e.Identifier == 1 {
					fmt_.Println(ctx, ctx.Fmt.Reset().S("Current time: ").DateTime(now))
				} else {
					fmt_.Println(ctx, ctx.Fmt.Reset().S("TIMER FIRED AT: ").DateTime(now))
				}

				ctx.Arena = save
			}
		}
	}
}
