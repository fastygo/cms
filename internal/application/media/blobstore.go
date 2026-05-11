package media

import (
	"context"
	"io"
)

type BlobStore interface {
	Put(context.Context, BlobWrite) (BlobObject, error)
	Open(context.Context, BlobRef) (io.ReadCloser, BlobInfo, error)
	Delete(context.Context, BlobRef) error
}

type BlobWrite struct {
	Filename    string
	ContentType string
	SizeBytes   int64
	Body        io.Reader
	Scope       BlobScope
}

type BlobObject struct {
	Ref  BlobRef
	Info BlobInfo
}

type BlobRef struct {
	Provider string
	Key      string
	URL      string
}

type BlobInfo struct {
	ContentType string
	SizeBytes   int64
	Checksum    string
}

type BlobScope string

const (
	BlobScopeMediaOriginal BlobScope = "media-original"
	BlobScopeMediaVariant  BlobScope = "media-variant"
	BlobScopeSandbox       BlobScope = "sandbox"
)
