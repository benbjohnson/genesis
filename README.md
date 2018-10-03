genesis [![Go Report Card](https://goreportcard.com/badge/github.com/benbjohnson/genesis)](https://goreportcard.com/report/github.com/benbjohnson/genesis) [![](https://godoc.org/github.com/benbjohnson/genesis?status.svg)](http://godoc.org/github.com/benbjohnson/genesis)
=======

Genesis is a utility for embedding static file assets into Go files to be
included in static compilation. Genesis includes support for `http.FileSystem`
as well versioning via SHA1 hashes.


## Installation

To install `genesis`, first install [Go](https://golang.org/) and then run:

```sh
$ go get -u github.com/benbjohnson/genesis/...
```

You should now have a `genesis` binary in your `$GOPATH/bin` directory.


## Usage

### Command line usage

The `genesis` command includes a few basic options available from the CLI:

```
usage: genesis [options] path [paths]

Embeds listed assets in a Go file as hex-encoded strings.

The following flags are available:

	-pkg name
		Package name of the generated Go file. Required.

	-o output
		Output filename for generated code. Optional.
		(default stdout)

	-C dir
		Execute genesis from dir. Optional.

	-tags tags
		Comma-delimited list of build tags. Optional.

```

You can generate a new Go file with your embedded assets by specifying the
package name and the list of files or directories you want to include. By
default the output will go to `STDOUT` so you can redirect to the file of your
choice.

```sh
$ genesis -pkg mypkg file1 file2 dir1 > assets.gen.go
```


#### Specifying the output file

You can also specify the output file using the `-o` flag. This is useful if you
are running via `//go:generate` and file redirection is not available:

```sh
$ genesis -pkg mypkg -o assets.gen.go file1 file2 dir1
```


#### Working directory

One common approach to asset organization is to create a separate `assets`
package that includes your asset files and your generated embedded file. You
can use the `-C` flag to change the present working directory to this package
and then execute relative to this directory:

```sh
$ genesis -C assets -pkg mypkg -o assets.gen.go file1 file2
```


#### Including build tags

Finally, you can optionally include build tags to optionally include assets
in your final build. Tags should be comma separated for each individual build
tag line:

```
$ genesis -pkg mypkg -o assets.386.gen.go -tags "linux darwin,386" file1 file2 dir1
```

This will output the following build tag comments at the top of the file:

```go
// +build linux darwin
// +build 386
```


### HTTP File System

Your generated embedded assets Go file will include an implementation of
[`http.FileSystem`](https://golang.org/pkg/net/http/#FileSystem) that can be
passed to [`http.FileServer`](https://golang.org/pkg/net/http/#FileServer) to automatically serve your embedded assets
through an [`http.Handler`](https://golang.org/pkg/net/http/#Handler):

```go
http.Handle("/", http.FileServer(assets.FileSystem()))
```

#### Cache control

Your embedded assets Go file also includes a method for returning the filename
with an included SHA1 hash. This hash will change whenever the contents of the
file changes so the `http.FileSystem` will tell the browser to cache the file
indefinitely using the `Cache-Control` header.

To use this feature, simply use the `AssetNameWithHash()` function when writing
your URLs that reference embedded assets:


```go
fmt.Fprintf(w, `<script src="%s"></script>`, assets.AssetNameWithHash("bundle.js"))
```

This will output something like the following:

```html
<script src="/bundle-25318a5755cba4f4147fcb2a535ba1caaebade1a.js"></script>
```

When a file is requested with a SHA1 hash it will be cached indefinitely which
leads to improved web site latency on return visits.


### Programmatic API

The generated embedded assets Go file includes several types and utility
functions. You can view this information via the `godoc` for your package.
A list of some of these types and functions are included below.


#### `File`

One `File` is generated for every asset you embed. This struct provides the
name, content hash, size, last modification time, and raw data.

```go
type File struct {}

func (f *File) Name() string
func (f *File) Hash() string
func (f *File) ModTime() time.Time
func (f *File) Data() []byte
```



#### Asset functions

Several utility functions are provided for accessing asset information. The
`Asset()` function simply returns the underlying asset data while `AssetFile()`
returns the entire `File` object. To retrieve a sorted list of all asset names,
use the `AssetNames()` function.

```go
func Asset(name string) []byte
func AssetFile(name string) *File
func AssetNames() []string
```


#### Content hash functions

The `AssetNameWithHash()` function will return a given asset name with its SHA1
content hash included. This is useful for embedding in HTML source names when
combining with the included `http.FileSystem` implementation.

```go
func AssetNameWithHash(name string) string
```

There are also utility functions for working with asset names and hashes. The
`JoinNameHash()` function will build an asset name hash with the given name
and hash. The `SplitNameHash()` function will return the base asset name and
an asset name hash. The `HasNameHash()` returns true if a given name
includes an asset hash.

```go
func JoinNameHash(name, hash string) string
func SplitNameHash(name string) string 
func HasNameHash(name string) bool
```

