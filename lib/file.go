package lib

import "os"

type Metadata struct {
	Name string `json:"name"`
	Size uint64 `json:"size"`
}

type Document struct {
	*Metadata
	File     *os.File
	Packet   []byte
}
