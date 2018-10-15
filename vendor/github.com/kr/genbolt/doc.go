/*

Command genbolt generates code for conveniently
reading and writing objects in a bolt database.
It reads a set of Go type definitions
describing the layout of data in a bolt database,
and generates code for reading and writing that data.

The following types are stored as records:

	string
	fixed-size basic types (bool, int8, uint8, int16, etc)
	slices of fixed-size basic types
	named types satisfying BinaryMarshaler (see below)
	named types satisfying json.Marshaler (see below)

The following types are stored as buckets:

	named types defined in the schema's package
	slices of any of these types (record or bucket)
	map[string]T where T is any of these types

All other types are not supported.

For example, here is a schema definition:

	package db

	type Root struct {
		Users  []*User
		Config *Config
	}

	type User struct {
		Name string
	}

	type Config struct {
		RateLimit int64
	}

Here, Root is the root bucket.
Field Users leads to a bucket indexed by
an automatically incrementing uint64,
holding all user records.
Type User is a bucket representing a single user.
Field Config leads to the single Config bucket,
holding a single number.

This schema produces the following package interface
(with some definitions elided):

	func (o *Root) Config() *Config
	func (o *Root) Users() *SeqOfUser

	func (o *Config) PutRateLimit(v int64)
	func (o *Config) RateLimit() int64

	func (o *User) Name() string
	func (o *User) PutName(v string)

	func (o *SeqOfUser) Add() (*User, uint64)
	func (o *SeqOfUser) Get(n uint64) *User

Named types from other packages can be used,
provided they're accompanied by
a variable declaration in the schema
asserting that they satisfy either json.Marshaler
or encoding.BinaryMarshaler.
Such types must also satisfy json.Unmarshaler
or encoding.BinaryUnmarshaler, respectively,
but this does not need to appear in the schema.

	var (
		_ json.Marshaler           = (*mypkg.MyType)(nil)
		_ encoding.BinaryMarshaler = (*mypkg.OtherType)(nil)
	)

	type MyBucket struct {
		MyField    *mypkg.MyType
		OtherField *mypkg.MyType
		MySeq      []*mypkg.MyType
		MyMap      map[string]*mypkg.MyType
	}

Values of those types are marshaled when written
and unmarshaled again when read.

It is conventional to put a +build ignore directive
in the schema file, so it can live in the same directory
as the generated code without its symbols conflicting.

*/
package main
