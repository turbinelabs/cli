/*
Copyright 2018 Turbine Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

//go:generate mockgen -source $GOFILE -destination mock_$GOFILE -package $GOPACKAGE --write_package_comment=false

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

// Gets the version metadata for the build. May be useful for forensics.
func GetVersionMetadata() VersionMetadata {
	return versionMetadata
}
