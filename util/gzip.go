package util

import (
	"bytes"
	"compress/gzip"
	"io"
)

// gzipCompress gzip压缩
func gzipCompress(in []byte) ([]byte, error) {
	if len(in) == 0 {
		return in, nil
	}

	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)

	if _, err := gz.Write(in); err != nil {
		return nil, err
	}

	if err := gz.Flush(); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// gzipDecompress gzip解压
func gzipDecompress(in []byte) ([]byte, error) {
	if len(in) == 0 {
		return in, nil
	}
	buffer := bytes.NewBuffer(in)
	var r io.Reader
	r, err := gzip.NewReader(buffer)
	if err != nil {
		return nil, err
	}

	var resB bytes.Buffer
	_, err = resB.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return resB.Bytes(), nil
}
