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

package redblacktree

import (
    "fmt"
    "reflect"
    "testing"
)

var funcs map[string]reflect.Method

func init() {
    var found bool
    var put, rotateLeft, rotateRight reflect.Method

    t := reflect.TypeOf(NewTree())
    put, found = t.MethodByName("Put")
    if !found {
        panic("No method `Put` in Tree")
    }
    rotateLeft, found = t.MethodByName("RotateLeft")
    if !found {
        panic("No method `RotateLeft` in Tree")
    }
    rotateRight, found = t.MethodByName("RotateRight")
    if !found {
        panic("No method `RotateRight` in Tree")
    }

    funcs = map[string]reflect.Method{
        "rotateRight": rotateRight,
        "rotateLeft": rotateLeft,
        "put": put,
    }

    TraceOff()
    fmt.Println("done .. redblacktree_test.init")
}

// using the inorder walk of the tree for equality
func assertEqualTree(tr *Tree, t *testing.T, expected string) {
    visitor := &InorderVisitor{}
    tr.Walk(visitor)
    if visitor.String() != expected {
        t.Errorf("Expected [ %s ] got [ %s ]", expected, visitor)
    }
}

type KV struct {
    key int
    arg string
}

// @TODO rename to display arity
// @TODO can I refactor to combine with ToArgs2
func ToArgs(t *Tree, kv KV) []reflect.Value {
    in := make([]reflect.Value, 3)
    in[0] = reflect.ValueOf(t)
    in[1] = reflect.ValueOf(kv.key)
    in[2] = reflect.ValueOf(kv.arg)
    return in
}

func ToArgs2(t *Tree, arg interface{}) []reflect.Value {
    in := make([]reflect.Value, 2)
    in[0] = reflect.ValueOf(t)
    in[1] = reflect.ValueOf(arg)
    return in
}

var fixtureSmall = []struct {
    ops string
    kv  KV
    expected string
}{
    {"put", KV{7, "payload7"}, "(.7.)"},
    {"put", KV{3, "payload3"}, "((.3.)7.)"},
    {"put", KV{8, "payload8"}, "((.3.)7(.8.))"},
}

// depth of tree <= 1
func TestRedBlackSmall(t *testing.T) {
    // empty tree
    t1 := NewTree()
    if t1.root != nil {
        t.Errorf("root starts out as nil but got (%t)", t1.root != nil)
    }
    assertEqualTree(t1, t, ".")

    t2 := NewTree()
    for _, tt := range fixtureSmall {
        method := funcs[tt.ops]
        in := ToArgs(t2, tt.kv)
        method.Func.Call(in) // @TODO returns []reflect.Value ?
        assertEqualTree(t2, t, tt.expected)
    }
}

var fixtureSimpleRightRotation = []struct {
    ops string
    kv  KV
    expected string
}{
    {"put", KV{7, "payload7"}, "(.7.)"},
    {"put", KV{3, "payload3"}, "((.3.)7.)"},
    {"put", KV{1, "payload1"}, "((.1.)3(.7.))"},
}

func TestRedBlackSimpleRightRotation(t *testing.T) {
    tr := NewTree()
    for _, tt := range fixtureSimpleRightRotation {
        method := funcs[tt.ops]
        method.Func.Call(ToArgs(tr, tt.kv))
        assertEqualTree(tr, t, tt.expected)
    }
}

var fixtureCase1 = []struct {
    ops string
    kv  KV
    expected string
}{
    {"put", KV{7, "payload7"}, "(.7.)"},
    {"put", KV{8, "payload8"}, "(.7(.8.))"},
    {"put", KV{9, "payload9"}, "((.7.)8(.9.))"},
    {"put", KV{11, "payload11"}, "((.7.)8(.9(.11.)))"},
    {"put", KV{10, "payload10"}, "((.7.)8((.9.)10(.11.)))"},
}

func TestRedBlackCase1(t *testing.T) {
    tr := NewTree()
    for _, tt := range fixtureCase1 {
        method := funcs[tt.ops]
        method.Func.Call(ToArgs(tr, tt.kv))
        assertEqualTree(tr, t, tt.expected)
    }
}

var fixtureRotationLeft = []struct {
    ops string
    kv  KV
    expected string
}{
    {"put", KV{7, "payload7"}, "(.7.)"},
    {"rotateLeft", KV{}, "(.7.)"},
    {"put", KV{8, "payload8"}, "(.7(.8.))"},
    {"put", KV{9, "payload9"}, "((.7.)8(.9.))"},
    {"put", KV{11, "payload11"}, "((.7.)8(.9(.11.)))"},
    {"rotateLeft", KV{}, "(((.7.)8.)9(.11.))"},
    {"rotateLeft", KV{}, "((((.7.)8.)9.)11.)"},
    {"rotateLeft", KV{}, "((((.7.)8.)9.)11.)"},
}

func TestRedBlackLeft(t *testing.T) {
    t1 := NewTree()
    assertEqualTree(t1, t, ".")
    t1.RotateLeft(t1.root)
    assertEqualTree(t1, t, ".")

    for _, tt := range fixtureRotationLeft {
        method := funcs[tt.ops]
        switch {
        case tt.ops == "put":
            method.Func.Call(ToArgs(t1, tt.kv))

        case tt.ops == "rotateLeft":
            method.Func.Call(ToArgs2(t1, t1.root))
        }
        assertEqualTree(t1, t, tt.expected)
    }
}

var fixtureRotationRight = []struct {
    ops string
    kv  KV
    expected string
}{
    {"put", KV{7, "payload7"}, "(.7.)"},
    {"rotateRight", KV{}, "(.7.)"},
    {"put", KV{3, "payload3"}, "((.3.)7.)"},
    {"put", KV{2, "payload2"}, "((.2.)3(.7.))"},
    {"put", KV{1, "payload1"}, "(((.1.)2.)3(.7.))"},
    {"rotateRight", KV{}, "((.1.)2(.3(.7.)))"},
    {"rotateRight", KV{}, "(.1(.2(.3(.7.))))"},
    {"rotateRight", KV{}, "(.1(.2(.3(.7.))))"},
}

func TestRedBlackRight(t *testing.T) {
    t1 := NewTree()
    assertEqualTree(t1, t, ".")
    t1.RotateRight(t1.root)
    assertEqualTree(t1, t, ".")

    for _, tt := range fixtureRotationRight {
        method := funcs[tt.ops]
        switch {
        case tt.ops == "put":
            method.Func.Call(ToArgs(t1, tt.kv))

        case tt.ops == "rotateRight":
            method.Func.Call(ToArgs2(t1, t1.root))
        }
        assertEqualTree(t1, t, tt.expected)
    }
}
