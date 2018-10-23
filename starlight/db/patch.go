package db

// TODO(kr): delete this file
//
// This file patches the ClientUpdateSeq type to
// provide access to its bucket because genbolt
// doesn't do that yet.

func (r *Root) DeleteAgent() {
	r.db.DeleteBucket(keyAgent)
}
