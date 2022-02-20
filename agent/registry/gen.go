package registry

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
)

func astLocalCopy(fn *deploy.Function) (bool, string) {
	if !validate(fn.GetFilePath(), fn.GetEntrypoint()) {
		fmt.Println("invalid")
		return false, ""
	}

	dir, _ := filepath.Split(fn.GetFilePath())
	pn := filepath.Base(dir)

	genFile, err := rewriteDeployDotGo(pn, fn.GetEntrypoint())
	if err != nil {
		fmt.Println(err.Error())
		return false, ""
	}
	fmt.Println("Generated file ", genFile, "for", fn.GetEntrypoint())
	return true, genFile
}

// todo check entrypoint only
func validate(fileName, entrypoint string) bool {

	set := token.NewFileSet()
	node, err := parser.ParseFile(set, fileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err.Error())
	}
	formatNode := func(node ast.Node) string {
		buf := new(bytes.Buffer)
		_ = format.Node(buf, token.NewFileSet(), node)
		return buf.String()
	}
	valid := false
	ast.Inspect(node, func(n ast.Node) bool {
		switch ret := n.(type) {
		case *ast.FuncDecl:
			params := ret.Type.Params.List
			if len(params) == 2 {
				firstParameterIsW := formatNode(params[0].Names[0]) == "w" &&
					formatNode(params[0].Type) == "http.ResponseWriter"
				secondParameterIsR := formatNode(params[1].Names[0]) == "r" &&
					formatNode(params[1].Type) == "*http.Request"
				if firstParameterIsW && secondParameterIsR {
					valid = true
				}
			}
		}
		return true
	})
	return valid
}

func rewriteDeployDotGo(pkgName, entrypoint string) (string, error) {
	var genFile string

	set := token.NewFileSet()
	node, err := parser.ParseFile(set, modFile(), nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
		return genFile, err
	}
	dir := importPath

	packageAlias := strings.ReplaceAll(pkgName, "-", "")
	for i := 0; i < len(node.Decls); i++ {
		d := node.Decls[i]
		switch d.(type) {
		case *ast.GenDecl:
			dd := d.(*ast.GenDecl)
			if dd.Tok == token.IMPORT {
				// add the new import
				iSpec := &ast.ImportSpec{
					Name: &ast.Ident{Name: packageAlias},
					Path: &ast.BasicLit{Value: strconv.Quote(dir + pkgName)},
				}
				dd.Specs = append(dd.Specs, iSpec)
				//iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value:
				// strconv.Quote("github.com/GoogleCloudPlatform/functions-framework-go/functions")}}
				//dd.Specs = append(dd.Specs, iSpec)
			}
		case *ast.FuncDecl:
			if d.(*ast.FuncDecl).Name.String() == "init" {
				stmt := &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "handlerFunc"},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.Ident{Name: packageAlias + `.` + entrypoint},
					},
				}
				stmt2 := &ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.Ident{Name: "entrypoint"},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.Ident{Name: "\"" + entrypoint + "\""},
					},
				}
				// add the new function call with relevant
				//newCallStmt := &ast.ExprStmt{ // functions.HTTP(
				//	X: &ast.CallExpr{
				//		Fun: &ast.Ident{
				//			Name: "deploy",
				//		},
				//		Args: []ast.Expr{
				//			&ast.BasicLit{
				//				Kind:  token.STRING,
				//				Value: "\"" + strings.ToLower(entrypoint) + "\"",
				//			},
				//			&ast.BasicLit{
				//				Kind:  token.STRING,
				//				Value: pkgName + `.` + entrypoint,
				//			},
				//		},
				//	},
				//}
				d.(*ast.FuncDecl).Body.List = append([]ast.Stmt{stmt, stmt2},
					d.(*ast.FuncDecl).Body.List...)
			}
		}
	}

	// Sort the imports
	ast.SortImports(set, node)

	// Generate the code
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, set, node); err != nil {
		log.Println(err)
		return genFile, err
	}
	out := buffer.Bytes()
	genFile = getGenFilePath(entrypoint)
	err = ioutil.WriteFile(genFile, out, os.ModePerm)
	if err != nil {
		log.Println(err)
		return genFile, err
	}
	return genFile, nil
}
