package gombie

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"

	"github.com/fatih/astrewrite"
	testParser "github.com/jstemmer/go-junit-report/parser"
)

func writeFilesToTmpDir(tmpDir string, files map[string]*ast.File, fset *token.FileSet) error {
	for fpath, file := range files {
		tfname := filepath.Join(tmpDir, filepath.Base(fpath))
		// fmt.Println("creating", tfname)
		f, err := os.Create(tfname)
		if err != nil {
			return err
		}

		printer.Fprint(f, fset, file)
	}

	return nil
}

func rewriteImportsInFile(fset *token.FileSet, files map[string]*ast.File, oldImport, newImport string) {
	for _, file := range files {
		astutil.RewriteImport(fset, file, oldImport, newImport)
	}
}

func runTests(testMatch string, fromPath string) bool {
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "go"
	cmdArgs := []string{"test", "-v", "-run", testMatch}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = fromPath

	cmdOut, err = cmd.Output()
	fmt.Println(string(cmdOut))
	// TODO: read stream directly
	_, err = testParser.Parse(bytes.NewBuffer(cmdOut), "mockpackage_test")
	if err != nil {
		fmt.Println(err)
	}

	return true
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// RunTestsOnceWithMutator does as it says on the tin
// It returns whether or not the tests passed, and any errors
func RunTestsOnceWithMutator(fset *token.FileSet, targetPackage *ast.Package, testPackage *ast.Package, testMatch string, m Mutator) (bool, error) {

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return false, errors.New("GOPATH is empty")
	}

	if exists, err := pathExists(gopath); !exists || err != nil {
		return false, errors.New("GOPATH is not valid")
	}

	// create temporary directory
	tmpDir, err := ioutil.TempDir(gopath, "src/gombie")
	if err != nil {
		fmt.Println("Failed to create temporary directory:", err)
		return false, err
	}

	fmt.Println(tmpDir)
	// defer os.RemoveAll(tmpDir)

	// rewriteImportMutator := RewriteMainPackageImports{
	// 	match:   targetPackage.Name,
	// 	rewrite: tmpDir,
	// }

	// change the import(s) in test package to the new path

	rewriteFunc := func(n ast.Node) (ast.Node, bool) {
		// updatedNode, updated := rewriteImportMutator.Mutate(n)
		updatedNode, updated := m.Mutate(n)
		return updatedNode, !updated
	}

	_ = astrewrite.Walk(targetPackage, rewriteFunc)
	// spew.Dump(targetPackage)

	err = writeFilesToTmpDir(tmpDir, targetPackage.Files, fset)
	if err != nil {
		return false, err
	}

	rewriteImportsInFile(fset, testPackage.Files, "github.com/oskanberg/gombie/mockpackage", filepath.Base(tmpDir))
	err = writeFilesToTmpDir(tmpDir, testPackage.Files, fset)
	if err != nil {
		return false, err
	}

	return runTests(testMatch, tmpDir), nil
}

// Go Runs
func Go(testFlag string, m Mutator) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		fmt.Println("Failed to parse directory:", err)
		os.Exit(1)
	}

	success, err := RunTestsOnceWithMutator(fset, pkgs["mockpackage"], pkgs["mockpackage_test"], testFlag, m)
	if err != nil {
		fmt.Println("Failed to execute:")
		os.Exit(1)
	}

	fmt.Println("Tests succeeded?", success)
}
