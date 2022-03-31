package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Porges/maat"
)

type Node struct {
	value int
	next  *Node
}

type LinkedList struct{ root *Node }

func (ll LinkedList) String() string {
	sb := strings.Builder{}

	node := ll.root
	first := true
	for node != nil {
		if first {
			first = false
		} else {
			sb.WriteString(", ")
		}

		sb.WriteString(fmt.Sprintf("%d", node.value))
		node = node.next
	}

	return sb.String()
}

var _ fmt.Stringer = LinkedList{}

func (ll LinkedList) PushFront(val int) {
	ll.root = &Node{
		value: val,
		next:  ll.root,
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

func (ll LinkedList) ReverseSimple() LinkedList {
	result := LinkedList{}

	node := ll.root
	for node != nil {
		result.PushFront(node.value)
		node = node.next
	}

	return result
}

func GenLinkedList(g maat.G, name string) LinkedList {
	return maat.Derive(
		g,
		name,
		func() LinkedList {
			length := maat.Byte(g, "length")
			result := LinkedList{}
			for i := 0; i < int(length); i++ {
				result.PushFront(maat.Int(g, fmt.Sprintf("value_%v", i)))
			}

			return result
		})
}

func TestRecursive(t *testing.T) {
	m := maat.New(t)
	m.Boolean(
		"linked list",
		func(g maat.G) bool {
			list := GenLinkedList(g, "list")
			return list.Reverse().Equals(list.ReverseSimple())
		})
}
