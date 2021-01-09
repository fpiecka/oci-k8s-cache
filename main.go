package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
)

var (
	cachePath           = flag.String("cache-path", "~/.oci-k8s-cache", "")
	expirationThreshold = flag.Duration("expiration-threshold", 30*time.Second, "how long before the expiration should we treat the tokens as expired")
	ociPath             = flag.String("oci-path", "oci", "path to the oci binary")
	clusterID           = flag.String("cluster-id", "", "cluster id to pass to the OCI CLI")
	region              = flag.String("region", "us-ashburn-1", "OCI region to use")
)

type credentials struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Status     struct {
		Token               string    `json:"token"`
		ExpirationTimestamp time.Time `json:"expirationTimestamp"`
	} `json:"status"`
}

func main() {
	// hack to support flags at the end of the args
	args := os.Args
	for i, arg := range args {
		if len(arg) > 0 && arg[0] == '-' {
			args = args[i:]
			break
		}
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		panic(err)
	}

	fullOCIPath, err := exec.LookPath(*ociPath)
	if err != nil {
		log.Fatalf("failed to look up the path to the OCI binary: %v", err)
		return
	}

	if *clusterID == "" || *region == "" {
		log.Fatalf("either clusterID (value: %v) or region (value: %v) are empty", clusterID, region)
		return
	}

	if *cachePath == "" {
		log.Fatal("cache path is requried")
		return
	}

	expandedCachePath, err := homedir.Expand(*cachePath)
	if err != nil {
		log.Fatalf("failed to expand the passed cache path: %v", err)
		return
	}
	if err := os.MkdirAll(expandedCachePath, 0700); err != nil {
		log.Fatalf("failed to prepare the directory for the cache: %v", err)
		return
	}

	cachedFilePath := filepath.Join(expandedCachePath, fmt.Sprintf("%s-%s", *region, *clusterID))
	cachedFile, err := processCachedFile(cachedFilePath)
	if err != nil {
		log.Fatalf("failed to process the cached file %s: %v", cachedFilePath, err)
		return
	}
	if cachedFile != nil {
		if _, err := os.Stdout.Write(cachedFile); err != nil {
			panic(err)
		}
		return
	}

	cmd := exec.Command(fullOCIPath, "ce", "cluster", "generate-token", "--cluster-id", *clusterID, "--region", *region)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Fatalf("failed to run the OCI command, error code %d, stderr:\n%v", exitError.ExitCode(), string(exitError.Stderr))
			return
		}
		log.Fatalf("failed to run the OCI command: %v", err.Error())
		return
	}

	var parsedCredentials credentials
	if err := json.Unmarshal(output, &parsedCredentials); err != nil {
		log.Fatalf("failed to parse the credentials output: %v", err)
	}
	if parsedCredentials.Status.ExpirationTimestamp.Before(time.Now().Add(-1 * *expirationThreshold)) {
		log.Fatalf("freshly fetched token has an expired timestamp: %v", string(output))
	}
	if err := ioutil.WriteFile(cachedFilePath, output, 0600); err != nil {
		log.Fatalf("failed to write the new cached file %s: %v", cachedFilePath, err)
	}
	if _, err := os.Stdout.Write(output); err != nil {
		panic(err)
	}
}

func processCachedFile(path string) ([]byte, error) {
	cachedBytes, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var parsedCredentials credentials
	if err := json.Unmarshal(cachedBytes, &parsedCredentials); err != nil {
		return nil, err
	}

	if parsedCredentials.Status.ExpirationTimestamp.Before(time.Now().Add(-1 * *expirationThreshold)) {
		return nil, nil
	}
	return cachedBytes, nil
}
