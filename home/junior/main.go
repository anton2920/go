package main

import (
	"runtime"

	"github.com/anton2920/gofa/bytes"
	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	_ "github.com/anton2920/gofa/debug/debug_"
	"github.com/anton2920/gofa/fmt/fmt_"
	"github.com/anton2920/gofa/ints"
	"github.com/anton2920/gofa/os"
	"github.com/anton2920/gofa/trace/trace_"
)

func Test(ctx *context.Context) bool {
	f, ok := os.OpenFile(ctx, "whatever", os.OpenForReading)
	if !ok {
		ctx.NewError().S("failed to open file: ").S(ctx.OldError())
		return false
	}
	os.CloseHandle(ctx, f)
	return true
}

func Main(ctx *context.Context) {
	if false {
		if !Test(ctx) {
			fmt_.Eprintln(ctx, ctx.Fmt.Reset().S("Failed for testing purposes: ").S(ctx.Error()))
			ctx.ResetError()
		}
		TestLinkedList(ctx)
		TestBinaryTree(ctx)
		TestAVLTree(ctx)
	}

	size := os.PageSize

	fd, ok := os.OpenOrCreateFile(ctx, "data.bin", os.OpenForReading|os.OpenForWriting, os.CreateFileIfItDoesNotExist, 0644)
	context_.Must(ctx, ok, "Failed to open/create data file")
	context_.Must(ctx, os.ResizeFile(ctx, fd, size), "Failed to resize data file")

	ptr, ok := os.AllocateFileBackedVirtualMemory(ctx, size, fd)
	context_.Must(ctx, ok, "Failed to allocate file-backed virtual memory")
	context_.Must(ctx, os.CloseHandle(ctx, fd), "Failed to close data file")

	buf := bytes.SliceFromUnsafePointer(ptr, size)
	copy(buf, "Hello, world!")

	context_.Must(ctx, os.DeallocateVirtualMemory(ctx, ptr, size), "Failed to deallocate file-backed virtual memory")

	EchoServer(ctx)
}

func main() {
	runtime.AllocationsAreDisabled = true

	var ctx context.Context
	context_.Must(&ctx, context_.BootstrapWithEvenlySplitSize(&ctx, ints.GiB(1)), "Failed to bootstrap context")

	trace_.BeginProfile()
	Main(ctx.Noescape())
	trace_.EndAndPrintProfile()
}
