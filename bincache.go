package onetimeserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
)

const baseURL = "https://raw.githubusercontent.com/osheroff/onetimeserver-binaries/master"

func getBinaryCachePath(pkg string, subpath string, program string, version string) string {
	dir := fmt.Sprintf("%s/.onetimeserver/bin/%s/%s%s", os.Getenv("HOME"), pkg, version, subpath)
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprint(dir, "/", program)
}

func GetBinary(pkg string, subpath string, program string, version string) string {
	path := getBinaryCachePath(pkg, subpath, program, version)
	_, err := os.Stat(path)
	if err == nil {
		return path
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s?raw=true", baseURL, pkg, runtime.GOOS, version, program)
	log.Printf("fetching %s\n", url)
	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal(fmt.Sprintf("Got status %d fetching %s", resp.StatusCode, url))
	}

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
