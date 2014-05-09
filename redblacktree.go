// A red-black tree implementation of Cormen's Algorithms book
package redblacktree

import (
    "fmt"
    "bytes"
    "io"
    "io/ioutil"
    "log"
    "os"
    "sync"
)

type Color bool
type Direction byte

// Black (true) & Red (false)
func (c Color) String() string {
    switch c {
    case true: return "Black"
    default: return "Red"
    }
}

func (d Direction) String() string {
    switch d {
    case LEFT: return "left"
    case RIGHT: return "right"
    case NODIR: return "center"
    default: return "not recognized"
    }
}

const (
    BLACK, RED Color = true, false
    LEFT Direction = iota
    RIGHT
    NODIR
)

// A node needs to be able to answer the query:
// - Who is my parent node ?
// - Who is my grandparent node ?
// The zero value for Node has color Red.
type Node struct {
    value   int
    color   Color
    left   *Node
    right  *Node
    parent *Node
}

func (n *Node) String() string {
    return fmt.Sprintf("(%#v : %s)", n.value, n.Color())
}

func (n *Node) Parent() *Node {
    return n.parent
}

func (n *Node) SetColor(color Color) {
    n.color = color
}

func (n *Node) Color() Color {
    return n.color
}

type Visitor interface {
    Visit(*Node)
}

type Visitable interface {
    Walk(Visitor)
}

type Tree struct {
    root *Node
}

// This mutex protects `logger`
var lock sync.Mutex
var logger *log.Logger

func init() {
    logger = log.New(ioutil.Discard, "", log.LstdFlags)
    fmt.Println("done .. redblacktree.init")
}

func TraceOn() {
    SetOutput(os.Stderr)
}

func TraceOff() {
    SetOutput(ioutil.Discard)
}

// Redirect log output
func SetOutput(w io.Writer) {
    lock.Lock()
    defer lock.Unlock()
    logger = log.New(w, "", log.LstdFlags)
}

func NewTree() *Tree {
    return &Tree{root: nil}
}

// @TODO Rename GetParent ?
func (t *Tree) Lookup(key int) (found bool, parent *Node, dir Direction) {
    if t.root == nil {
        return false, nil, NODIR
    }

    return t.internalLookup(nil, t.root, key, NODIR)
}

func (t *Tree) internalLookup(parent *Node, this *Node, key int, dir Direction) (bool, *Node, Direction) {
    switch {
    case this == nil:
        return false, parent, dir
    case this.value == key:
        return true, parent, dir
    case key < this.value:
        return t.internalLookup(this, this.left, key, LEFT)
    case key > this.value:
        return t.internalLookup(this, this.right, key, RIGHT)
    default:
        return false, parent, NODIR
    }
}

// Reverses actions of RotateLeft
func (t *Tree) RotateRight(y *Node) {
    if y == nil {
        logger.Printf("RotateRight: nil arg cannot be rotated. Noop\n")
        return
    }
    if y.left == nil {
        logger.Printf("RotateRight: y has nil left subtree. Noop\n")
        return
    }
    logger.Printf("\t\t\trotate right of %s\n", y)
    x := y.left
    y.left = x.right
    if x.right != nil {
        x.right.parent = y
    }
    x.parent = y.parent
    if y.parent == nil {
        t.root = x
    } else {
        if y == y.parent.left {
            y.parent.left = x
        } else {
            y.parent.right = x
        }
    }
    x.right = y
    y.parent = x
}

// Side-effect: red-black tree properties is maintained.
func (t *Tree) RotateLeft(x *Node) {
    if x == nil {
        logger.Printf("RotateLeft: nil arg cannot be rotated. Noop\n")
        return
    }
    if x.right == nil {
        logger.Printf("RotateLeft: x has nil right subtree. Noop\n")
        return
    }
    logger.Printf("\t\t\trotate left of %s\n", x)

    y := x.right
    x.right = y.left
    if y.left != nil {
        y.left.parent = x
    }
    y.parent = x.parent
    if x.parent == nil {
        t.root = y
    } else {
        if x == x.parent.left {
            x.parent.left = y
        } else {
            x.parent.right = y
        }
    }
    y.left = x
    x.parent = y
}

// Insert an element
// @TODO follow Java's TreeMap's Put(key K, value V). Should I return the V ? or an error
func (t *Tree) Put(key int, data interface{}) {
    if t.root == nil {
        t.root = &Node{value: key, color: BLACK}
        logger.Printf("Added %s as root node\n", t.root.String())
        return
    }

    found, parent, dir := t.internalLookup(nil, t.root, key, NODIR)
    if found {
        if parent == nil {
            fmt.Println("Parent nil and found. Overwrite ROOT NODE")
        } else {
            fmt.Println("Parent not nil and found")
            switch dir {
            case LEFT:
                fallthrough
            case RIGHT:
                logger.Printf("\tOverwrite %s node of parent %d\n", dir, parent.value)
            }
        }

    } else {
        if parent == nil {
            logger.Printf("???Parent nil and not found in tree")
        } else {
            newNode := &Node{value: key, parent: parent}
            switch dir {
            case LEFT: parent.left = newNode
            case RIGHT: parent.right = newNode
            }
            logger.Printf("Added %s to %s node of parent %s\n", newNode.String(), dir, parent.String())
            t.fixup(newNode)
        }
    }
}

func isBlack(n *Node) bool {
    return n == nil || n.color == BLACK
}

// fix possible violations of red-black-tree properties 
// with combinations of:
// 1. recoloring
// 2. rotations
//
// Preconditions:
// P1) z is not nil
// 
// @param z - the newly added Node to the tree.
func (t *Tree) fixup(z *Node) {
    logger.Printf("\tfixup new node z %s\n", z.String())
loop:
    for {
        switch {
        case z.parent == nil:
            fallthrough
        case z.parent.color == BLACK:
            fallthrough
        default:
            // When the loop terminates, it does so because p[z] is black.
            logger.Printf("\t\t=> bye\n");
            break loop
        case z.parent.color == RED:
            logger.Printf("\t\t=> do something here\n")

            grandparent := z.parent.parent
            if z.parent == grandparent.left {
                logger.Printf("\t\t%s is the left child of %s\n", z.parent, grandparent)
                y := grandparent.right
                if !isBlack(y) {
                    // case 1 - y is RED
                    logger.Printf("\t\t(*) case 1\n")
                    z.parent.color = BLACK
                    y.color = BLACK
                    grandparent.color = RED
                    z = grandparent

                } else {
                    if (z == z.parent.right) {
                        // case 2
                        logger.Printf("\t\t(*) case 2\n")
                        z = z.parent
                        t.RotateLeft(z)
                    }

                    // case 3
                    logger.Printf("\t\t(*) case 3\n")
                    z.parent.color = BLACK
                    grandparent.color = RED
                    t.RotateRight(grandparent)
                }
            } else {
                logger.Printf("\t\t%s is the right child of %s\n", z.parent, grandparent)
                y := grandparent.left
                if !isBlack(y) {
                    // case 1 - y is RED
                    logger.Printf("\t\t(*) case 1\n")
                    z.parent.color = BLACK
                    y.color = BLACK
                    grandparent.color = RED
                    z = grandparent

                } else {
                    if (z == z.parent.left) {
                        // case 2
                        logger.Printf("\t\t(*) case 2\n")
                        z = z.parent
                        t.RotateRight(z)
                    }

                    // case 3
                    logger.Printf("\t\t(*) case 3\n")
                    z.parent.color = BLACK
                    grandparent.color = RED
                    t.RotateLeft(grandparent)
                }
            }

            break loop
        }
    }
    t.root.color = BLACK
}

// Tree is Visitable
func (t *Tree) Walk(visitor Visitor) {
    visitor.Visit(t.root)
}

func (t *Tree) PrintRoot() {
    if t.root != nil {
        logger.Printf("\tRoot %s\n", t.root)
    } else {
        logger.Printf("\tRoot is nil\n")
    }
}

// @deprecated
func (t *Tree) Inorder() {
    if t.root != nil {
        t.root.visit()
    }
}

// @deprecated
func (n *Node) visit() {
    if n.left != nil {
        n.left.visit()
    }
    logger.Printf("%d ", n.value)
    if n.right != nil {
        n.right.visit()
    }
}

// Inorder traversal of tree
type InorderVisitor struct{
    buffer bytes.Buffer
}

func (v *InorderVisitor) Eq(other *InorderVisitor) bool {
    if other == nil {
        return false
    }
    return v.String() == other.String()
}

func (v *InorderVisitor) String() string {
    return v.buffer.String()
}

func (v *InorderVisitor) Visit(node *Node) {
    if node == nil {
        v.buffer.Write([]byte("."))
        return
    }
    v.buffer.Write([]byte("("))
    v.Visit(node.left)
    v.buffer.Write([]byte(fmt.Sprintf("%d", node.value)))
    v.Visit(node.right)
    v.buffer.Write([]byte(")"))
}

func main() {
    node1 := Node{value: 10, left: &Node{value:8}, right: &Node{value:11}}
    node2 := Node{value: 22, right: &Node{value:26}}
    tree := Tree{root: &Node{value:7, left: &Node{value:3}, right: &Node{value:18, left:&node1, right: &node2}}}
    tree.Inorder()
    fmt.Println()
    i0 := InorderVisitor{}
    tree.Walk(&i0)
    fmt.Println(i0.String())

    fmt.Println("1===========")
    fmt.Println("Direction", LEFT, RIGHT, NODIR)
    find(tree, 8)
    find(tree, 7)
    find(tree, 3)
    find(tree, 1)
    find(tree, 4)
    find(tree, 23)
    find(tree, 29)
    find(tree, 18)
    find(tree, 11)
    find(tree, 22)
    find(tree, 26)

    fmt.Println("2==============")
    tree2 := NewTree()
    find(*tree2, 4)

    fmt.Println("3==============")
    tree3 := NewTree()
    tree3.Put(7, "payload")
    tree3.Put(3, "payload")
    tree3.Put(18, "payload")
    tree3.Put(1, "payload")
    tree3.Put(10, "payload")
    tree3.Put(11, "payload")
    tree3.Put(22, "payload")
    tree3.Put(8, "payload")
    tree3.Put(26, "payload")
    i1 := &InorderVisitor{}
    tree3.Walk(i1)
    fmt.Println("i1 ->", i1)
    tree3.PrintRoot()
    tree3.Put(7, "payload")
    tree3.Put(3, "payload")
    tree3.Put(18, "payload")

    tree3.RotateLeft(nil)

    tree3.RotateLeft(tree3.root)
    i2 := &InorderVisitor{}
    tree3.Walk(i2)
    fmt.Println("i2 ->", i2)
    tree3.PrintRoot()

    tree3.RotateLeft(tree3.root)
    i3 := &InorderVisitor{}
    tree3.Walk(i3)
    fmt.Println("i3 ->", i3)
    tree3.PrintRoot()

    tree3.RotateLeft(tree3.root)
    i4 := &InorderVisitor{}
    tree3.Walk(i4)
    fmt.Println("i4 ->", i4)
    tree3.PrintRoot()

    tree3.RotateLeft(tree3.root)
    i5 := &InorderVisitor{}
    tree3.Walk(i5)
    fmt.Println("i5 ->", i5)
    tree3.PrintRoot()

    log.Printf("\ti4.Eq(i5) is %t\n", i4.Eq(i5))

    tree3.RotateRight(tree3.root)
    i6 := &InorderVisitor{}
    tree3.Walk(i6)
    fmt.Println("i6 ->", i6)
    tree3.PrintRoot()

    log.Printf("\ti6.Eq(i3) is %t\n", i6.Eq(i3))
    log.Printf("\ti6.Eq(i5) is %t\n", i6.Eq(i5))

    tree3.RotateRight(tree3.root)
    i7 := &InorderVisitor{}
    tree3.Walk(i7)
    fmt.Println("i7 ->", i7)
    tree3.PrintRoot()

    log.Printf("\ti7.Eq(i2) is %t\n", i7.Eq(i2))

    tree3.RotateRight(tree3.root)
    i8 := &InorderVisitor{}
    tree3.Walk(i8)
    fmt.Println("i8 ->", i8)
    tree3.PrintRoot()

    log.Printf("\ti8.Eq(i1) is %t\n", i8.Eq(i1))

    f, p, d := tree3.Lookup(10)
    if f {
        if p != nil {
            var node *Node
            switch d {
            case LEFT:
                node = p.left
            case RIGHT:
                node = p.right
            }
            if node != nil {
                tree3.RotateRight(node)
                i9 := &InorderVisitor{}
                tree3.Walk(i9)
                fmt.Println("i9 ->", i9)
                tree3.PrintRoot()
                log.Printf("\ti9.Eq(i8) is %t\n", i9.Eq(i8))
            }
        }
    }

    fmt.Println("4 ................")
    tree4 := NewTree()
    tree4.Put(7, "payload")
    tree4.Put(3, "payload")

    visitor1 := &InorderVisitor{}
    tree4.Walk(visitor1)
    fmt.Println("tree4: visitor1 ->", visitor1)

    tree4.Put(1, "payload")
    visitor2 := &InorderVisitor{}
    tree4.Walk(visitor2)
    fmt.Println("tree4: visitor2 ->", visitor2)

    fmt.Println("5 ................")
    tree5 := NewTree()
    tree5.Put(7, "payload")
    tree5.Put(8, "payload")

    visitor3 := &InorderVisitor{}
    tree5.Walk(visitor3)
    fmt.Println("tree5: visitor3 ->", visitor3)

    tree5.Put(9, "payload")
    tree5.Put(11, "payload")
    tree5.Put(10, "payload")
    visitor4 := &InorderVisitor{}
    tree5.Walk(visitor4)
    fmt.Println("tree5: visitor4 ->", visitor4)

    fmt.Println("6 ................")
    tree6 := NewTree()
    tree6.Put(7, "payload")
    tree6.Put(3, "payload")
    tree6.Put(9, "payload")

    visitor5 := &InorderVisitor{}
    tree6.Walk(visitor5)
    fmt.Println("tree6: visitor5 ->", visitor5)

    tree6.Put(1, "payload")
    tree6.Put(2, "payload")
    visitor6 := &InorderVisitor{}
    tree6.Walk(visitor6)
    fmt.Println("tree6: visitor6 ->", visitor6)
}

func find(tree Tree, key int) {
    found, parent, dir := tree.Lookup(key)
    if found {
        if parent != nil {
            log.Printf("Parent %d for %d in direction %s\n", parent.value, key, dir)
        } else {
            log.Printf("Found %d as the root node in direction %s\n", key, dir)
        }

    } else {
        log.Printf("%d not found in tree\n", key)
        if parent != nil {
            log.Printf("\tInsert it as the %s node of %d\n", dir, parent.value)
        } else {
            log.Printf("\tInsert it as root node in direction %s\n", dir)
        }
    }
}
