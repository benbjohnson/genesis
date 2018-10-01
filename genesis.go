package genesis

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func WriteAssetNames(w io.Writer, paths []string) error {
	if _, err := fmt.Fprintln(w, `var assetNames = []string{`); err != nil {
		return err
	}

	for _, path := range paths {
		if _, err := fmt.Fprintf(w, "	%q,\n", PrependSlash(filepath.ToSlash(path))); err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w, "}\n\n")
	return err
}

func WriteAssetMap(w io.Writer, paths []string) error {
	if _, err := fmt.Fprintln(w, `var assetMap = map[string]*File{`); err != nil {
		return err
	}

	for _, path := range paths {
		if err := WriteAsset(w, path); err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(w, "}\n\n")
	return err
}

func WriteAsset(w io.Writer, path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Generate stringified hex data.
	var hexdata bytes.Buffer
	if _, err := io.Copy(&hexWriter{&hexdata}, bytes.NewReader(data)); err != nil {
		return err
	}

	// Calculate mod time parts.
	sec := fi.ModTime().UnixNano() / int64(time.Second)
	nsec := fi.ModTime().UnixNano() % int64(time.Second)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "	%q: &File{\n", PrependSlash(filepath.ToSlash(path)))
	fmt.Fprintf(&buf, "		name:    %q,\n", PrependSlash(filepath.ToSlash(path)))
	fmt.Fprintf(&buf, "		hash:    \"%x\",\n", sha1.Sum(data))
	fmt.Fprintf(&buf, "		size:    %d,\n", len(data))
	fmt.Fprintf(&buf, "		modTime: time.Unix(%d, %d),\n", sec, nsec)
	fmt.Fprintf(&buf, "		data:    []byte(\"%s\"),\n", hexdata.String())
	fmt.Fprintf(&buf, "	},\n")
	_, err = buf.WriteTo(w)
	return err
}

func WriteFileType(w io.Writer) error {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, `type File struct {`)
	fmt.Fprintln(&buf, `	name    string`)
	fmt.Fprintln(&buf, `	hash    string`)
	fmt.Fprintln(&buf, `	size    int64`)
	fmt.Fprintln(&buf, `	modTime time.Time`)
	fmt.Fprintln(&buf, `	data    []byte`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `func (f *File) Name() string       { return f.name }`)
	fmt.Fprintln(&buf, `func (f *File) Hash() string       { return f.hash }`)
	fmt.Fprintln(&buf, `func (f *File) Size() int64        { return f.size }`)
	fmt.Fprintln(&buf, `func (f *File) ModTime() time.Time { return f.modTime }`)
	fmt.Fprintln(&buf, `func (f *File) Data() []byte       { return f.data }`)
	fmt.Fprintln(&buf, ``)
	_, err := buf.WriteTo(w)
	return err
}

func WriteAssetFuncs(w io.Writer) error {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, `func Asset(name string) []byte {`)
	fmt.Fprintln(&buf, `	if f := AssetFile(name); f != nil {`)
	fmt.Fprintln(&buf, `		return f.Data()`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, `	return nil`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	fmt.Fprintln(&buf, `func AssetFile(name string) *File {`)
	fmt.Fprintln(&buf, `	if f := assetMap[name]; f != nil {`)
	fmt.Fprintln(&buf, `		return f`)
	fmt.Fprintln(&buf, `	} else if f := assetMap[TrimNameHash(name)]; f != nil {`)
	fmt.Fprintln(&buf, `		return f`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, `	return nil`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	fmt.Fprintln(&buf, `func AssetNames() []string {`)
	fmt.Fprintln(&buf, `	return assetNames`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	fmt.Fprintln(&buf, `func AssetNameWithHash(name string) string {`)
	fmt.Fprintln(&buf, `	if f := AssetFile(name); f != nil {`)
	fmt.Fprintln(&buf, `		return JoinNameHash(f.Name(), f.Hash())`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, `	return name`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	_, err := buf.WriteTo(w)
	return err
}

func WriteFileSystem(w io.Writer) error {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, `func FileSystem() http.FileSystem { return &httpFileSystem{} }`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `type httpFileSystem struct{}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `func (fs *httpFileSystem) Open(name string) (http.File, error) {`)
	fmt.Fprintln(&buf, `	f := AssetFile(name)`)
	fmt.Fprintln(&buf, `	if f == nil {`)
	fmt.Fprintln(&buf, `		return nil, &os.PathError{Path: "/" + name, Err: os.ErrNotExist}`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, `	return newHTTPFile(f), nil`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	fmt.Fprintln(&buf, `type httpFileServer struct{}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `func FileServer() http.Handler {`)
	fmt.Fprintln(&buf, `	return &httpFileServer{}`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `func (h *httpFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {`)
	fmt.Fprintln(&buf, `	name := r.URL.Path`)
	fmt.Fprintln(&buf, `	if !strings.HasPrefix(name, "/") {`)
	fmt.Fprintln(&buf, `		name = "/" + name`)
	fmt.Fprintln(&buf, `		r.URL.Path = name`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `	f := AssetFile(path.Clean(name))`)
	fmt.Fprintln(&buf, `	if f == nil {`)
	fmt.Fprintln(&buf, `		http.Error(w, "404 page not found", http.StatusNotFound)`)
	fmt.Fprintln(&buf, `		return`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `	if HasNameHash(name) {`)
	fmt.Fprintln(&buf, `		w.Header().Set("Cache-Control", "max-age=31536000")`)
	fmt.Fprintln(&buf, `	}`)
	fmt.Fprintln(&buf, `	http.ServeContent(w, r, f.Name(), f.ModTime(), newHTTPFile(f))`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	fmt.Fprintln(&buf, `func newHTTPFile(f *File) *httpFile {`)
	fmt.Fprintln(&buf, `	return &httpFile{File: f, Reader: bytes.NewReader(f.data)}`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `type httpFile struct {`)
	fmt.Fprintln(&buf, `	*File`)
	fmt.Fprintln(&buf, `	*bytes.Reader`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)
	fmt.Fprintln(&buf, `func (f *httpFile) Close() error               { return nil }`)
	fmt.Fprintln(&buf, `func (f *httpFile) Stat() (os.FileInfo, error) { return f, nil }`)
	fmt.Fprintln(&buf, `func (f *httpFile) Size() int64                { return f.File.Size() }`)
	fmt.Fprintln(&buf, `func (f *httpFile) Mode() os.FileMode          { return 0444 }`)
	fmt.Fprintln(&buf, `func (f *httpFile) ModTime() time.Time         { return time.Time{} }`)
	fmt.Fprintln(&buf, `func (f *httpFile) IsDir() bool                { return false }`)
	fmt.Fprintln(&buf, `func (f *httpFile) Sys() interface{}           { return nil }`)
	fmt.Fprintln(&buf, `func (f *httpFile) Readdir(count int) ([]os.FileInfo, error) {`)
	fmt.Fprintln(&buf, `	return nil, &os.PathError{Path: "/" + f.name, Err: os.ErrPermission}`)
	fmt.Fprintln(&buf, `}`)
	fmt.Fprintln(&buf, ``)

	_, err := buf.WriteTo(w)
	return err
}

func WriteHashFuncs(w io.Writer) error {
	var buf bytes.Buffer
	fmt.Fprintln(w, `func JoinNameHash(name, hash string) string {`)
	fmt.Fprintln(w, `	dir, file := path.Split(name)`)
	fmt.Fprintln(w, `	if i := strings.Index(file, "."); i != -1 {`)
	fmt.Fprintln(w, `		return path.Join(dir, file[0:i]+"-"+hash+file[i:])`)
	fmt.Fprintln(w, `	}`)
	fmt.Fprintln(w, `	return name + "-" + hash`)
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w, ``)

	fmt.Fprintln(w, `func TrimNameHash(name string) string {`)
	fmt.Fprintln(w, `	dir, file := path.Split(name)`)
	fmt.Fprintln(w, `	pre, post := file, ""`)
	fmt.Fprintln(w, `	if i := strings.Index(file, "."); i != -1 {`)
	fmt.Fprintln(w, `		pre, post = file[0:i], file[i:]`)
	fmt.Fprintln(w, `	}`)
	fmt.Fprintln(w, `	if len(pre) < 41 || pre[len(pre)-41] != '-' || !hashRegex.MatchString(pre[len(pre)-40:]) {`)
	fmt.Fprintln(w, `		return name`)
	fmt.Fprintln(w, `	}`)
	fmt.Fprintln(w, `	return path.Join(dir, pre[:len(pre)-41]+post)`)
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w, ``)

	fmt.Fprintln(w, `func HasNameHash(name string) bool {`)
	fmt.Fprintln(w, `	_, file := path.Split(name)`)
	fmt.Fprintln(w, `	if i := strings.Index(file, "."); i != -1 {`)
	fmt.Fprintln(w, `		file = file[0:i]`)
	fmt.Fprintln(w, `	}`)
	fmt.Fprintln(w, `	return len(file) >= 41 && file[len(file)-41] == '-' && hashRegex.MatchString(file[len(file)-40:])`)
	fmt.Fprintln(w, `}`)
	fmt.Fprintln(w, ``)

	fmt.Fprintln(w, `var hashRegex = regexp.MustCompile("^[0-9a-f]+$")`)

	_, err := buf.WriteTo(w)
	return err
}

func PrependSlash(s string) string {
	if strings.HasPrefix(s, "/") {
		return s
	}
	return "/" + s
}

// WriteHeader writes the header comment, package name, build tags, & import to w.
func WriteHeader(w io.Writer, pkg, tags string) error {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "// Code generated by genesis.")
	fmt.Fprintln(&buf, "// DO NOT EDIT.")
	fmt.Fprintln(&buf, "")

	// Write build tags.
	if tags != "" {
		fmt.Fprintf(&buf, "// +build %s\n\n", tags)
		fmt.Fprintln(&buf, "")
	}

	fmt.Fprintf(&buf, "package %s", pkg)
	fmt.Fprintln(&buf, "")

	// Write imports.
	fmt.Fprintln(&buf, "")
	fmt.Fprintln(&buf, `import (`)
	fmt.Fprintln(&buf, `	"bytes"`)
	fmt.Fprintln(&buf, `	"net/http"`)
	fmt.Fprintln(&buf, `	"os"`)
	fmt.Fprintln(&buf, `	"path"`)
	fmt.Fprintln(&buf, `	"regexp"`)
	fmt.Fprintln(&buf, `	"strings"`)
	fmt.Fprintln(&buf, `	"time"`)
	fmt.Fprintln(&buf, `)`)
	fmt.Fprintln(&buf, "")

	_, err := buf.WriteTo(w)
	return err
}

// hexWriter converts all writes to \x00 format.
type hexWriter struct {
	w io.Writer
}

func (w *hexWriter) Write(p []byte) (n int, err error) {
	const hex = "0123456789abcdef"
	for _, b := range p {
		var buf [4]byte
		buf[0] = '\\'
		buf[1] = 'x'
		buf[2] = hex[b>>4]
		buf[3] = hex[b&0x0F]

		if _, err := w.w.Write(buf[:]); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
