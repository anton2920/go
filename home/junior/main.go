package main

import (
	"unsafe"

	"github.com/anton2920/gofa/container/binary"
	"github.com/anton2920/gofa/container/linked"
	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/cpu"
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

func NewIntListNode(ctx *context.Context, value int) *linked.ListNode {
	n := (*IntListNode)(ctx.Arena.PushSizeWithAlignment(unsafe.Sizeof(IntListNode{}), unsafe.Alignof(IntListNode{})))
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

func NewIntTreeNode(ctx *context.Context, value int) *binary.TreeNode {
	n := (*IntTreeNode)(ctx.Arena.PushSizeWithAlignment(unsafe.Sizeof(IntTreeNode{}), unsafe.Alignof(IntTreeNode{})))
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

func PrintTreeInOrder(ctx *context.Context, bt *binary.SearchTree) {
	ctx.Fmt.Reset()
	traverseTreeNodesInOrder(ctx, ToIntTreeNode(bt.Root))
	fmt_.Println(ctx, ctx.Fmt.Backspace(2))
}

func traverseTreeNodesPreOrderPretty(ctx *context.Context, n *IntTreeNode, level int) {
	if n == nil {
		for i := 0; i < level; i++ {
			ctx.Fmt.S("\t")
		}
		ctx.Fmt.S("nil").Ln()
	} else {
		for i := 0; i < level; i++ {
			ctx.Fmt.S("\t")
		}
		ctx.Fmt.D(n.Value).Ln()

		traverseTreeNodesPreOrderPretty(ctx, ToIntTreeNode(n.Left), level+1)
		traverseTreeNodesPreOrderPretty(ctx, ToIntTreeNode(n.Right), level+1)
	}
}

func PrintTreePreOrderPretty(ctx *context.Context, bt *binary.SearchTree) {
	ctx.Fmt.Reset()
	traverseTreeNodesPreOrderPretty(ctx, ToIntTreeNode(bt.Root), 0)
	fmt_.Println(ctx, &ctx.Fmt)
}

func Plural(n int, s string) string {
	if n == 1 {
		return s[:len(s)-1]
	}
	return s
}

func Main(ctx *context.Context) {
	if !Test(ctx) {
		fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed for testing purposes: ").S(ctx.Error()))
	}

	{
		save := ctx.Arena

		ll := (*linked.List)(ctx.Arena.PushSizeWithAlignment(unsafe.Sizeof(linked.List{}), unsafe.Alignof(linked.List{})))
		for i := 0; i < 10; i++ {
			ll.Append(NewIntListNode(ctx, i))
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
		bt.Comparator = func(a binary.TreeKeyPointer, b binary.TreeKeyPointer) int {
			return *(*int)(a) - *(*int)(b)
		}

		keys := ctx.Arena.PushIntArray(10)
		for i := 0; i < len(keys); i++ {
			n, _ := cpu.ReadRandomUint16()
			keys[i] = int(n % 100)
			bt.Add(NewIntTreeNode(ctx, keys[i]))
		}

		duplicates := len(keys) - bt.N
		if duplicates > 0 {
			//fmt_.Println(ctx, ctx.Fmt.Reset().S("Generated ").D(duplicates).S(" ").S(Plural(duplicates, "duplicates")))
		}

		for i := 0; i < len(keys); i++ {
			context_.Must(ctx, bt.Has(binary.TreeKeyPointer(&keys[i])), "Failed to find existing key in the tree")
		}

		PrintTreeInOrder(ctx, bt)
		for i := 0; i < len(keys); i++ {
			//PrintTreeInOrder(ctx, bt)
			//PrintTreePreOrderPretty(ctx, bt)
			//fmt_.Println(ctx, ctx.Fmt.Reset().S("Removing ").D(keys[i]).S("..."))
			if !bt.Has(binary.TreeKeyPointer(&keys[i])) {
				duplicates--
			}
			context_.Must(ctx, (bt.Del(binary.TreeKeyPointer(&keys[i])) != nil) || (duplicates >= 0), "Failed to delete existing key in the tree")
		}
		//PrintTreeInOrder(ctx, bt)

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

func main() {
	//runtime.AllocationsAreDisabled = true

	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithEvenlySplitSize(&ctx, ints.GiB(1)), "Failed to bootstrap context")

	trace_.BeginProfile()
	Main(ctx.Noescape())
	trace_.EndAndPrintProfile()
}
