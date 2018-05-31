package tpl

import (
	"strings"
	"unicode"

	"strconv"

	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"
)

type TPL func(i *types.Interface) *Statement

const (
	_next_   = "next"
	_logger_ = "logger"
)

const (
	PackageTime     = "time"
	PackageContext  = "context"
	PackageGoKitLog = "github.com/go-kit/kit/log"
)

func rec(name string) string {
	return lastWordFromName(name)
}

func methodDefinitionFull(obj string, signature *types.Function, fillEmpty bool) *Statement {
	return Func().
		Params(Id(rec(obj)).Id(obj)).
		Add(functionDefinition(signature, fillEmpty))
}

func functionDefinition(signature *types.Function, fillEmpty bool) *Statement {
	return Id(signature.Name).
		Params(funcDefinitionParams(signature.Args, fillEmpty, "arg")).
		Params(funcDefinitionParams(signature.Results, fillEmpty, "res"))
}

func interfaceType(p *types.Interface) (code []Code) {
	for _, x := range p.Methods {
		code = append(code, functionDefinition(x, false))
	}
	return
}

func funcDefinitionParams(fields []types.Variable, fillEmpty bool, pref string) *Statement {
	c := &Statement{}
	c.ListFunc(func(g *Group) {
		if fillEmpty {
			fields = addNamesToVars(fields, pref)
		}
		for _, field := range fields {
			g.Id(toLowerFirst(field.Name)).Add(fieldType(field.Type, true))
		}
	})
	return c
}

func fieldType(field types.Type, allowEllipsis bool) *Statement {
	c := &Statement{}
	for field != nil {
		switch f := field.(type) {
		case types.TImport:
			if f.Import != nil {
				c.Qual(f.Import.Package, "")
			}
			field = f.Next
		case types.TName:
			c.Id(f.TypeName)
			field = nil
		case types.TArray:
			if f.IsSlice {
				c.Index()
			} else if f.ArrayLen > 0 {
				c.Index(Lit(f.ArrayLen))
			}
			field = f.Next
		case types.TMap:
			return c.Map(fieldType(f.Key, false)).Add(fieldType(f.Value, false))
		case types.TPointer:
			c.Op(strings.Repeat("*", f.NumberOfPointers))
			field = f.Next
		case types.TInterface:
			mhds := interfaceType(f.Interface)
			return c.Interface(mhds...)
		case types.TEllipsis:
			if allowEllipsis {
				c.Op("...")
			} else {
				c.Index()
			}
			field = f.Next
		default:
			return c
		}
	}
	return c
}

func toLowerFirst(s string) string {
	x := []rune(s)
	if len(x) == 0 {
		return ""
	}
	x[0] = unicode.ToLower(x[0])
	return string(x)
}

func toUpperFirst(s string) string {
	x := []rune(s)
	if len(x) == 0 {
		return ""
	}
	x[0] = unicode.ToUpper(x[0])
	return string(x)
}

func lastWordFromName(name string) string {
	lastUpper := strings.LastIndexFunc(name, unicode.IsUpper)
	if lastUpper == -1 {
		lastUpper = 0
	}
	return strings.ToLower(name[lastUpper:])
}

func isInStringSlice(what string, where []string) bool {
	for _, item := range where {
		if item == what {
			return true
		}
	}
	return false
}

// Remove from function fields error if it is last in slice
func removeErrorIfLast(fields []types.Variable) []types.Variable {
	if isErrorLast(fields) {
		return fields[:len(fields)-1]
	}
	return fields
}

func isErrorLast(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[len(fields)-1].Type)
	return name != nil &&
		types.TypeImport(fields[len(fields)-1].Type) == nil &&
		*name == "error"
}

// Return name of error, if error is last result, else return `err`
func nameOfLastResultError(fn *types.Function) string {
	if isErrorLast(fn.Results) {
		return fn.Results[len(fn.Results)-1].Name
	}
	return "err"
}

// Remove from function fields context if it is first in slice
func removeContextIfFirst(fields []types.Variable) []types.Variable {
	if isContextFirst(fields) {
		return fields[1:]
	}
	return fields
}

func isContextFirst(fields []types.Variable) bool {
	if len(fields) == 0 {
		return false
	}
	name := types.TypeName(fields[0].Type)
	return name != nil &&
		types.TypeImport(fields[0].Type) != nil &&
		types.TypeImport(fields[0].Type).Package == PackageContext &&
		*name == "Context"
}

func paramNames(fields []types.Variable) *Statement {
	var list []Code
	for _, field := range fields {
		v := Id(toLowerFirst(field.Name))
		if types.IsEllipsis(field.Type) {
			v.Op("...")
		}
		list = append(list, v)
	}
	return List(list...)
}

func addNamesToVars(fields []types.Variable, pref string) (res []types.Variable) {
	res = make([]types.Variable, len(fields))
	copy(res, fields)
	for i := range res {
		if res[i].Name == "" {
			res[i].Name = pref + strconv.Itoa(i)
		}
	}
	return
}
