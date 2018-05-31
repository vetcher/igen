package tpl

import (
	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"
)

func MiddlewareTPL(i *types.Interface) *Statement {
	return Line().Type().Id(mwTypeName(i)).Func().Call(Id(i.Name)).Id(i.Name)
}

func mwTypeName(i *types.Interface) string {
	return i.Name + "Middleware"
}
