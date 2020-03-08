package store

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/config"
	"io/ioutil"
)

var magicNumber = []byte{0x1f, 0x8b}

func encodeResource(res *config.Resource) (string, error) {
	data, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	_, err = writer.Write(data)
	if err != nil {
		return "", err
	}
	writer.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func decodeResource(data string) (*config.Resource, error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// verify gzip format data by magic number
	if bytes.Equal(decodedData[0:2], magicNumber) {
		reader := bytes.NewReader(decodedData)
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(gzipReader)
		if err != nil {
			return nil, err
		}
		var res config.Resource
		err = json.Unmarshal(b, &res)
		if err != nil {
			return nil, err
		}
		return &res, nil
	} else {
		return nil, fmt.Errorf("resource format error: not gzip format")
	}
}
