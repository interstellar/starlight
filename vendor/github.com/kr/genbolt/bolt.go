package main

// Without this import statement, genbolt itself doesn't
// depend on bbolt, even though the generated code does.
// This lets us express a version reqirement on bbolt
// without it showing up as "indirect" in go.mod.
import _ "github.com/coreos/bbolt"
