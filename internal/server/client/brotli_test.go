package client

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
)

func TestDecompressBrotli(t *testing.T) {
	t.Parallel()

	originalData := "This is a test string for Brotli compression."

	var compressedData bytes.Buffer
	writer := brotli.NewWriter(&compressedData)

	_, err := writer.Write([]byte(originalData))
	if err != nil {
		t.Fatalf("Failed to write compressed data: %v", err)
	}

	writer.Close()

	compressedReader := io.NopCloser(&compressedData)

	decompressedReader, err := decompressBrotli(compressedReader)
	if err != nil {
		t.Fatalf("decompressBrotli returned an error: %v", err)
	}
	defer decompressedReader.Close()

	decompressedData, err := io.ReadAll(decompressedReader)
	if err != nil {
		t.Fatalf("Failed to read decompressed data: %v", err)
	}

	if string(decompressedData) != originalData {
		t.Errorf("Decompressed data does not match original data. Got: %s, Want: %s", string(decompressedData), originalData)
	}
}

func TestBrotliReader_Close(t *testing.T) {
	t.Parallel()

	dummyReader := io.NopCloser(strings.NewReader("dummy data"))

	br := &brotliReader{s: dummyReader, r: brotli.NewReader(dummyReader)}

	if err := br.Close(); err != nil {
		t.Errorf("brotliReader.Close() returned an error: %v", err)
	}
}
