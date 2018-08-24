//By TimTheSinner
package main

import (
	"fmt"
	"io"
	"os"
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

func comskip(iniFile, outputDir, localEpisode string) bool {
	com := &Program{command: "comskip", workingDir: outputDir}
	com.runCommand("--ini="+iniFile, "--ts", "--hwassist", localEpisode)
	return true
}

func checkForInterlaced(ffmpeg *Program, episode string) string {
	return ffmpeg.runCommandCaptureError("-hide_banner", "-i", episode, "-map", "0:0", "-vf", "idet", "-frames:v", "5000", "-an", "-c", "rawvideo", "-y", "-f", "rawvideo", "NUL")
}

func remove(fileName string) bool {
	return runCommand("rm", "-f", fileName)
}

func copy(source, destination string) (res bool) {
	in, err := os.Open(source)
	if err != nil {
		res = false
		return
	}
	defer in.Close()

	fmt.Println("Copying: " + source + " to " + destination)
	out, err := os.Create(destination)
	if err != nil {
		res = false
		return
	}

	defer func() {
		cerr := out.Close()
		if cerr != nil {
			res = false
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		res = false
		return
	}

	err = out.Sync()
	res = (err == nil)
	return
}
