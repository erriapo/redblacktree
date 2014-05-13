package main

import (
    "fmt"
    rbtree "github.com/erriapo/redblacktree"
)

func main() {
    t := rbtree.NewTree()
    t.Put(7, "payload7")
    t.Put(3, "payload3")
    t.Put(1, "payload1")
    fmt.Println("size = 3") // @TODO size API
}
