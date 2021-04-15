package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golife/goscript/exec"
)

var gsExt = ".gs"

func fatal(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

func printVersion() {
	fmt.Fprintln(os.Stderr, "Goscript Version 0.1.0")
}

func main() {
	flag.Usage = func() {
		printVersion()
		fmt.Fprintln(os.Stderr, "\nUsage of:", os.Args[0])
		fmt.Fprintln(os.Stderr, os.Args[0], "[flags] <filename>")
		flag.PrintDefaults()
	}
	var (
		ver = flag.Bool("version", false, "Print version number and exit")
	)
	flag.Parse()

	if *ver {
		printVersion()
		os.Exit(1)
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Arg(0)

	if filepath.Ext(filename) != gsExt {
		fatal(`Goscript source files should have the '.${gsExt}' extension`)
	}

	src, err := ioutil.ReadFile(filename)
	if err != nil {
		fatal(err)
	}
	fmt.Println("Compiling:", filename)

	filename = filename[:len(filename)-len(gsExt)]
	ret := exec.Exec(filename, string(src))
	fmt.Println(ret)
}
