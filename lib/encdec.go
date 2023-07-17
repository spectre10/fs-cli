package lib

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// encodes the SDP object into base64 and returns the string
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

// Decodes the base64 string into SDP object.
func Decode(in string, obj interface{}) error {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		return err
	}

	br := bytes.NewReader(b)
	gz, err := gzip.NewReader(br)
	if err != nil {
		panic(err)
	}
	data, err := io.ReadAll(gz)
	if err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	return json.Unmarshal(data, obj)
}

// Reads the SDP from terminal.
// Currently maybe not working on windows.
func ReadSDP() (string, error) {
	// var sdpString string
	// _, err := fmt.Scanln(&sdpString)
	// if err != nil {
	// 	return "", err
	// }
	// sdpString = strings.TrimSpace(sdpString)
	// return sdpString, nil

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
