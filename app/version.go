package app

//go:generate mockgen -source $GOFILE -destination mock_$GOFILE -package $GOPACKAGE

import (
	"fmt"
	"time"
)

var versionMetadata VersionMetadata = VersionMetadata{
	BranchName: "unknown",
	Revision:   "unknown",
	BuiltBy:    "unknown",
}

type Version interface {
	Describe() string

	Version() string
	Metadata() VersionMetadata
}

type VersionMetadata struct {
	BranchName string
	Revision   string
	BuiltAt    time.Time
	BuiltBy    string
}

type versionT struct {
	name     string
	version  string
	metadata VersionMetadata
}

func (v versionT) Describe() string {
	return fmt.Sprintf(
		"%s version %s (%s @ %s, built at %s by %s)",
		v.name,
		v.version,
		v.metadata.BranchName,
		v.metadata.Revision,
		v.metadata.BuiltAt.Format(time.RFC3339),
		v.metadata.BuiltBy,
	)
}

func (v versionT) Version() string {
	return v.version
}

func (v versionT) Metadata() VersionMetadata {
	return v.metadata
}

// Sets the version metadata for the build. One pattern for setting
// version metadata is to provide a generated file in the project's
// main package that constructs a VersionMetadata object and invokes
// this method.
//
// Note that go build tags (see https://golang.org/pkg/go/build/) may
// be used to provide compile-time alternates for release builds.
func SetVersionMetadata(metadata VersionMetadata) {
	versionMetadata = metadata
}
