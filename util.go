//
// nazuna :: util.go
//
//   Copyright (c) 2013 Akinori Hattori <hattya@gmail.com>
//
//   Permission is hereby granted, free of charge, to any person
//   obtaining a copy of this software and associated documentation files
//   (the "Software"), to deal in the Software without restriction,
//   including without limitation the rights to use, copy, modify, merge,
//   publish, distribute, sublicense, and/or sell copies of the Software,
//   and to permit persons to whom the Software is furnished to do so,
//   subject to the following conditions:
//
//   The above copyright notice and this permission notice shall be
//   included in all copies or substantial portions of the Software.
//
//   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//   EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
//   MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//   NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS
//   BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
//   ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
//   CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//   SOFTWARE.
//

package nazuna

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func isEmptyDir(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()
	_, err = f.Readdir(1)
	return err == io.EOF
}

func splitPath(path string) (string, string) {
	dir, name := filepath.Split(path)
	dir = strings.TrimRightFunc(dir, func(r rune) bool {
		return r < utf8.RuneSelf && os.IsPathSeparator(uint8(r))
	})
	return dir, name
}

func sortKeys(i interface{}) []string {
	v := reflect.Indirect(reflect.ValueOf(i))
	if v.Kind() != reflect.Map {
		return nil
	}
	keys := v.MapKeys()
	if len(keys) == 0 || keys[0].Kind() != reflect.String {
		return nil
	}
	list := make(sort.StringSlice, len(keys))
	for i, k := range keys {
		list[i] = k.String()
	}
	list.Sort()
	return list
}

func marshal(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return ioutil.WriteFile(path, data, 0666)
}

func unmarshal(path string, v interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func format(s string, m map[string]string) string {
	if strings.Contains(s, "{") {
		for k, v := range m {
			s = strings.Replace(s, "{"+k+"}", v, -1)
		}
	}
	return s
}

func httpGet(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", uri, resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
