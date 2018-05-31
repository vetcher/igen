package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"io/ioutil"

	"github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"
	"github.com/vetcher/igen/tpl"
)

var (
	flagTargets     = StringListFlagValue{}
	flagFileName    = flag.String("file", "", "")
	flagMiddlewares = StringListFlagValue{}
)

func init() {
	flag.Var(&flagTargets, "target", "")
	flag.Var(&flagMiddlewares, "mws", "")
	flag.Parse()

	if flag.Arg(0) == "" {
		fmt.Println("input file is requered")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *flagFileName == "" {
		fmt.Println("-file is requered")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if len(flagTargets.L) == 0 {
		fmt.Println("-targets is requered")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if len(flagMiddlewares.L) == 0 {
		fmt.Println("-mws is requered")
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	file, err := astra.ParseFile(flag.Arg(0), astra.AllowAnyImportAliases, astra.IgnoreStructs, astra.IgnoreMethods, astra.IgnoreVariables, astra.IgnoreConstants, astra.IgnoreComments, astra.IgnoreFunctions)
	if err != nil {
		fmt.Println("pasre file:", err)
		os.Exit(1)
	}
	var tIfaces []types.Interface
	for _, iface := range file.Interfaces {
		if isInStringSlice(iface.Name, flagTargets.L) {
			tIfaces = append(tIfaces, iface)
		}
	}
	f := jen.NewFile(file.Name)
	for _, i := range tIfaces {
		f.Line().Add(tpl.MiddlewareTPL(&i))
		for _, mwName := range flagMiddlewares.L {
			var t tpl.TPL
			switch mwName {
			case "logs":
				t = tpl.LoggingMiddleware
			default:
				fmt.Println("Warning: middleware", mwName, "not found")
				continue
			}
			f.Line().Add(t(&i))
		}
	}
	var b bytes.Buffer
	err = f.Render(&b)
	if err != nil {
		fmt.Println("render error:", err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(*flagFileName, b.Bytes(), os.ModePerm)
	if err != nil {
		fmt.Println("write file error:", err)
		os.Exit(1)
	}
}

type StringListFlagValue struct {
	L []string
}

func (v *StringListFlagValue) Set(s string) error {
	v.L = strings.Split(s, ",")
	return nil
}

func (v *StringListFlagValue) String() string {
	return "[" + strings.Join(v.L, ", ") + "]"
}

func isInStringSlice(what string, where []string) bool {
	for _, item := range where {
		if item == what {
			return true
		}
	}
	return false
}
