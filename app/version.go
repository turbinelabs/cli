package app

//go:generate mockgen -source $GOFILE -destination mock_$GOFILE -package $GOPACKAGE

import "fmt"

type Version interface {
	Print()
}

type versionT struct {
	name    string
	version string
}

func (v versionT) Print() {
	fmt.Println(v.name, "version", v.version)
}
