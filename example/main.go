package main

import (
    "fmt"
    rbt "github.com/erriapo/redblacktree"
)

type Key struct {
    Path, Country string
}

func KeyComparator(o1, o2 interface{}) int {
    k1 := o1.(Key)
    k2 := o2.(Key)
    return rbt.StringComparator(k1.Path+k1.Country, k2.Path+k2.Country)
}

func main() {
    t := rbt.NewTree()

    fmt.Printf("Starting with empty tree\n")
    fmt.Printf("t.Has(7) = %t\n", t.Has(7)) // false

    fmt.Printf("Add 3 nodes with keys 7, 3 and 1 in succession\n")
    t.Put(7, "payload7")
    t.Put(3, "payload3")
    t.Put(1, "payload1")

    fmt.Printf("size = %d\n", t.Size()) // 3

    inorder := &rbt.InorderVisitor{}
    t.Walk(inorder)
    fmt.Printf("tree = %s\n", inorder) // tree = ((.1.)3(.7.))

    if ok, payload := t.Get(3); ok {
        fmt.Printf("%d is mapped to %s\n", 3, payload.(string))
    }
    fmt.Printf("t.Has(7) = %t\n", t.Has(7)) // true

    t.Delete(1)
    fmt.Printf("\nt.Delete(1)\n")
    inorder2 := &rbt.InorderVisitor{}
    t.Walk(inorder2)
    fmt.Printf("tree = %s\n", inorder2) // tree = (.3(.7.))
    fmt.Printf("t.Has(1) = %t\n\n", t.Has(1)) // false

    tr := rbt.NewTreeWith(KeyComparator)
    kAU, kNZ := Key{"/", "au"}, Key{"/tmp", "nz"}
    tr.Put(kAU, 999)
    if ok, payload := tr.Get(kAU); ok {
        fmt.Printf("%#v is mapped to %#v\n", kAU, payload)
    }
    tr.Put(kNZ, 666)
    fmt.Printf("tr.Put(kNZ, 666)\n")
    fmt.Printf("size = %d\n", tr.Size())
}
