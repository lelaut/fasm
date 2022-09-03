package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"
)

const EXAMPLE_FILENAME = "examples"

func TestExamples(t *testing.T) {
	items, err := os.ReadDir(EXAMPLE_FILENAME)
	if err != nil {
		t.Fatal(err)
	}

	for _, item := range items {
		if strings.HasSuffix(item.Name(), ".asm") {
			input := path.Join(EXAMPLE_FILENAME, item.Name()+".in")
			if _, err := os.Stat(input); errors.Is(err, os.ErrNotExist) {
				input = ""
			}

			fmt.Println("Running", item.Name())
			res, err := Run(path.Join(EXAMPLE_FILENAME, item.Name()), input)
			if err != nil {
				t.Error(err)
			}
			dat, err := os.ReadFile(path.Join(EXAMPLE_FILENAME, item.Name()+".out"))
			if err != nil {
				t.Fatal(err)
			}
			out := strings.Split(string(dat), "\n")

			for io, o := range out {
				o = strings.TrimSpace(o)
				if len(res) <= io {
					t.Errorf("\nExpected: '%v'\nDidn't received anything", o)
				}
				if o != res[io].ToString() {
					t.Errorf("\nExpected: '%v'\nReceived: '%v'", o, res[io].ToString())
				}
			}
		}
	}
}
