package main

import (
	"runtime"
	"unsafe"

	"github.com/anton2920/gofa/container/binary"
	"github.com/anton2920/gofa/container/linked"
	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	_ "github.com/anton2920/gofa/debug/debug_"
	"github.com/anton2920/gofa/event"
	"github.com/anton2920/gofa/event/event_"
	"github.com/anton2920/gofa/fmt/fmt_"
	"github.com/anton2920/gofa/ints"
	"github.com/anton2920/gofa/log/log_"
	"github.com/anton2920/gofa/mem/mem_"
	"github.com/anton2920/gofa/net/net_"
	"github.com/anton2920/gofa/os"
	"github.com/anton2920/gofa/time"
	"github.com/anton2920/gofa/time/time_"
	"github.com/anton2920/gofa/trace/trace_"
)

type IntListNode struct {
	linked.ListNode
	Value int
}

func ToIntListNode(n *linked.ListNode) *IntListNode {
	return (*IntListNode)(unsafe.Pointer(n))
}

func NewIntListNode(ll *linked.List, value int) *linked.ListNode {
	n := ToIntListNode(ll.NewNode())
	n.Value = value
	return (*linked.ListNode)(unsafe.Pointer(n))
}

type IntTreeNode struct {
	binary.TreeNode
	Value int
}

func ToIntTreeNode(n *binary.TreeNode) *IntTreeNode {
	return (*IntTreeNode)(unsafe.Pointer(n))
}

func NewIntTreeNode(bt *binary.SearchTree, value int) *binary.TreeNode {
	n := ToIntTreeNode(bt.NewNode())
	n.Value = value
	return (*binary.TreeNode)(unsafe.Pointer(n))
}

func Test(ctx *context.Context) bool {
	f, ok := os.OpenFile(ctx, "whatever", os.OpenForReading)
	if !ok {
		ctx.NewError().S("failed to open file: ").S(ctx.OldError())
		return false
	}
	os.CloseHandle(ctx, f)
	return true
}

func traverseTreeNodesInOrder(ctx *context.Context, n *IntTreeNode) {
	if n != nil {
		traverseTreeNodesInOrder(ctx, ToIntTreeNode(n.Left))
		ctx.Fmt.D(n.Value).S(", ")
		traverseTreeNodesInOrder(ctx, ToIntTreeNode(n.Right))
	}
}

func TraverseTreeInOrder(ctx *context.Context, bt *binary.SearchTree) {
	ctx.Fmt.Reset()
	traverseTreeNodesInOrder(ctx, ToIntTreeNode(bt.Root))
	fmt_.Println(ctx, ctx.Fmt.Backspace(2))
}

func Main(ctx *context.Context) {
	if !Test(ctx) {
		fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed for testing purposes: ").S(ctx.Error()))
	}

	{
		save := ctx.Arena

		ll := (*linked.List)(ctx.Arena.PushSizeWithAlignment(unsafe.Sizeof(linked.List{}), unsafe.Alignof(linked.List{})))
		ll.Init(&ctx.Arena, unsafe.Sizeof(0), unsafe.Alignof(0))
		for i := 0; i < 10; i++ {
			ll.Append(NewIntListNode(ll, i))
		}

		f := ctx.Fmt.Reset().S("[")
		for it := ll.First; it != nil; it = it.Next {
			f.D(ToIntListNode(it).Value).S(", ")
		}
		fmt_.Println(ctx, f.Backspace(2).S("]"))

		f = ctx.Fmt.Reset().S("[")
		for it := ll.Last; it != nil; it = it.Prev {
			f.D(ToIntListNode(it).Value).S(", ")
		}
		fmt_.Println(ctx, f.Backspace(2).S("]"))

		ctx.Arena = save
	}

	{
		save := ctx.Arena

		bt := (*binary.SearchTree)(ctx.Arena.PushSizeWithAlignment(unsafe.Sizeof(binary.SearchTree{}), unsafe.Alignof(binary.SearchTree{})))
		bt.Init(&ctx.Arena, unsafe.Sizeof(0), unsafe.Alignof(0), func(a *binary.TreeNode, b *binary.TreeNode) int {
			return ToIntTreeNode(a).Value - ToIntTreeNode(b).Value
		})

		for i := 0; i < 10; i++ {
			bt.Insert(NewIntTreeNode(bt, i))
		}
		TraverseTreeInOrder(ctx, bt)

		ctx.Arena = save
	}

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
				now := time_.NowInNanoseconds()
				if e.Identifier == 1 {
					fmt_.Println(ctx, ctx.Fmt.Reset().S("Current time: ").DateTime(now))
				} else {
					fmt_.Println(ctx, ctx.Fmt.Reset().S("TIMER FIRED AT: ").DateTime(now))
				}
			}
		}
	}
}

func main() {
	runtime.AllocationsAreDisabled = true

	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithEvenlySplitSize(&ctx, ints.GiB(1)), "Failed to bootstrap context")

	trace_.BeginProfile()
	Main(ctx.Noescape())
	trace_.EndAndPrintProfile()
}
