package main

import (
    "fmt"
    rbt "github.com/erriapo/redblacktree"
)

func main() {
    t := rbt.NewTree()
    t.Put(7, "payload7")
    t.Put(3, "payload3")
    t.Put(1, "payload1")
    fmt.Printf("size = %d\n", t.Size())
    if ok, payload := t.Get(3); ok {
        fmt.Printf("%d is mapped to %s\n", 3, payload.(string))
    }
}
