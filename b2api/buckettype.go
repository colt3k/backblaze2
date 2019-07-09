package b2api

// BucketType enum type
type BucketType int

const (
	All BucketType = 1 + iota
	AllPublic
	AllPrivate
	Snapshot
)

var bucketTypes = [...]string{
	"all", "allPublic", "allPrivate", "snapshot",
}

// String output type
func (bt BucketType) String() string {
	return bucketTypes[bt-1]
}

// Types show all types
func (bt BucketType) Types() []string {
	return bucketTypes[:]
}

// Id show int value
func (bt BucketType) Id(val string) int {
	for i, d := range bucketTypes {
		if d == val {
			return i + 1
		}
	}
	return -1
}
