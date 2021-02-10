package test

import (
	"fmt"
	"testing"
)

type Node struct {
	value int
	next  *Node
}

type LinkedList struct{ root *Node }

func (ll LinkedList) PushFront(val int) LinkedList {
	return LinkedList{
		root: &Node{
			value: val,
			next:  ll.root,
		},
	}
}

func (ll LinkedList) Equals(other LinkedList) bool {

	n1 := ll.root
	n2 := other.root

	for {
		if n1 == n2 {
			return true
		}

		if n1 == nil || n2 == nil || n1.value != n2.value {
			return false
		}

		n1 = n1.next
		n2 = n2.next
	}
}

func (ll LinkedList) Reverse() LinkedList {
	var output *Node = nil

	for curr := ll.root; curr != nil; curr = curr.next {
		output = &Node{
			value: curr.value,
			next:  output,
		}
	}

	return LinkedList{output}
}

func (g G) LinkedList(name string) LinkedList {
	return g.Derived(
		name,
		func() interface{} {
			length := g.Byte("length")
			result := LinkedList{}
			for i := 0; i < int(length); i++ {
				result = result.PushFront(g.Int(fmt.Sprintf("value_%v", i)))
			}

			return result
		}).(LinkedList)
}

func TestRecursive(t *testing.T) {
	m := NewMaat(t)
	m.Boolean(
		"linked list",
		func(g G) bool {
			list := g.LinkedList("list")
			return list.Reverse().Equals(list)
		})
}
