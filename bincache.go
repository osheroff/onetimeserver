package onetimeserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

const baseURL = "https://github.com/osheroff/onetimeserver-binaries/raw/master"

func getBinaryCachePath(pkg string, subpath string, program string, version string) string {
	dir := fmt.Sprintf("%s/.onetimeserver/bin/%s/%s%s", os.Getenv("HOME"), pkg, version, subpath)
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprint(dir, "/", program)
}

func makeHTTPRequest(pkg string, subpath string, os string, program string, version string) *http.Response {
	url := fmt.Sprintf("%s/%s/%s/%s/%s?raw=true", baseURL, pkg, os, version, program)
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
	print(pkg)
	dir := buildInstallCachePath(pkg, version)
	dirFiles := fmt.Sprintf("%s/", dir)
	print(dir)
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

	sourceFiles := fmt.Sprintf("%s/", sourcePath)
	cmd := exec.Command("cp", "-avp", sourceFiles, dir)
	print(cmd)
	cmd.Run()
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
	return path
}

func MakeSymlink(pkg string, subpath string, program string, version string, alias string) {
	linkFrom := getBinaryCachePath(pkg, subpath, program, version)
	linkTo := getBinaryCachePath(pkg, subpath, alias, version)
	os.Symlink(linkFrom, linkTo)
}
