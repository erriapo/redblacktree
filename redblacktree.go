/*
Copyright 2014 Gavin Bong.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the
License.
*/

// Package redblacktree provides a pure Golang implementation 
// of a red-black tree as described by Thomas H. Cormen's et al. 
// in their seminal Algorithms book (3rd ed).
package redblacktree

import (
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "reflect"
    "strings"
    "sync"
)

// Color of a redblack tree node is either 
// `Black` (true) & `Red` (false)
type Color bool

// Direction points to either the Left or Right subtree
type Direction byte

func (c Color) String() string {
    switch c {
    case true:
        return "Black"
    default:
        return "Red"
    }
}

func (d Direction) String() string {
    switch d {
    case LEFT:
        return "left"
    case RIGHT:
        return "right"
    case NODIR:
        return "center"
    default:
        return "not recognized"
    }
}

const (
    BLACK, RED Color     = true, false
    LEFT       Direction = iota
    RIGHT
    NODIR
)

// A node needs to be able to answer the query:
// - Who is my parent node ?
// - Who is my grandparent node ?
// The zero value for Node has color Red.
type Node struct {
    value  int
    color  Color
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

// A redblack tree is `Visitable` by a `Visitor`.
type Visitable interface {
    Walk(Visitor)
}

// Tree encapsulates the data structure.
type Tree struct {
    root *Node
}

// `lock` protects `logger`
var lock sync.Mutex
var logger *log.Logger

func init() {
    logger = log.New(ioutil.Discard, "", log.LstdFlags)
}

// TraceOn turns on logging output to Stderr
func TraceOn() {
    SetOutput(os.Stderr)
}

// TraceOff turns off logging.
// By default logging is turned off.
func TraceOff() {
    SetOutput(ioutil.Discard)
}

// SetOutput redirects log output
func SetOutput(w io.Writer) {
    lock.Lock()
    defer lock.Unlock()
    logger = log.New(w, "", log.LstdFlags)
}

// NewTree returns an empty Tree
func NewTree() *Tree {
    return &Tree{root: nil}
}

// Get looks for the node with supplied key and returns its mapped payload
func (t *Tree) Get(key int) (bool, *Node) {
    found, parent, dir := t.GetParent(key)
    if found {
        if parent == nil {
            return true, t.root
        } else {
            var node *Node
            switch dir {
            case LEFT:
                node = parent.left
            case RIGHT:
                node = parent.right
            }

            if node != nil {
                return true, node
            }
        }
    }
    return false, nil
}

// GetParent looks for the node with supplied key and returns the parent node.
func (t *Tree) GetParent(key int) (found bool, parent *Node, dir Direction) {
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
            case LEFT:
                parent.left = newNode
            case RIGHT:
                parent.right = newNode
            }
            logger.Printf("Added %s to %s node of parent %s\n", newNode.String(), dir, parent.String())
            t.fixup(newNode)
        }
    }
}

func isRed(n *Node) bool {
    value := reflect.ValueOf(n)
    if value.IsNil() {
        return false
    } else {
        return n.color == RED
    }
}

func isBlack(n *Node) bool {
    if n == nil {
        return true
    }
    value := reflect.ValueOf(n)
    if value.IsNil() {
        return true
    } else {
        return n.color == BLACK
    }
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
        logger.Printf("\tcurrent z %s\n", z.String())
        switch {
        case z.parent == nil:
            fallthrough
        case z.parent.color == BLACK:
            fallthrough
        default:
            // When the loop terminates, it does so because p[z] is black.
            logger.Printf("\t\t=> bye\n")
            break loop
        case z.parent.color == RED:
            grandparent := z.parent.parent
            if z.parent == grandparent.left {
                logger.Printf("\t\t%s is the left child of %s\n", z.parent, grandparent)
                y := grandparent.right
                logger.Printf("\t\ty (right) %s\n", y)
                if isRed(y) {
                    // case 1 - y is RED
                    logger.Printf("\t\t(*) case 1\n")
                    z.parent.color = BLACK
                    y.color = BLACK
                    grandparent.color = RED
                    z = grandparent

                } else {
                    if z == z.parent.right {
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
                logger.Printf("\t\ty (left) %s\n", y)
                if isRed(y) {
                    // case 1 - y is RED
                    logger.Printf("\t\t..(*) case 1\n")
                    z.parent.color = BLACK
                    y.color = BLACK
                    grandparent.color = RED
                    z = grandparent

                } else {
                    logger.Printf("\t\t## %s\n", z.parent.left)
                    if z == z.parent.left {
                        // case 2
                        logger.Printf("\t\t..(*) case 2\n")
                        z = z.parent
                        t.RotateRight(z)
                    }

                    // case 3
                    logger.Printf("\t\t..(*) case 3\n")
                    z.parent.color = BLACK
                    grandparent.color = RED
                    t.RotateLeft(grandparent)
                }
            }
        }
    }
    t.root.color = BLACK
}

// Walk accepts a Visitor
func (t *Tree) Walk(visitor Visitor) {
    visitor.Visit(t.root)
}

// InorderVisitor walks the tree in inorder fashion.
// This visitor maintains internal state; thus do not
// reuse after the completion of a walk.
type InorderVisitor struct {
    buffer bytes.Buffer
}

func (v *InorderVisitor) Eq(other *InorderVisitor) bool {
    if other == nil {
        return false
    }
    return v.String() == other.String()
}

func (v *InorderVisitor) trim(s string) string {
    return strings.TrimRight(strings.TrimRight(s, "ed"), "lack")
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
    //v.buffer.Write([]byte(fmt.Sprintf("%d{%s}", node.value, v.trim(node.color.String()))))
    v.Visit(node.right)
    v.buffer.Write([]byte(")"))
}

func main() {
    // example manual tree construction
    node1 := Node{value: 10, left: &Node{value: 8}, right: &Node{value: 11}}
    node2 := Node{value: 22, right: &Node{value: 26}}
    tree := Tree{root: &Node{value: 7, left: &Node{value: 3}, right: &Node{value: 18, left: &node1, right: &node2}}}
    visitor := &InorderVisitor{}
    tree.Walk(visitor)
}
