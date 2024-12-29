package onetimeserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const baseURL = "https://github.com/osheroff/onetimeserver-binaries/raw/master"

func getBinaryCachePath(pkg string, subpath string, program string, version string) string {
	if strings.HasSuffix(program, ".gz") { 
		program = strings.Replace(program, ".gz", "", 0)
	}
	dir := fmt.Sprintf("%s/.onetimeserver/bin/%s/%s%s", os.Getenv("HOME"), pkg, version, subpath)
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprint(dir, "/", program)
}

func makeHTTPRequestURL(url string) *http.Response {
	log.Printf("fetching %s\n", url)
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode == 404 {
		return nil
	} else if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf("Got status %d fetching %s", resp.StatusCode, url))
		return nil
	} else {
		return resp
	}
}

func makeHTTPRequest(pkg string, subpath string, os string, program string, version string) *http.Response {
	url := fmt.Sprintf("%s/%s/%s/%s/%s?raw=true", baseURL, pkg, os, version, program)
	return makeHTTPRequestURL(url)
}

func buildInstallCachePath(pkg string, version string) string {
	return fmt.Sprintf("%s/.onetimeserver/install/%s/%s", os.Getenv("HOME"), pkg, version)
}

func GetInstallPathCache(pkg string, version string) string {
	dir := fmt.Sprintf("%s/.onetimeserver/install/%s/%s", os.Getenv("HOME"), pkg, version)
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		log.Fatal(err)
	}

	return dir
}

func CopyFromInstallCache(pkg string, version string, destPath string) bool {
	dir := buildInstallCachePath(pkg, version)
	dirFiles := fmt.Sprintf("%s/.", dir)
	stat, err := os.Stat(dir)
	if err == nil && stat.IsDir() {
		os.MkdirAll(destPath, 0755)
		cmd := exec.Command("cp", "-avp", dirFiles, destPath)
		cmd.Run()
		return true
	} else {
		return false
	}
}

func CopyToInstallCache(pkg string, version string, sourcePath string) {
	dir := buildInstallCachePath(pkg, version)
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		log.Fatal(err)
	}

	sourceFiles := fmt.Sprintf("%s/.", sourcePath)
	cmd := exec.Command("cp", "-avp", sourceFiles, dir)
	cmd.Run()
}

func RemoveFromInstallCache(pkg string, version string, sourcePath string, filename string) {
	dir := buildInstallCachePath(pkg, version)
	file := fmt.Sprintf("%s/%s", dir, filename)
	os.Remove(file)
}

func GetBinary(pkg string, subpath string, program string, version string) string {
	path := getBinaryCachePath(pkg, subpath, program, version)
	_, err := os.Stat(path)
	if err == nil {
		return path
	}

	resp := makeHTTPRequest(pkg, subpath, runtime.GOOS, program, version)
	if resp == nil {
		resp = makeHTTPRequest(pkg, subpath, runtime.GOOS, program, "common")
	}
	if resp == nil {
		resp = makeHTTPRequest(pkg, subpath, "common", program, version)
	}
	if resp == nil {
		log.Fatal(fmt.Sprintf("Couldn't find %s/%s %s for platform", pkg, program, version))
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Write(body)


	if strings.HasSuffix(path, ".gz") {
		log.Printf("unzipping")
		cmd := exec.Command("gunzip", path)
		cmd.Run()
	}

	return path
}

func GetManifest(pkg string) string {
	manifestPath := fmt.Sprintf("%s/.onetimeserver/bin/%s/manifest.json", os.Getenv("HOME"), pkg)
	_, err := os.Stat(manifestPath)
	if err == nil {
		log.Printf("Using cached manifest at %s\n", manifestPath)
		return manifestPath
	}
	url := fmt.Sprintf("%s/%s/manifest.json?raw=true", baseURL, pkg)
	resp := makeHTTPRequestURL(url)
	if resp == nil {
		log.Fatal(fmt.Sprintf("Couldn't find manifest at %s", url))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	file, err := os.OpenFile(manifestPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	file.Write(body)
	log.Printf("downloaded manifest to %s\n", manifestPath)
	return manifestPath
}

func downloadRange(pkg string, version string, files []interface{}) {
	for _, item := range files {
		file := item.(string)
		split := strings.Split(file, "/")
		GetBinary(pkg, strings.Join(split[0:len(split)-1], "/"), split[len(split)-1], version)
	}
}

func DownloadFromManifest(pkg string, version string) {
	path := GetManifest(pkg)
	var result map[string]interface{}

	jsonFile, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()
	bytes, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(bytes, &result)

	all, ok := result["all"]
	if !ok {
		log.Fatal("No key for 'all' in manifest.json!")
	}
	downloadRange(pkg, version, all.([]interface{}))

	ver, version_ok := result[version]
	if !version_ok {
		log.Fatal(fmt.Sprintf("No key for '%s' in manifest.json!", version))
	}
	version_map := ver.(map[string]interface{})

	common, ok := version_map["common"]
	if !ok {
		log.Fatal(fmt.Sprintf("No key for '%s/common' in manifest.json!", version))
	}

	downloadRange(pkg, version, common.([]interface{}))

	platform, platform_ok := version_map[runtime.GOOS]
	if platform_ok {
		downloadRange(pkg, version, platform.([]interface{}))
	}
}

func MakeSymlink(pkg string, subpath string, program string, version string, alias string) {
	linkFrom := getBinaryCachePath(pkg, subpath, program, version)
	linkTo := getBinaryCachePath(pkg, subpath, alias, version)
	fmt.Printf("symlinking from %s to %s\n", linkFrom, linkTo)
	os.Symlink(linkFrom, linkTo)
}
