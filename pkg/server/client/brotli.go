package client

import (
	"io"

	"github.com/andybalholm/brotli"
)

func decompressBrotli(r io.ReadCloser) (io.ReadCloser, error) {
	br := &brotliReader{s: r, r: brotli.NewReader(r)}

	return br, nil
}

type brotliReader struct {
	s io.ReadCloser
	r *brotli.Reader
}

func (b *brotliReader) Read(p []byte) (int, error) {
	return b.r.Read(p)
}

func (b *brotliReader) Close() error {
	return b.s.Close()
}
