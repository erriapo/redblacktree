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
// in their seminal Algorithms book (3rd ed). This data structure
// is not multi-goroutine safe.
package redblacktree

import (
    "bytes"
    "errors"
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
// (i) Who is my parent node ?
// (ii) Who is my grandparent node ?
// The zero value for Node has color Red.
type Node struct {
    key     interface{}
    payload interface{}
    color  Color
    left   *Node
    right  *Node
    parent *Node
}

func (n *Node) String() string {
    return fmt.Sprintf("(%#v : %s)", n.key, n.Color())
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

// Keys must be comparable. It's mandatory to provide a Comparator,
// which returns zero if o1 == o2, -1 if o1 < o2, 1 if o1 > o2
type Comparator func(o1, o2 interface{}) int

// Default comparator expects keys to be of type `int`.
// Warning: if either one of `o1` or `o2` cannot be asserted to `int`, it panics.
func IntComparator(o1, o2 interface{}) int {
    i1 := o1.(int); i2 := o2.(int)
    switch {
    case i1 > i2:
        return 1
    case i1 < i2:
        return -1
    default:
        return 0
    }
}

// Tree encapsulates the data structure.
type Tree struct {
    root *Node
    cmp Comparator
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

// NewTree returns an empty Tree with default comparator `IntComparator`.
// `IntComparator` expects keys to be type-assertable to `int`.
func NewTree() *Tree {
    return &Tree{root: nil, cmp: IntComparator}
}

// NewTreeWith returns an empty Tree with a supplied `Comparator`.
func NewTreeWith(c Comparator) *Tree {
    return &Tree{root: nil, cmp: c}
}

// Get looks for the node with supplied key and returns its mapped payload.
// Return value in 1st position indicates whether any payload was found.
func (t *Tree) Get(key interface{}) (bool, interface{}) {
    if err := mustBeValidKey(key); err != nil {
        logger.Printf("Get was prematurely aborted: %s\n", err.Error())
        return false, nil
    }

    ok, node := t.getNode(key)
    if ok {
        return true, node.payload
    } else {
        return false, nil
    }
}

func (t *Tree) getNode(key interface{}) (bool, *Node) {
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

// getMinimum returns the node with minimum key starting
// at the subtree rooted at node x. Assume x is not nil.
func (t *Tree) getMinimum(x *Node) *Node {
    for {
        if x.left != nil {
            x = x.left
        } else {
            return x
        }
    }
}

// GetParent looks for the node with supplied key and returns the parent node.
func (t *Tree) GetParent(key interface{}) (found bool, parent *Node, dir Direction) {
    if err := mustBeValidKey(key); err != nil {
        logger.Printf("GetParent was prematurely aborted: %s\n", err.Error())
        return false, nil, NODIR
    }

    if t.root == nil {
        return false, nil, NODIR
    }

    return t.internalLookup(nil, t.root, key, NODIR)
}

func (t *Tree) internalLookup(parent *Node, this *Node, key interface{}, dir Direction) (bool, *Node, Direction) {
    switch {
    case this == nil:
        return false, parent, dir
    case t.cmp(key, this.key) == 0:
        return true, parent, dir
    case t.cmp(key, this.key) < 0:
        return t.internalLookup(this, this.left, key, LEFT)
    case t.cmp(key, this.key) > 0:
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

// Put saves the mapping (key, data) into the tree.
// If a mapping identified by `key` already exists, it is overwritten.
// Constraint: Not everything can be a key.
func (t *Tree) Put(key interface{}, data interface{}) error {
    if err := mustBeValidKey(key); err != nil {
        logger.Printf("Put was prematurely aborted: %s\n", err.Error())
        return err
    }

    if t.root == nil {
        t.root = &Node{key: key, color: BLACK}
        logger.Printf("Added %s as root node\n", t.root.String())
        return nil
    }

    found, parent, dir := t.internalLookup(nil, t.root, key, NODIR)
    if found {
        if parent == nil {
            logger.Printf("Put: parent=nil & found. Overwrite ROOT node\n")
            t.root.payload = data
        } else {
            logger.Printf("Put: parent!=nil & found. Overwriting\n")
            switch dir {
            case LEFT:
                parent.left.payload = data
            case RIGHT:
                parent.right.payload = data
            }
        }

    } else {
        if parent != nil {
            newNode := &Node{key: key, parent: parent, payload: data}
            switch dir {
            case LEFT:
                parent.left = newNode
            case RIGHT:
                parent.right = newNode
            }
            logger.Printf("Added %s to %s node of parent %s\n", newNode.String(), dir, parent.String())
            t.fixupPut(newNode)
        }
    }
    return nil
}

func isRed(n *Node) bool {
    key := reflect.ValueOf(n)
    if key.IsNil() {
        return false
    } else {
        return n.color == RED
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
func (t *Tree) fixupPut(z *Node) {
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
            logger.Printf("\t\tgrandparent is nil %t\n", grandparent == nil)
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

// Size returns the number of items in the tree.
func (t *Tree) Size() uint64 {
    visitor := &countingVisitor{}
    t.Walk(visitor)
    return visitor.Count
}

// Has checks for existence of a item identified by supplied key.
func (t *Tree) Has(key interface{}) bool {
    if err := mustBeValidKey(key); err != nil {
        logger.Printf("Has was prematurely aborted: %s\n", err.Error())
        return false
    }
    found, _, _ := t.internalLookup(nil, t.root, key, NODIR)
    return found
}

func (t *Tree) transplant(u *Node, v *Node) {
    if u.parent == nil {
        t.root = v
    } else if u == u.parent.left {
        u.parent.left = v
    } else {
        u.parent.right = v
    }
    if v != nil && u != nil {
        v.parent = u.parent
    }
}

// Delete removes the item identified by the supplied key.
// Delete is a noop if the supplied key doesn't exist.
func (t *Tree) Delete(key interface{}) {
    if !t.Has(key) {
        logger.Printf("Delete: bail as no node exists for key %d\n", key)
        return
    }
    _, z := t.getNode(key)
    logger.Printf("Delete: attempt to delete %s\n", z)
    y := z
    yOriginalColor := y.color
    var x *Node

    if z.left == nil {
        // one child (RIGHT)
        logger.Printf("\t\tDelete: case (a)\n")
        x = z.right
        logger.Printf("\t\t\t--- x is right of z")
        t.transplant(z, z.right)

    } else if z.right == nil {
        // one child (LEFT)
        logger.Printf("\t\tDelete: case (b)\n")
        x = z.left
        logger.Printf("\t\t\t--- x is left of z")
        t.transplant(z, z.left)

    } else {
        // two children
        logger.Printf("\t\tDelete: case (c) & (d)\n")
        y = t.getMinimum(z.right)
        logger.Printf("\t\t\tminimum of z.right is %s (color=%s)\n", y, y.color)
        yOriginalColor = y.color
        x = y.right
        logger.Printf("\t\t\t--- x is right of minimum")

        if y.parent == z {
            if x != nil {
                x.parent = y
            }
        } else {
            t.transplant(y, y.right)
            y.right = z.right
            y.right.parent = y
        }
        t.transplant(z, y)
        y.left = z.left
        y.left.parent = y
        y.color = z.color
    }
    if yOriginalColor == BLACK {
        t.fixupDelete(x)
    }
}

func (t *Tree) fixupDelete(x *Node) {
    logger.Printf("\t\t\tfixupDelete of node %s\n", x)
    if x == nil {
        return
    }
loop:
    for {
        switch {
        case x == t.root:
            logger.Printf("\t\t\t=> bye .. is root\n")
            break loop
        case x.color == RED:
            logger.Printf("\t\t\t=> bye .. RED\n")
            break loop
        case x == x.parent.right:
            logger.Printf("\t\tBRANCH: x is right child of parent\n")
            w := x.parent.left // is nillable
            if isRed(w) {
                // Convert case 1 into case 2, 3, or 4
                logger.Printf("\t\t\tR> case 1\n")
                w.color = BLACK
                x.parent.color = RED
                t.RotateRight(x.parent)
                w = x.parent.left
            }
            if w != nil {
                switch {
                case !isRed(w.left) && !isRed(w.right):
                    // case 2 - both children of w are BLACK
                    logger.Printf("\t\t\tR> case 2\n")
                    w.color = RED
                    x = x.parent // recurse up tree
                case isRed(w.right) && !isRed(w.left):
                    // case 3 - right child RED & left child BLACK
                    // convert to case 4
                    logger.Printf("\t\t\tR> case 3\n")
                    w.right.color = BLACK
                    w.color = RED
                    t.RotateLeft(w)
                    w = x.parent.left
                }
                if isRed(w.left) {
                    // case 4 - left child is RED
                    logger.Printf("\t\t\tR> case 4\n")
                    w.color = x.parent.color
                    x.parent.color = BLACK
                    w.left.color = BLACK
                    t.RotateRight(x.parent)
                    x = t.root
                }
            }
        case x == x.parent.left:
            logger.Printf("\t\tBRANCH: x is left child of parent\n")
            w := x.parent.right // is nillable
            if isRed(w) {
                // Convert case 1 into case 2, 3, or 4
                logger.Printf("\t\t\tL> case 1\n")
                w.color = BLACK
                x.parent.color = RED
                t.RotateLeft(x.parent)
                w = x.parent.right
            }
            if w != nil {
                switch {
                case !isRed(w.left) && !isRed(w.right):
                    // case 2 - both children of w are BLACK
                    logger.Printf("\t\t\tL> case 2\n")
                    w.color = RED
                    x = x.parent // recurse up tree
                case isRed(w.left) && !isRed(w.right):
                    // case 3 - left child RED & right child BLACK
                    // convert to case 4
                    logger.Printf("\t\t\tL> case 3\n")
                    w.left.color = BLACK
                    w.color = RED
                    t.RotateRight(w)
                    w = x.parent.right
                }
                if isRed(w.right) {
                    // case 4 - right child is RED
                    logger.Printf("\t\t\tL> case 4\n")
                    w.color = x.parent.color
                    x.parent.color = BLACK
                    w.right.color = BLACK
                    t.RotateLeft(x.parent)
                    x = t.root
                }
            }
        }
    }
    x.color = BLACK
}

// Walk accepts a Visitor
func (t *Tree) Walk(visitor Visitor) {
    visitor.Visit(t.root)
}

// countingVisitor counts the number
// of nodes in the tree.
type countingVisitor struct {
    Count uint64
}

func (v *countingVisitor) Visit(node *Node) {
    if node == nil {
        return
    }

    v.Visit(node.left)
    v.Count = v.Count + 1
    v.Visit(node.right)
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
    v.buffer.Write([]byte(fmt.Sprintf("%d", node.key))) // @TODO
    //v.buffer.Write([]byte(fmt.Sprintf("%d{%s}", node.key, v.trim(node.color.String()))))
    v.Visit(node.right)
    v.buffer.Write([]byte(")"))
}

var (
    ErrorKeyIsNil = errors.New("The literal nil not allowed as keys")
    ErrorKeyDisallowed = errors.New("Disallowed key type")
)

// Allowed key types are: Boolean, Integer, Floating point, Complex, String values
// And structs containing these. 
// @TODO Should pointer type be allowed ?
func mustBeValidKey(key interface{}) error {
    if key == nil {
        return ErrorKeyIsNil
    }

    keyValue := reflect.ValueOf(key)
    switch keyValue.Kind() {
    case reflect.Chan:
        fallthrough
    case reflect.Func:
        fallthrough
    case reflect.Interface:
        fallthrough
    case reflect.Map:
        fallthrough
    case reflect.Ptr:
        fallthrough
    case reflect.Slice:
        return ErrorKeyDisallowed
    default:
        return nil
    }
}

func main() {
    // example manual tree construction
    node1 := Node{key: 10, left: &Node{key: 8}, right: &Node{key: 11}}
    node2 := Node{key: 22, right: &Node{key: 26}}
    tree := Tree{root: &Node{key: 7, left: &Node{key: 3}, right: &Node{key: 18, left: &node1, right: &node2}}}
    visitor := &InorderVisitor{}
    tree.Walk(visitor)
}
