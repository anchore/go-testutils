package testutils

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/anchore/stereoscope/pkg/image"
	"github.com/logrusorgru/aurora"
)

const (
	goldenFileDirName = "snapshot"
	goldenFileExt     = ".golden"
)

var goldenFileDirPath = path.Join(testFixturesDir, goldenFileDirName)

func GetGoldenFixtureImage(t *testing.T, name string) *image.Image {
	t.Helper()

	imageName, _ := getFixtureImageInfo(t, name)
	tarFileName := imageName + goldenFileExt
	tarPath := getFixtureImageTarPath(t, name, goldenFileDirPath, tarFileName)
	return getFixtureImageFromTar(t, tarPath)
}

func UpdateGoldenFixtureImage(t *testing.T, name string) {
	t.Helper()

	t.Log(aurora.Reverse(aurora.Red("!!! UPDATING GOLDEN FIXTURE IMAGE !!!")), name)

	imageName, _ := getFixtureImageInfo(t, name)
	goldenTarFilePath := path.Join(goldenFileDirPath, imageName+goldenFileExt)
	tarPath := GetFixtureImageTarPath(t, name)
	copyFile(t, tarPath, goldenTarFilePath)
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()

	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("could not open src (%s): %+v", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("could not open dst (%s): %+v", dst, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		t.Fatalf("could not copy file (%s -> %s): %+v", src, dst, err)
	}
}

func GetGoldenFilePath(t *testing.T) string {
	t.Helper()
	// When using table-driven-tests, the `t.Name()` results in a string with slashes
	// which makes it impossible to reference in a filesystem, producing a "No such file or directory"
	filename := strings.ReplaceAll(t.Name(), "/", "_")
	return path.Join(goldenFileDirPath, filename+goldenFileExt)
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
