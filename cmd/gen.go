package cmd

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

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
)

var httpFn = "FuncFw.Export.Http"

// var httpAsyncFn = "FuncFw.Export.HttpAsync"
var httpNatsAsyncFn = "FuncFw.Export.NatsAsync"

var pkgPath = "/pkg"
var templateFile = getDir() + pkgPath + "/" + "template/template.go"

const tmpDir = "deployments/tmp" + "/"
const ImportPath = "github.com/Ishan27g/ryo-Faas/" + tmpDir

type function struct {
	pkgName    string
	entrypoint string
	isAsync    bool
}

func GenerateFile(isMain bool, toDir string, fns []*deploy.Function) (bool, string) {
	var deployments []function
	for _, fn := range fns {
		if !validate(isMain, fn.GetFilePath(), fn.GetEntrypoint()) {
			fmt.Println("invalid")
			return false, ""
		}
		dir, _ := filepath.Split(fn.GetFilePath())
		pn := filepath.Base(dir)
		deployments = append(deployments, function{
			pn, fn.GetEntrypoint(), fn.Async,
		})
	}
	// generate a single file per deployment
	genFile, err := generate(isMain, toDir, deployments...)
	if err != nil {
		fmt.Println(err.Error())
		return false, ""
	}
	return true, genFile
}

// todo check entrypoint only
func validate(isMain bool, fileName, entrypoint string) bool {
	fmt.Println("Validating - ", fileName, entrypoint)

	set := token.NewFileSet()
	node, err := parser.ParseFile(set, fileName, nil, parser.ParseComments)
	if err != nil {
		log.Fatal("Cannot validate ", err.Error())
		return false
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
			if isMain {
				if ret.Name.String() == "Init" {
					params := ret.Type.Params.List
					if len(params) == 0 {
						valid = true
					}
				}
			} else {
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
		}
		return true
	})
	return valid
}

func generate(isMain bool, toDir string, fns ...function) (string, error) {
	var genFile string

	set := token.NewFileSet()
	node, err := parser.ParseFile(set, templateFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatal("Unable to open generate.go template ", err.Error())
		return genFile, err
	}
	dir := ImportPath

	for _, fn := range fns {
		var fnFwCall = httpFn
		if fn.isAsync {
			fnFwCall = httpNatsAsyncFn
		}
		packageAlias := strings.ReplaceAll(fn.pkgName, "-", "")
		for i := 0; i < len(node.Decls); i++ {
			switch isMain {
			case true:
				d := node.Decls[i]
				switch d.(type) {
				case *ast.GenDecl:
					dd := d.(*ast.GenDecl)
					if dd.Tok == token.IMPORT {
						// add the new import
						iSpec := &ast.ImportSpec{
							Name: &ast.Ident{Name: packageAlias},
							Path: &ast.BasicLit{Value: strconv.Quote(dir + fn.pkgName)},
						}
						dd.Specs = append(dd.Specs, iSpec)
					}
				case *ast.FuncDecl:
					if d.(*ast.FuncDecl).Name.String() == "init" {
						newCallStmt := &ast.ExprStmt{ // functions.HTTP(
							X: &ast.CallExpr{
								Fun: &ast.Ident{
									Name: packageAlias + "." + "Init",
								},
							},
						}
						d.(*ast.FuncDecl).Body.List = append([]ast.Stmt{newCallStmt},
							d.(*ast.FuncDecl).Body.List...)
					}
				}
			case false:
				d := node.Decls[i]
				switch d.(type) {
				case *ast.GenDecl:
					dd := d.(*ast.GenDecl)
					if dd.Tok == token.IMPORT {
						// add the new import
						iSpec := &ast.ImportSpec{
							Name: &ast.Ident{Name: packageAlias},
							Path: &ast.BasicLit{Value: strconv.Quote(dir + fn.pkgName)},
						}
						dd.Specs = append(dd.Specs, iSpec)
					}
				case *ast.FuncDecl:
					if d.(*ast.FuncDecl).Name.String() == "init" {
						newCallStmt := &ast.ExprStmt{ // functions.HTTP(
							X: &ast.CallExpr{
								Fun: &ast.Ident{
									Name: fnFwCall,
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"" + fn.entrypoint + "\"",
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: "\"/" + strings.ToLower(fn.entrypoint) + "\"",
									},
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: packageAlias + `.` + fn.entrypoint,
									},
								},
							},
						}
						d.(*ast.FuncDecl).Body.List = append([]ast.Stmt{newCallStmt},
							d.(*ast.FuncDecl).Body.List...)
					}
				}
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
	genFile = toDir + "exported.go"

	_, err = os.Create(genFile)
	if err != nil {
		log.Println("cant create", err.Error())
		return "", err
	}
	err = ioutil.WriteFile(genFile, out, os.ModePerm)
	if err != nil {
		log.Println(err)
		return genFile, err
	}
	return genFile, nil
}