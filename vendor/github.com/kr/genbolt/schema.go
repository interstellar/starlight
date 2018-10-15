package main

import (
	"go/ast"
	"go/types"
	"text/template"
)

type schema struct {
	InputFile string
	Package   string
	Imports   map[string]string // path -> local name
	Keys      map[string]bool   // all statically-known db keys

	// Bucket types containing static keys. These appear in
	// the schema definition as 'struct { ... }'.
	StructTypes []*schemaStruct

	// Bucket types containing dynamic keys. These appear in
	// the schema as '[]T' or 'map[string]T'.
	//
	// The map key in Go is the bucket name and the local
	// type name of container e.g. 'MapOfFoo' or 'SeqOfFoo'.
	// The map value is the type of the container's element.
	MapOfBucketTypes map[string]string
	SeqOfBucketTypes map[string]string
	MapOfRecordTypes map[string]types.Type
	SeqOfRecordTypes map[string]types.Type

	// Fields pointing to single records, including plain
	// Go data types like int64 and []byte, and types that
	// satisfy json.Marshaler.
	RecordFields []*schemaField

	// Fields pointing to buckets.
	BucketFields []*schemaField

	funcs template.FuncMap
}

type schemaStruct struct {
	Name   string
	IsRoot bool
	Doc    *ast.CommentGroup
}

type schemaField struct {
	Name   string // field name
	Type   types.Type
	Bucket string // containing struct's type
	Doc    *ast.CommentGroup
}
