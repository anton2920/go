package main

import (
	"unsafe"

	"github.com/anton2920/gofa/container/binary"
	"github.com/anton2920/gofa/context"
	"github.com/anton2920/gofa/context/context_"
	"github.com/anton2920/gofa/cpu"
	"github.com/anton2920/gofa/fmt/fmt_"
)

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

func traverseTreeNodesInOrder(ctx *context.Context, n *IntTreeNode) {
	if n != nil {
		traverseTreeNodesInOrder(ctx, ToIntTreeNode(n.Left))
		ctx.Fmt.D(n.Value).S(", ")
		traverseTreeNodesInOrder(ctx, ToIntTreeNode(n.Right))
	}
}

func PrintTreeInOrder(ctx *context.Context, node *binary.TreeNode) {
	ctx.Fmt.Reset()
	traverseTreeNodesInOrder(ctx, ToIntTreeNode(node))
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

func PrintTreePreOrderPretty(ctx *context.Context, node *binary.TreeNode) {
	ctx.Fmt.Reset()
	traverseTreeNodesPreOrderPretty(ctx, ToIntTreeNode(node), 0)
	fmt_.Println(ctx, &ctx.Fmt)
}

func Plural(n int, s string) string {
	if n == 1 {
		return s[:len(s)-1]
	}
	return s
}

func TestBinaryTree(ctx *context.Context) {
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

	PrintTreeInOrder(ctx, bt.Root)
	for i := 0; i < len(keys); i++ {
		//PrintTreeInOrder(ctx, bt.Root)
		//PrintTreePreOrderPretty(ctx, bt.Root)
		//fmt_.Println(ctx, ctx.Fmt.Reset().S("Removing ").D(keys[i]).S("..."))
		if !bt.Has(binary.TreeKeyPointer(&keys[i])) {
			duplicates--
		}
		context_.Must(ctx, (bt.Del(binary.TreeKeyPointer(&keys[i])) != nil) || (duplicates >= 0), "Failed to delete existing key in the tree")
	}
	//PrintTreeInOrder(ctx, bt.Root)

	ctx.Arena = save
}

func TestAVLTree(ctx *context.Context) {
	save := ctx.Arena

	bt := (*binary.AVLTree)(ctx.Arena.PushSizeWithAlignment(unsafe.Sizeof(binary.AVLTree{}), unsafe.Alignof(binary.AVLTree{})))
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

	PrintTreeInOrder(ctx, bt.Root)
	for i := 0; i < len(keys); i++ {
		//PrintTreeInOrder(ctx, bt.Root)
		//PrintTreePreOrderPretty(ctx, bt.Root)
		//fmt_.Println(ctx, ctx.Fmt.Reset().S("Removing ").D(keys[i]).S("..."))
		if !bt.Has(binary.TreeKeyPointer(&keys[i])) {
			duplicates--
		}
		context_.Must(ctx, (bt.Del(binary.TreeKeyPointer(&keys[i])) != nil) || (duplicates >= 0), "Failed to delete existing key in the tree")
	}
	//PrintTreeInOrder(ctx, bt.Root)

	ctx.Arena = save
}
