package lib

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

func Encode(obj interface{}) (string, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	var data bytes.Buffer
	gz, err := gzip.NewWriterLevel(&data, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	if _, err := gz.Write([]byte(b)); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	sdp := base64.StdEncoding.EncodeToString(data.Bytes())
	return sdp, nil
}

func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	br := bytes.NewReader(b)
	gz, err := gzip.NewReader(br)
	data, err := ioutil.ReadAll(gz)
	if err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	return json.Unmarshal(data, obj)
}

func MustReadStdin() (string, error) {
	var sdpString string
	_, err := fmt.Scanln(&sdpString)
	if err != nil {
		return "", err
	}
	sdpString = strings.TrimSpace(sdpString)
	return sdpString, nil
}
