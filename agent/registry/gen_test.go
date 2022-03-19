package registry

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAstGen(t *testing.T) {
	var exports = []function{
		{
			pkgName:    "method1",
			entrypoint: "Method1",
		},
		{
			pkgName:    "method2",
			entrypoint: "Method2",
		},
	}

	// change relative path
	pathToDeployment = path() + FnFw
	PathToFns = pathToDeployment + "functions/"
	modFile = func() string {
		return path() + "/template/template.go"
	}
	os.MkdirAll(PathToFns, os.ModePerm)

	fmt.Println("pathToDeployment", pathToDeployment)
	fmt.Println("PathToFns", PathToFns)
	fmt.Println("modFile", modFile())

	dotGo, err := rewriteDeployDotGo(exports...)
	assert.NoError(t, err)
	assert.FileExists(t, dotGo)
	os.Remove(dotGo)
}

//package registry

// import (
// 	"bytes"
// 	"fmt"
// 	"go/ast"
// 	"go/format"
// 	"go/parser"
// 	"go/printer"
// 	"go/token"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"strconv"
// 	"strings"

// 	deploy "github.com/Ishan27g/ryo-Faas/proto"
// )

// func astLocalCopy(fns []*deploy.Function) (bool, string) {
// 	var deployments []function
// 	for _, fn := range fns {
// 		var args []string
// 		var valid = false
// 		if valid, args = validate(fn.GetFilePath(), fn.GetEntrypoint(), functionType); !valid {
// 			fmt.Println("invalid")
// 			return false, ""
// 		}
// 		dir, _ := filepath.Split(fn.GetFilePath())
// 		pn := filepath.Base(dir)
// 		deployments = append(deployments, function{
// 			pn, fn.GetEntrypoint(), args,
// 		})
// 	}

// 	genFile, err := rewriteDeployDotGo(functionType, deployments...)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return false, ""
// 	}
// 	fmt.Println("Generated file ", genFile)
// 	return true, genFile
// }

// func validate(fileName, entrypoint string, fnType string) (bool, []string) {
// 	var args []string

// 	if (strings.Compare(fnType, "http") != 0) || (strings.Compare(fnType, "event") != 0) {
// 		return false, args
// 	}
// 	set := token.NewFileSet()
// 	node, err := parser.ParseFile(set, fileName, nil, parser.ParseComments)
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	formatNode := func(node ast.Node) string {
// 		buf := new(bytes.Buffer)
// 		_ = format.Node(buf, token.NewFileSet(), node)
// 		return buf.String()
// 	}
// 	valid := false
// 	ast.Inspect(node, func(n ast.Node) bool {
// 		switch ret := n.(type) {
// 		case *ast.FuncDecl:
// 			if ret.Name.Name == entrypoint {
// 				params := ret.Type.Params.List
// 				if len(params) == 2 {
// 					switch fnType {
// 					case "http":
// 						firstParameterIsW := formatNode(params[0].Names[0]) == "w" &&
// 							formatNode(params[0].Type) == "http.ResponseWriter"
// 						secondParameterIsR := formatNode(params[1].Names[0]) == "r" &&
// 							formatNode(params[1].Type) == "*http.Request"
// 						if firstParameterIsW && secondParameterIsR {
// 							args = append(args, formatNode(params[1].Names[0]))
// 							args = append(args, formatNode(params[1].Names[0]))
// 							valid = true
// 						}
// 					case "event":
// 						firstParameterIsE := formatNode(params[1].Type) == "store.Event" // && formatNode(params[1].Names[0]) == "eventCb" &&
// 						if firstParameterIsE {
// 							valid = true
// 						}
// 						args = append(args, formatNode(params[0].Names[0]))
// 					}
// 				}
// 			}
// 		}
// 		return true
// 	})
// 	return valid, args
// }

// type function struct {
// 	pkgName    string
// 	entrypoint string
// 	args       []string
// }

// func rewriteDeployDotGo(functionType string, fns ...function) (string, error) {
// 	var genFile string

// 	set := token.NewFileSet()
// 	node, err := parser.ParseFile(set, modFile(), nil, parser.ParseComments)
// 	if err != nil {
// 		log.Fatal(err)
// 		return genFile, err
// 	}
// 	dir := importPath

// 	for _, fn := range fns {
// 		packageAlias := strings.ReplaceAll(fn.pkgName, "-", "")
// 		for i := 0; i < len(node.Decls); i++ {
// 			d := node.Decls[i]
// 			switch d.(type) {
// 			case *ast.GenDecl:
// 				dd := d.(*ast.GenDecl)
// 				if dd.Tok == token.IMPORT {
// 					// add the new import
// 					iSpec := &ast.ImportSpec{
// 						Name: &ast.Ident{Name: packageAlias},
// 						Path: &ast.BasicLit{Value: strconv.Quote(dir + fn.pkgName)},
// 					}
// 					dd.Specs = append(dd.Specs, iSpec)
// 					if functionType == "event" {
// 						iSpec = &ast.ImportSpec{Path: &ast.BasicLit{Value: strconv.Quote("github.com/Ishan27g/ryo-Faas/store")}}
// 						dd.Specs = append(dd.Specs, iSpec)
// 					}
// 				}
// 			case *ast.FuncDecl:
// 				if d.(*ast.FuncDecl).Name.String() == "init" {
// 					//stmt := &ast.AssignStmt{
// 					//	Lhs: []ast.Expr{
// 					//		&ast.Ident{Name: "handlerFunc"},
// 					//	},
// 					//	Tok: token.ASSIGN,
// 					//	Rhs: []ast.Expr{
// 					//		&ast.Ident{Name: packageAlias + `.` + entrypoint},
// 					//	},
// 					//}
// 					//stmt2 := &ast.AssignStmt{
// 					//	Lhs: []ast.Expr{
// 					//		&ast.Ident{Name: "entrypoint"},
// 					//	},
// 					//	Tok: token.ASSIGN,
// 					//	Rhs: []ast.Expr{
// 					//		&ast.Ident{Name: "\"" + entrypoint + "\""},
// 					//	},
// 					//}
// 					var newCallStmt *ast.ExprStmt
// 					switch functionType {
// 					case "http":
// 						newCallStmt = &ast.ExprStmt{ // functions.HTTP(
// 							X: &ast.CallExpr{
// 								Fun: &ast.Ident{
// 									Name: "FuncFw.Export.Http",
// 								},
// 								Args: []ast.Expr{
// 									&ast.BasicLit{
// 										Kind:  token.STRING,
// 										Value: "\"" + fn.entrypoint + "\"",
// 									},
// 									&ast.BasicLit{
// 										Kind:  token.STRING,
// 										Value: "\"/" + strings.ToLower(fn.entrypoint) + "\"",
// 									},
// 									&ast.BasicLit{
// 										Kind:  token.STRING,
// 										Value: packageAlias + `.` + fn.entrypoint,
// 									},
// 								},
// 							},
// 						}
// 						d.(*ast.FuncDecl).Body.List = append([]ast.Stmt{newCallStmt},
// 							d.(*ast.FuncDecl).Body.List...)
// 					case "event":
// 						newCallStmt = &ast.ExprStmt{
// 							X: &ast.CallExpr{
// 								Fun: &ast.Ident{
// 									Name: "store.Document." + fn.entrypoint,
// 								},
// 								Args: []ast.Expr{
// 									&ast.BasicLit{
// 										Kind:  token.STRING,
// 										Value: fn.args[0],
// 									},
// 								},
// 							},
// 						}
// 						d.(*ast.FuncDecl).Body.List = append([]ast.Stmt{newCallStmt},
// 							d.(*ast.FuncDecl).Body.List...)
// 					}

// 					// add the new function call with relevant
// 					//newCallStmt := &ast.ExprStmt{ // functions.HTTP(
// 					//	X: &ast.CallExpr{
// 					//		Fun: &ast.Ident{
// 					//			Name: "deploy",
// 					//		},
// 					//		Args: []ast.Expr{
// 					//			&ast.BasicLit{
// 					//				Kind:  token.STRING,
// 					//				Value: "\"" + strings.ToLower(entrypoint) + "\"",
// 					//			},
// 					//			&ast.BasicLit{
// 					//				Kind:  token.STRING,
// 					//				Value: pkgName + `.` + entrypoint,
// 					//			},
// 					//		},
// 					//	},
// 					//}

// 				}
// 			}
// 		}
// 	}

// 	// Sort the imports
// 	ast.SortImports(set, node)

// 	// Generate the code
// 	var output []byte
// 	buffer := bytes.NewBuffer(output)
// 	fmt.Println("WTF", string(output))
// 	if err := printer.Fprint(buffer, set, node); err != nil {
// 		log.Println(err)
// 		return genFile, err
// 	}
// 	out := buffer.Bytes()
// 	genFile = getGenFilePath("exported-function")
// 	os.Create(genFile)
// 	err = ioutil.WriteFile(genFile, out, os.ModePerm)
// 	if err != nil {
// 		log.Println(err)
// 		return genFile, err
// 	}
// 	return genFile, nil
// }
