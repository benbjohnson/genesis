package genesis_test

import (
	"bytes"
	"errors"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/benbjohnson/genesis"
)

func TestEncoder_Encode(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		var buf bytes.Buffer
		enc := genesis.NewEncoder(&buf)
		enc.Package = "mypkg"

		if err := enc.Encode(&genesis.Asset{Name: "/a.txt", Data: []byte("abc"), ModTime: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
			t.Fatal(err)
		} else if err := enc.Encode(&genesis.Asset{Name: "/a/b.txt", Data: []byte("foobar"), ModTime: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
			t.Fatal(err)
		} else if err := enc.Encode(&genesis.Asset{Name: "c/d/e", Data: []byte("testing123"), ModTime: time.Date(2002, time.January, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
			t.Fatal(err)
		}

		if err := enc.Close(); err != nil {
			t.Fatal(err)
		} else if b := buf.Bytes(); !bytes.Equal(b, testdata("assets.txt")) {
			t.Fatalf("unexpected output, see: %s", tempfile(b))
		}

		if err := ensureGoFormatted(buf.Bytes()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Empty", func(t *testing.T) {
		var buf bytes.Buffer
		enc := genesis.NewEncoder(&buf)
		enc.Package = "mypkg"
		if err := enc.Close(); err != nil {
			t.Fatal(err)
		} else if b := buf.Bytes(); !bytes.Equal(b, testdata("assets.empty.txt")) {
			t.Fatalf("unexpected output, see: %s", tempfile(b))
		}
	})

	t.Run("BuildTags", func(t *testing.T) {
		var buf bytes.Buffer
		enc := genesis.NewEncoder(&buf)
		enc.Package = "mypkg"
		enc.Tags = []string{"linux darwin", "386"}
		if err := enc.Close(); err != nil {
			t.Fatal(err)
		} else if b := buf.Bytes(); !bytes.Equal(b, testdata("assets.build.txt")) {
			t.Fatalf("unexpected output, see: %s", tempfile(b))
		}
	})
}

func testdata(name string) []byte {
	buf, err := ioutil.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		panic(err)
	}
	return buf
}

func tempfile(data []byte) string {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	} else if _, err := f.Write(data); err != nil {
		panic(err)
	} else if err := f.Close(); err != nil {
		panic(err)
	}
	return f.Name()
}

func ensureGoFormatted(src []byte) error {
	buf, err := format.Source(src)
	if err != nil {
		return err
	} else if !bytes.Equal(buf, src) {
		return errors.New("source not gofmt formatted")
	}
	return nil
}
