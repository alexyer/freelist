Implementation of Lock-free list in Go.
Current implementation uses strings as items.
It could easily be enhanced to work with any type.

Uses [taggedptr](https://github.com/alexyer/taggedptr) to atomically work with marked pointers.
