package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/benbjohnson/genesis"
)

func main() {
	if err := run(os.Args[1:]); err == flag.ErrHelp {
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Parse flags.
	fs := flag.NewFlagSet("genesis", flag.ContinueOnError)
	cwd := fs.String("C", "", "")
	out := fs.String("o", "", "")
	pkg := fs.String("pkg", "", "")
	tags := fs.String("tags", "", "")
	fs.Usage = usage
	if err := fs.Parse(args); err != nil {
		return err
	} else if fs.NArg() == 0 {
		usage()
		return flag.ErrHelp
	} else if *pkg == "" {
		return errors.New("package name required")
	}

	// Change working directory, if specified.
	if *cwd != "" {
		if err := os.Chdir(*cwd); err != nil {
			return err
		}
	}

	// Find all matching files.
	var paths []string
	for _, arg := range fs.Args() {
		a, err := expandPath(arg)
		if err != nil {
			return err
		}
		paths = append(paths, a...)
	}

	// Determine output writer.
	var w io.Writer
	if *out == "" {
		w = os.Stdout
	} else {
		f, err := os.Create(*out)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	// Write generated file.
	if err := genesis.WriteHeader(w, *pkg, *tags); err != nil {
		return err
	} else if err := genesis.WriteAssetNames(w, paths); err != nil {
		return err
	} else if err := genesis.WriteAssetMap(w, paths); err != nil {
		return err
	} else if err := genesis.WriteFileType(w); err != nil {
		return err
	} else if err := genesis.WriteAssetFuncs(w); err != nil {
		return err
	} else if err := genesis.WriteFileSystem(w); err != nil {
		return err
	} else if err := genesis.WriteHashFuncs(w); err != nil {
		return err
	}
	return nil
}

// expandPath converts path into a list of all files within path.
// If path is a file then path is returned.
func expandPath(path string) ([]string, error) {
	if fi, err := os.Stat(path); err != nil {
		return nil, err
	} else if !fi.IsDir() {
		return []string{path}, nil
	}

	// Read files in directory.
	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// Iterate over files and expand.
	expanded := make([]string, 0, len(fis))
	for _, fi := range fis {
		a, err := expandPath(filepath.Join(path, fi.Name()))
		if err != nil {
			return nil, err
		}
		expanded = append(expanded, a...)
	}
	return expanded, nil
}

func usage() {
	fmt.Print(`usage: genesis [options] path [paths]

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

`)
}
