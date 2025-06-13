package storage

import "io"

type FileReader func() io.ReadCloser

type Artifacts map[string]FileReader
