package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var protoFileToPkgName map[string]string

func main() {
	if len(os.Args) == 1 {
		log.Print("Need to provide source directory as an argument")
		return
	}

	srcDir := os.Args[1]
	tgtDir := "protobuf/"

	os.MkdirAll(tgtDir, os.ModePerm)

	protoFileToPkgName = make(map[string]string)
	err := buildProtobufs(srcDir, tgtDir)
	if err != nil {
		log.Printf("Failed to compile .proto: %s", err)
	}
}

func buildProtobufs(srcDir, tgtDir string) error {
	if len(srcDir) == 0 {
		return fmt.Errorf("No source directory found")
	}
	if len(tgtDir) == 0 {
		return fmt.Errorf("No target directory found")
	}

	err := dirExists(srcDir)
	if err != nil {
		return fmt.Errorf("%s; please create said directory", err.Error())
	}

	fileList, err := getSameTypeOfFileList(srcDir, ".proto")
	if err != nil {
		return err
	}
	if len(fileList) == 0 {
		return fmt.Errorf("Looks like there aren't any .proto files")
	}

	err = buildProtobufPkgDir(tgtDir, fileList)
	if err != nil {
		return err
	}

	err = fixImportStmts(fileList)
	if err != nil {
		return err
	}

	err = compileProtos(tgtDir, fileList)
	if err != nil {
		return err
	}

	return nil
}

func copyProtoFilesToPkgDirs(tgtDir string, fileList []string) error {
	for _, protoFilePath := range fileList {
		protoFileName := filepath.Base(protoFilePath)
		pkgName := protoFileToPkgName[protoFileName]
		cmd := exec.Command("cp", protoFilePath, fmt.Sprintf("%s/%s", tgtDir, pkgName))
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Could not copy protofile %s to package directory: %v", protoFileName, err)
		}
	}
	return nil
}

func compileProtos(tgtDir string, fileList []string) error {
	for _, protoFilePath := range fileList {
		protoFileName := filepath.Base(protoFilePath)
		pkgName := protoFileToPkgName[protoFileName]
		cmd := exec.Command("protoc", fmt.Sprintf("--go_out=%s/%s", tgtDir, pkgName), protoFilePath)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Could not compile protofile %s: %v", protoFileName, err)
		}
	}
	return nil
}

func buildProtobufPkgDir(tgtDir string, fileList []string) error {
	for _, protoFilePath := range fileList {
		contents, err := getFileContents(protoFilePath)
		if err != nil {
			return err
		}

		pkgName := getPkgName(contents)
		if len(pkgName) == 0 {
			continue
		}
		os.MkdirAll(fmt.Sprintf("%s/%s", tgtDir, pkgName), os.ModePerm)

		protoFileName := filepath.Base(protoFilePath)
		protoFileToPkgName[protoFileName] = pkgName
	}
	return nil
}

func getPkgName(contents string) string {
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()
		if isPkgStmt(line) {
			line = strings.Replace(line, "option go_package = \"", "", -1)
			line = strings.Replace(line, "\";", "", -1)
			return line
		}
	}
	return ""
}

func isPkgStmt(stmt string) bool {
	pattern := regexp.MustCompile(`^option go_package = "(.*)";`)
	matches := pattern.FindAllString(stmt, -1)
	if matches == nil {
		return false
	}
	if matches[0] == stmt {
		return true
	}
	return false
}

func fixImportStmts(fileList []string) error {
	for _, protoFilePath := range fileList {
		contents, err := getFileContents(protoFilePath)
		if err != nil {
			return err
		}

		oldImportStmt := getImportStmt(contents)
		if len(oldImportStmt) == 0 {
			continue
		}

		newImportStmt := fixImportStmt(oldImportStmt)
		contents = strings.Replace(contents, oldImportStmt, newImportStmt, -1)
		err = ioutil.WriteFile(protoFilePath, []byte(contents), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func fixImportStmt(stmt string) string {
	fmt.Println(stmt)
	tmp := strings.Replace(stmt, "import \"", "", -1)
	protoName := strings.Replace(tmp, "\";", "", -1)
	pkg := protoFileToPkgName[protoName]
	return fmt.Sprintf("import \"%s/%s\";", pkg, protoName)
}

func getImportStmt(contents string) string {
	scanner := bufio.NewScanner(strings.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()
		if isImportStmt(line) {
			return line
		}
	}
	return ""
}

func isImportStmt(stmt string) bool {
	pattern := regexp.MustCompile(`^import "(.*)\.proto\";`)
	matches := pattern.FindAllString(stmt, -1)
	if matches == nil {
		return false
	}
	if matches[0] == stmt {
		return true
	}
	return false
}

func getFileContents(filename string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Could not parse file: %v", err)
	}
	return string(fileBytes), nil
}

func getSameTypeOfFileList(srcDir, fileType string) ([]string, error) {
	fileList := []string{}
	fmt.Println(srcDir)
	err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), fileType) {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Could not parse %s files", fileType)
	}
	return fileList, nil
}

func dirExists(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("No directory named '%s' found", dir)
		} else {
			return err
		}
	}
	return nil
}
