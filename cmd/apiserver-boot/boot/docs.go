/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package boot

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bytes"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generates docs for types",
	Long:  `Generates docs for types`,
	Run:   RunDocs,
}

var operations bool

func AddDocs(cmd *cobra.Command) {
	docsCmd.Flags().StringVar(&server, "server", "", "path to apiserver binary to run to get openapi.json")
	docsCmd.Flags().BoolVar(&operations, "operations", false, "if true, include operations in docs.")
	cmd.AddCommand(docsCmd)
}

func RunDocs(cmd *cobra.Command, args []string) {
	if len(server) == 0 {
		fmt.Fprintf(os.Stderr, "apiserver-boot run requires the --server flag\n")
		os.Exit(-1)
	}

	c := exec.Command(server,
		"--delegated-auth=false",
		"--etcd-servers=http://localhost:2379",
		"--secure-port=9443",
		"--print-openapi",
	)
	var b bytes.Buffer
	c.Stdout = &b
	log.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	err = ioutil.WriteFile(filepath.Join("docs", "openapi-spec", "swagger.json"), b.Bytes(), 0644)
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	dir, err := os.Executable()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	dir = filepath.Dir(dir)
	c = exec.Command(filepath.Join(dir, "gen-apidocs"),
		fmt.Sprintf("--build-operations=%v", operations),
		"--allow-errors",
		"--use-tags",
		"--config-dir=docs")
	log.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	// Run the docker command to build the docs
	c = exec.Command("docker", "run",
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, "docs", "includes"), "/source"),
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, "docs", "build"), "/build"),
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, "docs", "build"), "/build"),
		"-v", fmt.Sprintf("%s:%s", filepath.Join(wd, "docs"), "/manifest"),
		"pwittrock/brodocs",
	)
	log.Printf("%s\n", strings.Join(c.Args, " "))
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	err = c.Run()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}