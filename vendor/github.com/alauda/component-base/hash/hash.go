package hash

import (
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/davecgh/go-spew/spew"
)

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	// copy from k8s.io/kubernetes/pkg/util/hash/hash.go to avoid import this monster
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}

func genHash(data interface{}) uint64 {
	hasher := fnv.New64()
	DeepHashObject(hasher, data)
	return hasher.Sum64()
}

//GenHashStr ... This is a very stupid function, need to learn from kubernetes at how to
// generate pod name for deployment
func GenHashStr(data interface{}) string {
	s := fmt.Sprintf("%d", genHash(data))
	return s
}
