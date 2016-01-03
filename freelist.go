// Implementation of Lock Free List
// For more information read Timothy L. Harris A Pragmatic Implementation of Non-Blocking Linked-List
package freelist

import (
	"unsafe"

	"github.com/alexyer/taggedptr"
)

type Node struct {
	item string
	key  uint32
	next unsafe.Pointer // Pointer to the next node
}

type LockFreeList struct {
	head unsafe.Pointer
}

// Create new list instance.
func New() *LockFreeList {
	tail := &Node{
		key:  ^uint32(0),
		next: nil,
	}

	head := &Node{
		key:  0,
		next: unsafe.Pointer(tail),
	}

	return &LockFreeList{unsafe.Pointer(head)}
}

// Add item to list.
func (l *LockFreeList) Add(item string) bool {
	key := FNV1a_32([]byte(item))

	for {
		pred, curr := l.find(key)

		if (*Node)(curr).key == key {
			return false
		} else {
			node := &Node{
				item: item,
				key:  key,
				next: curr,
			}

			return taggedptr.CompareAndSwap(&(*Node)(pred).next, curr, unsafe.Pointer(node), 0, 0)
		}
	}
}

// Check if the list contains item.
func (l *LockFreeList) Contains(item string) bool {
	var (
		key  uint32         = FNV1a_32([]byte(item))
		curr unsafe.Pointer = l.head
		tag  uint
	)

	for (*Node)(curr).key < key {
		curr = (*Node)(curr).next
		tag = taggedptr.GetTag((*Node)(curr).next)
	}

	return (*Node)(curr).key == key && tag == 0
}

// Remove item from the list.
func (l *LockFreeList) Remove(item string) bool {
	key := FNV1a_32([]byte(item))

	for {
		pred, curr := l.find(key)

		if (*Node)(curr).key != key {
			return false
		} else {
			succ := taggedptr.GetPointer((*Node)(curr).next)

			if !taggedptr.AttemptTag(&(*Node)(curr).next, succ, 1) {
				continue
			}

			taggedptr.CompareAndSwap(&(*Node)(pred).next, curr, succ, 0, 0)
			return true
		}
	}
}

func (l *LockFreeList) find(key uint32) (unsafe.Pointer, unsafe.Pointer) {
	pred := l.head
	curr := taggedptr.GetPointer((*Node)(l.head).next)

Retry:
	for {
		succ, tag := taggedptr.Get((*Node)(curr).next)

		for tag != 0 {
			if !taggedptr.CompareAndSwap(&(*Node)(pred).next, curr, succ, 0, 0) {
				continue Retry
			}
			curr = succ
			succ = taggedptr.GetPointer((*Node)(curr).next)
		}
		if (*Node)(curr).key >= key {
			return pred, curr
		}
		pred = curr
		curr = succ
	}
}
