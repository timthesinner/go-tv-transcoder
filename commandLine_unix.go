// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

//By TimTheSinner
package main

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
	return runCommand("./Comskip/comskip", "--ini="+iniFile, "-ts", "--hwassist", "-w", "--output="+outputDir, localEpisode)
}

func checkForInterlaced(ffmpeg *Program, episode string) string {
	return ffmpeg.runCommandCaptureError("-hide_banner", "-i", episode, "-map", "0:0", "-vf", "idet", "-frames:v", "5000", "-an", "-c", "rawvideo", "-y", "-f", "rawvideo", "/dev/null")
}

func remove(fileName string) bool {
	return runCommand("rm", "-f", fileName)
}

func copy(source, destination string) bool {
	return runCommand("cp", source, destination)
}
