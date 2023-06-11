package lib

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	r := bufio.NewReader(os.Stdin)

	var in string
	for {
		var err error
		in, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				return "", err
			}
		}
		in = strings.TrimSpace(in)
		if len(in) > 0 {
			break
		}
	}

	fmt.Println("")
	return in, nil
}
