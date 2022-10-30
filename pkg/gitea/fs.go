package gitea

import (
	"io"
	"io/fs"
	"time"
)

type fileInfo struct {
	size  int64
	isdir bool
	name  string
}

type openFile struct {
	content []byte
	offset  int64
	name    string
	isdir   bool
}

func (g fileInfo) Name() string {
	return g.name
}

func (g fileInfo) Size() int64 {
	return g.size
}

func (g fileInfo) Mode() fs.FileMode {
	return 0o444
}

func (g fileInfo) ModTime() time.Time {
	return time.Time{}
}

func (g fileInfo) Sys() any {
	return nil
}

func (g fileInfo) IsDir() bool {
	return g.isdir
}

var _ io.Seeker = (*openFile)(nil)

func (o *openFile) Close() error {
	return nil
}

func (o *openFile) Stat() (fs.FileInfo, error) {
	return fileInfo{
		size:  int64(len(o.content)),
		isdir: o.isdir,
		name:  o.name,
	}, nil
}

func (o *openFile) Read(b []byte) (int, error) {
	if o.offset >= int64(len(o.content)) {
		return 0, io.EOF
	}

	if o.offset < 0 {
		return 0, &fs.PathError{Op: "read", Path: o.name, Err: fs.ErrInvalid}
	}

	n := copy(b, o.content[o.offset:])

	o.offset += int64(n)

	return n, nil
}

func (o *openFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		offset += 0
	case 1:
		offset += o.offset
	case 2:
		offset += int64(len(o.content))
	}

	if offset < 0 || offset > int64(len(o.content)) {
		return 0, &fs.PathError{Op: "seek", Path: o.name, Err: fs.ErrInvalid}
	}

	o.offset = offset

	return offset, nil
}
