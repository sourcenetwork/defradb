package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var (
	mainTpl = template.Must(template.New("main").Parse(mainTplString))

	FailedBuildErr = fmt.Errorf("build")

	tmpDir = os.TempDir()
)

var ignorePaths = []string{
	// non code folders
	"x",
	"tools",
	"tests",
	"licenses",
	"examples",
	"docs",
	"build",
	".",
	".git",
	".github",

	// specific packages
	"query/graphql/schema/examples",
	"datastore/badger",
	"cmd",
	"cli",
	"version",
	"net",
	"api",
}

type Package struct {
	Name      string
	Path      string
	Error     string
	Buildable bool
}

func NewPackage(pkgpath string) *Package {
	pkg1 := Package{
		Path: pkgpath,
	}
	pkg1.Name = path.Base(pkg1.Path)
	return &pkg1
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must specify path to repo root")
		os.Exit(1)
	}
	defraRootPath := os.Args[1]

	pkgs, err := doAllPackages(defraRootPath)
	if err != nil {
		fmt.Println("failed:", err)
	} else {
		fmt.Println("succesful!")
	}

	success := 0
	for _, p := range pkgs {
		if p.Buildable {
			success++
		}
	}
	fmt.Printf("Built %v/%v\n", success, len(pkgs))
}

func doAllPackages(rootpath string) ([]*Package, error) {
	pkgs := make([]*Package, 0)
	err := filepath.Walk(rootpath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		path = cleanPath(path)

		if contains(ignorePaths, path) {
			return filepath.SkipDir
		}

		pkg := NewPackage(path)
		if err := testPackage(pkg, tmpDir); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		printPackageSummary(*pkg)
		pkgs = append(pkgs, pkg)

		return nil
	})
	return pkgs, err
}

// func doPackages(rootpath string, ps ...string) ([]*Package, error) {
// 	pkgs := make([]*Package, 0)
// 	for _, p := range  ps {
// 		pkg := NewPackage(path)
// 		if err := testPackage(pkg, tmpDir); err != nil {
// 			fmt.Println("error:", err)
// 			os.Exit(1)
// 		}
// 		printPackageSummary(*pkg)
// 		pkgs = append(pkgs, pkg)
// 	}
// 	return nil, nil
// }

func printPackageSummary(pkg Package) {
	fmt.Println("== PACKAGE BUILD SUMMARY ==")
	fmt.Printf("Name: '%v'\nStatus: %v\nReason:	%v\n", pkg.Name, statusEmoji(pkg.Buildable), cleanPath(pkg.Error))
}

func testPackage(pkg *Package, tempDir string) error {
	outpath := path.Join(tempDir, fmt.Sprintf("main_pkg_%s.go", pkg.Name))
	if err := execMainTemplate(*pkg, outpath); err != nil {
		return err
	}

	resp, err := tinygoBuild(outpath)
	if err != nil && errors.Is(err, FailedBuildErr) {
		pkg.Buildable = false
		pkg.Error = resp
	} else if err != nil {
		return fmt.Errorf("running: %w", err)
	}
	// fmt.Println("Output Run Package:", resp)

	return nil
}

func execMainTemplate(pkg Package, outpath string) error {
	var buf bytes.Buffer
	err := mainTpl.Execute(&buf, pkg)
	if err != nil {
		return fmt.Errorf("compiling template: %w", err)
	}

	if err := os.WriteFile(outpath, buf.Bytes(), 0700); err != nil {
		return fmt.Errorf("writting compiled main file: %w", err)
	}

	return nil
}

func tinygoBuild(pathfile string) (string, error) {
	tmp := os.TempDir()
	outpath := path.Join(tmp, "tinygo_wasm_pkg_test")
	// fmt.Printf("> tinygo run %v\n", pathfile)
	cmd := exec.Command("tinygo", "build", "-o", outpath, pathfile)
	out, err := cmd.CombinedOutput()
	if err != nil && errors.Is(err, exec.ErrNotFound) {
		return "", err
	} else if err != nil {
		return cleanResp(string(out)), fmt.Errorf("failed to 'go run': %w (%s)", FailedBuildErr, out)
	}

	return "", nil
}

func statusEmoji(b bool) string {
	if b {
		return "✔️"
	}
	return "❌"
}

func cleanPath(s string) string {
	return strings.ReplaceAll(s, "../", "")
}

func cleanResp(s string) string {
	i := strings.Index(s, "no input files")
	if i == -1 {
		return s
	}
	return s[:i+len("no input files")]
}

func contains(s []string, e string) bool {
	if e == "" {
		return false
	}
	for _, a := range s {
		if strings.HasPrefix(a, e) {
			return true
		}
	}
	return false
}

var (
	mainTplString = `
package main

import (
	_ "github.com/sourcenetwork/defradb/{{.Path}}"
)

func main() {}
`
)
