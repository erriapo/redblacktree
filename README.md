[![Build Status](https://secure.travis-ci.org/erriapo/redblacktree.png)](https://travis-ci.org/erriapo/redblacktree)

## redblacktree

A [redblack tree](http://en.wikipedia.org/wiki/Red%E2%80%93black_tree). 

[![GoDoc](https://godoc.org/github.com/erriapo/redblacktree?status.png)](https://godoc.org/github.com/erriapo/redblacktree)

## Usage

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
    fmt.Printf("size = %d\n", t.Size())
    if ok, payload := t.Get(3); ok {
        fmt.Printf("%d is mapped to %s\n", 3, payload.(string))
    }
}
```

## License

* Code is released under Apache license. See [LICENSE][license] file.

[license]: https://github.com/erriapo/redblacktree/blob/master/LICENSE
