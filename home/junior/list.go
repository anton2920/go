package main

import (
	"unsafe"

	"github.com/anton2920/gofa/container/linked"
	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/fmt/fmt_"
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

func TestLinkedList(ctx *context.Context) {
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
