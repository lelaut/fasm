package main

import (
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
			res, err := Run(path.Join(EXAMPLE_FILENAME, item.Name()))
			dat, err := os.ReadFile(path.Join(EXAMPLE_FILENAME, item.Name()+".out"))
			if err != nil {
				t.Fatal(err)
			}
			out := strings.Split(string(dat), "\n")

			for io, o := range out {
				o = strings.TrimSpace(o)
				if o != res[io].ToString() {
					t.Errorf("\nExpected: '%v'\nReceived: '%v'", o, res[io].ToString())
				}
			}
		}
	}
}
