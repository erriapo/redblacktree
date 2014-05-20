[![Build Status](https://secure.travis-ci.org/erriapo/redblacktree.png)](https://travis-ci.org/erriapo/redblacktree) 

## redblacktree

A [redblack tree](http://en.wikipedia.org/wiki/Red%E2%80%93black_tree). [![GoDoc](https://godoc.org/github.com/erriapo/redblacktree?status.png)](https://godoc.org/github.com/erriapo/redblacktree)

## Usage

### Default tree expects keys to be of `int` type. 

```go
import (
    "fmt"
    rbt "github.com/erriapo/redblacktree"
)

func main() {
    t := rbt.NewTree()
    t.Put(7, "payload7")
    t.Put(3, "payload3")
    t.Put(1, "payload1")
    fmt.Printf("size = %d\n", t.Size()) // size = 3

    inorder := &rbt.InorderVisitor{}; t.Walk(inorder)
    fmt.Printf("tree = %s\n", inorder) // tree = ((.1.)3(.7.))

    if ok, payload := t.Get(3); ok {
        fmt.Printf("%d is mapped to %s\n", 3, payload.(string))
    }

    fmt.Printf("t.Has(1) = %t\n", t.Has(1)) // true
    t.Delete(1)
    fmt.Printf("t.Has(1) = %t\n", t.Has(1)) // false
}
```

### Example of a tree with custom keys

```go
type Key struct {
    Path, Country string
}

func KeyComparator(o1, o2 interface{}) int {
    k1 := o1.(Key); k2 := o2.(Key)
    return rbt.StringComparator(k1.Path + k1.Country, k2.Path + k2.Country)
}

func main() {
    tree := rbt.NewTreeWith(KeyComparator)
    kAU, kNZ := Key{"/", "au"}, Key{"/tmp", "nz"}
    tree.Put(kAU, 999)
    if ok, payload := tree.Get(kAU); ok {
        fmt.Printf("%#v is mapped to %#v\n", kAU, payload)
        // main.Key{Path:"/", Country:"au"} is mapped to 999
    }
    tree.Put(kNZ, 666)
    fmt.Printf("size = %d\n", tree.Size()) // size = 2
}
```

## License

* Code is released under Apache license. See [LICENSE][license] file.

[license]: https://github.com/erriapo/redblacktree/blob/master/LICENSE
