package tpl

import (
	. "github.com/dave/jennifer/jen"
	"github.com/vetcher/go-astra/types"
)

func LoggingMiddleware(i *types.Interface) *Statement {
	s := &Statement{}
	s.Line().Func().Id(toUpperFirst(logsStructName(i))).Params(Id(_logger_).Qual(PackageGoKitLog, "Logger")).Params(Id(mwTypeName(i))).
		Block(newLoggingBody(i))

	s.Line()

	// Render type logger
	s.Type().Id(toLowerFirst(logsStructName(i))).Struct(
		Id(_logger_).Qual(PackageGoKitLog, "Logger"),
		Id(_next_).Id(i.Name),
	)

	// Render functions
	for _, signature := range i.Methods {
		s.Line()
		s.Add(loggingFunc(i, signature)).Line()
	}
	s.Type().Op("(")
	for _, signature := range i.Methods {
		if params := removeContextIfFirst(signature.Args); len(params) > 0 {
			s.Line().Add(loggingEntity("log"+inputStructName(signature), addNamesToVars(params, "arg")))
		}
		if params := removeErrorIfLast(signature.Results); len(params) > 0 {
			s.Line().Add(loggingEntity("log"+outputStructName(signature), addNamesToVars(params, "res")))
		}
	}
	s.Op(")")

	return s
}

func loggingFunc(i *types.Interface, signature *types.Function) *Statement {
	return methodDefinitionFull(toLowerFirst(logsStructName(i)), signature, true).
		BlockFunc(loggingFuncBody(i, signature))
}

func newLoggingBody(i *types.Interface) *Statement {
	return Return(Func().Params(
		Id(_next_).Id(i.Name),
	).Params(
		Id(i.Name),
	).BlockFunc(func(g *Group) {
		g.Return(Op("&").Id(toLowerFirst(logsStructName(i))).Values(
			Dict{
				Id(_logger_): Id(_logger_),
				Id(_next_):   Id(_next_),
			},
		))
	}))
}

func loggingEntity(name string, params []types.Variable) Code {
	if len(params) == 0 {
		return Empty()
	}
	return Id(name).StructFunc(func(g *Group) {
		for _, field := range params {
			g.Id(toUpperFirst(field.Name)).Add(fieldType(field.Type, false))
		}
	})
}

func loggingFuncBody(i *types.Interface, signature *types.Function) func(g *Group) {
	return func(g *Group) {
		g.Defer().Func().Params(Id("begin").Qual(PackageTime, "Time")).Block(
			Id(rec(logsStructName(i))).Dot(_logger_).Dot("Log").CallFunc(func(g *Group) {
				g.Line().Lit("method")
				g.Lit(signature.Name)
				g.Line().List(Lit("input"), logRequest(signature))
				g.Line().List(Lit("output"), logResponse(signature))
				if isErrorLast(signature.Results) {
					g.Line().List(Lit(nameOfLastResultError(signature)), Id(nameOfLastResultError(signature)))
				}
				g.Line().List(Lit("took"), Qual(PackageTime, "Since").Call(Id("begin")))
			}),
		).Call(Qual(PackageTime, "Now").Call())
		g.Return().Id(rec(logsStructName(i))).Dot(_next_).Dot(signature.Name).Call(paramNames(signature.Args))
	}
}

func logsStructName(i *types.Interface) string {
	return "Logs" + i.Name + "mw"
}

func logRequest(fn *types.Function) *Statement {
	return Id("log" + inputStructName(fn)).Add(fillMap(addNamesToVars(removeContextIfFirst(fn.Args), "arg")))
}

func logResponse(fn *types.Function) *Statement {
	return Id("log" + outputStructName(fn)).Add(fillMap(addNamesToVars(removeErrorIfLast(fn.Results), "res")))
}

func fillMap(params []types.Variable) *Statement {
	return Values(DictFunc(func(d Dict) {
		for i, field := range params {
			d[Id(toUpperFirst(field.Name))] = Id(params[i].Name)
		}
	}))
}

func inputStructName(signature *types.Function) string {
	return signature.Name + "Input"
}

func outputStructName(signature *types.Function) string {
	return signature.Name + "Output"
}
