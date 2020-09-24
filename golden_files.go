package testutils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/logrusorgru/aurora"
)

const (
	TestFixturesDir   = "test-fixtures"
	GoldenFileDirName = "snapshot"
	GoldenFileExt     = ".golden"
	GoldenFileDirPath = TestFixturesDir + string(filepath.Separator) + GoldenFileDirName
)

func GetGoldenFilePath(t *testing.T) string {
	t.Helper()
	// When using table-driven-tests, the `t.Name()` results in a string with slashes
	// which makes it impossible to reference in a filesystem, producing a "No such file or directory"
	filename := strings.ReplaceAll(t.Name(), "/", "_")
	return path.Join(GoldenFileDirPath, filename+GoldenFileExt)
}

func UpdateGoldenFileContents(t *testing.T, contents []byte) {
	t.Helper()

	goldenFilePath := GetGoldenFilePath(t)

	t.Log(aurora.Reverse(aurora.Red("!!! UPDATING GOLDEN FILE !!!")), goldenFilePath)

	err := ioutil.WriteFile(goldenFilePath, contents, 0600)
	if err != nil {
		t.Fatalf("could not update golden file (%s): %+v", goldenFilePath, err)
	}
}

func GetGoldenFileContents(t *testing.T) []byte {
	t.Helper()

	goldenPath := GetGoldenFilePath(t)
	if !fileOrDirExists(t, goldenPath) {
		t.Fatalf("golden file does not exist: %s", goldenPath)
	}
	f, err := os.Open(goldenPath)
	if err != nil {
		t.Fatalf("could not open file (%s): %+v", goldenPath, err)
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("could not read file (%s): %+v", goldenPath, err)
	}
	return bytes
}

func fileOrDirExists(t *testing.T, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
