//By TimTheSinner
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

/**
 * Copyright (c) 2016 TimTheSinner All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

type Program struct {
	command    string
	workingDir string
}

func (prog *Program) runCommand(args ...string) bool {
	cmd := exec.Command(prog.command, args...)
	cmd.Dir = prog.workingDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	fmt.Println("Running "+prog.command+" with:", args)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing "+prog.command, err)
		return false
	}
	return true
}

func (prog *Program) runCommandOutput(args ...string) string {
	cmd := exec.Command(prog.command, args...)
	cmd.Dir = prog.workingDir
	cmd.Stderr = os.Stderr

	fmt.Println("Running "+prog.command+" with:", args)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing "+prog.command, err)
	}

	return strings.TrimSpace(string(output))
}

func (prog *Program) runCommandCaptureError(args ...string) string {
	pr, pw := io.Pipe()
	// we need to wait for everything to be done
	wg := sync.WaitGroup{}
	wg.Add(2)

	cmd := exec.Command(prog.command, args...)
	cmd.Dir = prog.workingDir
	cmd.Stdout = os.Stdout
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error executing "+prog.command, err)
	}
	tee := io.TeeReader(stdErr, pw)

	var output string
	go func() {
		defer wg.Done()
		defer pw.Close()

		if b, _err := ioutil.ReadAll(tee); _err != nil {
			log.Fatal(_err)
		} else {
			output = fmt.Sprintf("%s", b)
		}
	}()

	go func() {
		defer wg.Done()

		if _, _err := io.Copy(os.Stderr, pr); _err != nil {
			log.Fatal(_err)
		}
	}()

	fmt.Println("Running "+prog.command+" with:", args)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing "+prog.command, err)
	}

	wg.Wait()
	return strings.TrimSpace(output)
}

func runCommand(command string, args ...string) bool {
	prog := &Program{command: command}
	return prog.runCommand(args...)
}

func runCommandOutput(command string, args ...string) string {
	prog := &Program{command: command}
	return prog.runCommandOutput(args...)
}

func runCommandCaptureError(command string, args ...string) string {
	prog := &Program{command: command}
	return prog.runCommandCaptureError(args...)
}
