//
// nazuna :: util.go
//
//   Copyright (c) 2013-2022 Akinori Hattori <hattya@gmail.com>
//
//   SPDX-License-Identifier: MIT
//

package nazuna

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
)

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func IsEmptyDir(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()
	_, err = f.Readdir(1)
	return err == io.EOF
}

func SplitPath(path string) (string, string) {
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

func marshal(repo *Repository, path string, v interface{}) error {
	rel, err := filepath.Rel(repo.root, path)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("%v: %v", rel, err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o666); err != nil {
		return fmt.Errorf("cannot write '%v'", rel)
	}
	return nil
}

func unmarshal(repo *Repository, path string, v interface{}) error {
	rel, err := filepath.Rel(repo.root, path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read '%v'", rel)
		}
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("%v: %v", rel, err)
		}
	}
	return nil
}

func httpGet(uri string) ([]byte, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%v: %v", uri, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}
