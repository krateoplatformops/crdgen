package coder

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/krateoplatformops/crdgen/internal/strutil"
)

func GenerateManaged(workdir string, res *Resource) error {
	path, err := makeDirs(workdir, "apis", strings.ToLower(res.Kind), normalizeVersion(res.Version))
	if err != nil {
		return err
	}

	g := jen.NewFile(normalizeVersion(res.Version))
	g.ImportAlias(pkgCommon, pkgCommonAlias)

	g.Add(generateConditionFuncs(res))
	g.Line()

	src, err := os.Create(filepath.Join(path, "managed.go"))
	if err != nil {
		return err
	}
	defer src.Close()

	return g.Render(src)
}

func generateConditionFuncs(res *Resource) jen.Code {
	kind := strutil.ToGolangName(res.Kind)

	getter := jen.Func().Params(jen.Id("mg").Op("*").Id(kind)).
		Id("GetCondition").Params(
		jen.Id("ct").Qual(pkgCommon, "ConditionType"),
	).Qual(pkgCommon, "Condition").Block(
		jen.Return(jen.Id("mg").Dot("Status").Dot("GetCondition").Call(jen.Id("ct"))),
	)

	setter := jen.Func().Params(jen.Id("mg").Op("*").Id(kind)).
		Id("SetConditions").Params(
		jen.Id("c").Op("...").Qual(pkgCommon, "Condition"),
	).Block(
		jen.Id("mg").Dot("Status").Dot("SetConditions").Call(jen.Id("c").Op("...")),
	)

	return getter.Line().Line().Add(setter)
}
