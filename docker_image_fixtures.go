package testutils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/anchore/stereoscope"
	"github.com/anchore/stereoscope/pkg/image"
)

var (
	testFixturesDir = "test-fixtures"
	tarCacheDir     = path.Join(testFixturesDir, "tar-cache")
	imagePrefix     = "anchore-fixture"
)

func GetFixtureImage(t *testing.T, source, name string) (*image.Image, func()) {
	t.Helper()

	sourceObj := image.ParseSource(source)

	var location string
	switch sourceObj {
	case image.DockerTarballSource:
		location = GetFixtureImageTarPath(t, name)
	case image.DockerDaemonSource:
		location = LoadFixtureImageIntoDocker(t, name)
	default:
		t.Fatalf("could not determine source: %+v", source)
	}
	request := fmt.Sprintf("%s://%s", source, location)

	i, err := stereoscope.GetImage(request)
	if err != nil {
		t.Fatal("could not get tar image:", err)
	}

	return i, stereoscope.Cleanup
}

func getFixtureImageFromTar(t *testing.T, tarPath string) *image.Image {
	t.Helper()

	request := fmt.Sprintf("docker-archive://%s", tarPath)

	i, err := stereoscope.GetImage(request)
	if err != nil {
		t.Fatal("could not get tar image:", err)
	}

	return i
}

func getFixtureImageInfo(t *testing.T, name string) (string, string) {
	t.Helper()
	version := fixtureVersion(t, name)
	imageName := fmt.Sprintf("%s-%s", imagePrefix, name)
	return imageName, version
}

func LoadFixtureImageIntoDocker(t *testing.T, name string) string {
	t.Helper()
	imageName, imageVersion := getFixtureImageInfo(t, name)
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageVersion)

	if !hasImage(t, fullImageName) {
		contextPath := path.Join(testFixturesDir, name)
		err := buildImage(t, contextPath, imageName, imageVersion)
		if err != nil {
			t.Fatal("could not build fixture image:", err)
		}
	}

	return fullImageName
}

func getFixtureImageTarPath(t *testing.T, fixtureName, tarStoreDir, tarFileName string) string {
	t.Helper()
	imageName, imageVersion := getFixtureImageInfo(t, fixtureName)
	fullImageName := fmt.Sprintf("%s:%s", imageName, imageVersion)
	tarPath := path.Join(tarStoreDir, tarFileName)

	// create the tar-cache dir if it does not already exist...
	if !fileOrDirExists(t, tarCacheDir) {
		err := os.Mkdir(tarCacheDir, 0o755)
		if err != nil {
			t.Fatalf("could not create tar cache dir (%s): %+v", tarCacheDir, err)
		}
	}

	// if the image tar does not exist, make it
	if !fileOrDirExists(t, tarPath) {
		if !hasImage(t, fullImageName) {
			contextPath := path.Join(testFixturesDir, fixtureName)
			err := buildImage(t, contextPath, imageName, imageVersion)
			if err != nil {
				t.Fatal("could not build fixture image:", err)
			}
		}

		err := saveImage(t, fullImageName, tarPath)
		if err != nil {
			t.Fatal("could not save fixture image:", err)
		}
	}

	return tarPath
}

func GetFixtureImageTarPath(t *testing.T, name string) string {
	t.Helper()
	imageName, imageVersion := getFixtureImageInfo(t, name)
	tarFileName := fmt.Sprintf("%s-%s.tar", imageName, imageVersion)
	return getFixtureImageTarPath(t, name, tarCacheDir, tarFileName)
}

func fixtureVersion(t *testing.T, name string) string {
	t.Helper()
	contextPath := path.Join(testFixturesDir, name)
	dockerfileHash, err := dirHash(t, contextPath)
	if err != nil {
		panic(err)
	}
	return dockerfileHash
}

func fileOrDirExists(t *testing.T, filename string) bool {
	t.Helper()
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func dirHash(t *testing.T, root string) (string, error) {
	t.Helper()
	hasher := sha256.New()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			err := f.Close()
			if err != nil {
				panic(err)
			}
		}()

		if _, err := io.Copy(hasher, f); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func hasImage(t *testing.T, imageName string) bool {
	t.Helper()
	cmd := exec.Command("docker", "image", "inspect", imageName)
	cmd.Env = os.Environ()
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

func buildImage(t *testing.T, contextDir, name, tag string) error {
	t.Helper()
	cmd := exec.Command("docker", "build", "-t", name+":"+tag, "-t", name+":latest", ".")
	cmd.Env = os.Environ()
	cmd.Dir = contextDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func saveImage(t *testing.T, image, path string) error {
	t.Helper()

	outfile, err := os.Create(path)
	if err != nil {
		t.Fatal("unable to create file for docker image tar:", err)
	}
	defer func() {
		err := outfile.Close()
		if err != nil {
			panic(err)
		}
	}()

	// note: we are not using -o since some CI providers need root access for the docker client, however,
	// we don't want the resulting tar to be owned by root, thus we write the file piped from stdout.
	cmd := exec.Command("docker", "image", "save", image)
	cmd.Env = os.Environ()

	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
