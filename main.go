//By TimTheSinner
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	astisub "github.com/asticode/go-astisub"
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

//Adapted from https://raw.githubusercontent.com/BrettSheleski/comchap/master/comchap

//Config Metadata Datum
type Config struct {
	Ffmpeg  string `json:"ffmpeg"`
	Comskip string `json:"comskip"`
}

type Chapter struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Title string  `json:"title"`
}

type Transcode struct {
	Original    string            `json:"original"`
	Transcoded  string            `json:"transcoded"`
	Series      string            `json:"series"`
	Episode     string            `json:"episode"`
	Number      string            `json:"number"`
	Interlacing map[string]string `json:"idef"`
	Chapters    []*Chapter        `json:"chapters"`
}

const metadata = `;FFMETADATA1
title={{$.Title}}
language=English
artist=TimTheSinner
{{range $.Chapters}}
[CHAPTER]
TIMEBASE=1/1000
START={{printf "%.0f" .Start}}
END={{printf "%.0f" .End}}
title={{.Title}}
{{end}}`

func handle(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var regex = regexp.MustCompile(`(?P<series>.+)\..*;-?(?P<episode>.*?)-?(?P<number>S\d+E\d+).*\.ts`)
var deinterlaceRegex = regexp.MustCompile(`(?i)Multi frame detection: TFF:\s+(?P<tff>\d+)\s*BFF:\s+(?P<bff>\d+)\s*Progressive:\s+(?P<progressive>\d+)\s*Undetermined:\s+(?P<undetermined>\d+)`)

func checkInterlaced(episode string) (bool, map[string]string) {
	metadata := runCommandCaptureError("ffmpeg", "-hide_banner", "-i", episode, "-map", "0:0", "-vf", "idet", "-frames:v", "5000", "-an", "-c", "rawvideo", "-y", "-f", "rawvideo", "/dev/null")
	results := groupsFromRegex(deinterlaceRegex, metadata)

	tff, err := strconv.Atoi(results["tff"])
	handle(err)

	bff, err := strconv.Atoi(results["bff"])
	handle(err)

	progressive, err := strconv.Atoi(results["progressive"])
	handle(err)

	// If tff or bff is > progressive then the video is 'interlaced'
	// http://www.aktau.be/2013/09/22/detecting-interlaced-video-with-ffmpeg/
	return tff > progressive || bff > progressive, results
}

func episodeMeta(episode string) map[string]string {
	md := map[string]string{}
	_, fileName := filepath.Split(episode)

	names := regex.SubexpNames()
	matches := regex.FindAllStringSubmatch(fileName, -1)
	if len(matches) == 0 {
		fmt.Printf("Episode %s does not match regex\n", fileName)
		return md
	}

	for i, match := range matches[0] {
		md[names[i]] = strings.TrimSpace(strings.Replace(match, "-", " ", -1))
	}
	return md
}

func subtitles(dir string, localEpisode string) string {
	episodeSrt := path.Join(dir, "episode-raw.srt")
	runCommand("ffmpeg", "-f", "lavfi", "-i", fmt.Sprintf("movie=%s[out0+subcc]", localEpisode), episodeSrt)

	rawSubs, err := ioutil.ReadFile(episodeSrt)
	handle(err)

	cleanSubs := strings.Replace(string(rawSubs), "\\h", "", -1)
	subs, err := astisub.ReadFromSRT(bytes.NewReader([]byte(cleanSubs)))
	handle(err)

	// Optimize subtitles
	subs.Optimize()

	// Unfragment the subtitles
	subs.Unfragment()

	outputSrt := path.Join(dir, "episode.srt")
	handle(subs.Write(outputSrt))
	return outputSrt
}

func chapters(dir string, edl string, duration float64) []*Chapter {
	edlFile, err := os.Open(path.Join(dir, edl))
	handle(err)
	defer edlFile.Close()

	var commercials []*Chapter
	scanner := bufio.NewScanner(edlFile)
	counter := 0
	for scanner.Scan() {
		bounds := strings.Fields(scanner.Text())
		start, err := strconv.ParseFloat(bounds[0], 64)
		handle(err)
		end, err := strconv.ParseFloat(bounds[1], 64)
		handle(err)
		commercials = append(commercials, &Chapter{Start: start * 1000, End: end * 1000, Title: fmt.Sprintf("Commercial %v", counter)})
		counter++
	}

	var chapters []*Chapter
	if commercials[0].Start > 0.0 {
		chapters = append(chapters, &Chapter{Start: 0, End: commercials[0].Start, Title: "Content 0"})
	}

	var previous *Chapter
	for i, commercial := range commercials {
		if i == 0 {
			previous = commercial
			chapters = append(chapters, commercial)
			continue
		}

		chapters = append(chapters, &Chapter{Start: previous.End, End: commercial.Start, Title: fmt.Sprintf("Content %v", i+1)}, commercial)
		previous = commercial
	}

	if duration*1000 > previous.End {
		chapters = append(chapters, &Chapter{Start: previous.End, End: duration * 1000, Title: fmt.Sprintf("Content %v", len(commercials)+1)})
	}

	return chapters
}

func transcode(episode, outputDir, tempDir string, episodeMeta map[string]string) *Transcode {
	dir, err := ioutil.TempDir(tempDir, "transcode-with-commercials")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	localEpisode := path.Join(dir, "episode.ts")
	runCommand("cp", episode, localEpisode)

	// Export and clean subtitles
	episodeSrt := subtitles(dir, localEpisode)

	iniFile := path.Join(dir, "comskip.ini")
	handle(ioutil.WriteFile(iniFile, []byte("output_edl=1\n"), 0644))

	runCommand("./Comskip/comskip", "--ini="+iniFile, "-ts", "--hwassist", "-w", "--output="+dir, localEpisode)
	duration, err := strconv.ParseFloat(runCommandOutput("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", localEpisode), 64)
	handle(err)

	chapters := chapters(dir, "episode.edl", duration)
	episodeTitle := fmt.Sprintf("%s.%s.%s", episodeMeta["series"], episodeMeta["number"], episodeMeta["episode"])

	metaTemplate := template.New("main")
	metaTemplate, _ = metaTemplate.Parse(metadata) // parsing of template string
	var tpl bytes.Buffer
	metaTemplate.Execute(&tpl, map[string]interface{}{
		"Title":    episodeTitle,
		"Chapters": chapters,
	})

	metadataFile := path.Join(dir, "episode.ffmpeg.metadata")
	ffmpegMetadata := strings.TrimSpace(tpl.String()) + "\n"
	ioutil.WriteFile(metadataFile, []byte(ffmpegMetadata), 0644)

	showDir := filepath.Join(outputDir, episodeMeta["series"])
	os.MkdirAll(showDir, os.ModePerm)
	outputEpisode := filepath.Join(showDir, fmt.Sprintf("%s.mkv", episodeTitle))

	// Check for need to deinterlace
	interlaced, frameData := checkInterlaced(localEpisode)
	vf := "scale=1920:-2"
	if interlaced {
		vf += ", yadif"
	}

	transcodeArgs := []string{
		"-nostdin",
		"-hide_banner",
	}

	if strings.TrimSpace(*hwaccel) != "" {
		transcodeArgs = append(transcodeArgs, "-hwaccel", *hwaccel)
	}

	transcodeArgs = append(transcodeArgs,
		"-analyzeduration", "250M", "-probesize", "250M",
		"-i", localEpisode,
		"-max_muxing_queue_size", "4096",
		"-i", episodeSrt,
		"-i", metadataFile,
		"-map_metadata", "1",
		"-map", "0:v:0", "-map", "0:a:0", "-map", "1:s",
		"-c:v", *codec, "-vf", vf, "-crf", strconv.Itoa(*crf), "-preset", *speed,
		"-pix_fmt", *pixFmt, "-tune", "fastdecode", "-movflags", "+faststart",
		"-c:a", "libopus", "-af", "aformat=channel_layouts='7.1|6.1|5.1|stereo'",
		"-c:s", "copy", outputEpisode)

	runCommand("rm", "-f", outputEpisode)
	result := runCommand("ffmpeg", transcodeArgs...)
	if !result {
		return nil
	}

	return &Transcode{
		Original:    episode,
		Transcoded:  outputEpisode,
		Series:      episodeMeta["series"],
		Episode:     episodeMeta["episode"],
		Number:      episodeMeta["number"],
		Interlacing: frameData,
		Chapters:    chapters,
	}
}

func readMetadata(seriesDir string) (ret map[string]*Transcode) {
	seriesMeta, err := os.Open(path.Join(seriesDir, "transcode-metadata.json"))
	if err != nil {
		return make(map[string]*Transcode)
	}
	defer seriesMeta.Close()

	json.NewDecoder(seriesMeta).Decode(&ret)
	return
}

func writeMetadata(seriesDir string, metadata map[string]*Transcode) {
	seriesMeta, err := os.OpenFile(path.Join(seriesDir, "transcode-metadata.json"), os.O_WRONLY|os.O_CREATE, 0644)
	handle(err)
	defer seriesMeta.Close()

	json.NewEncoder(seriesMeta).Encode(metadata)

}

var hwaccel = flag.String("hwaccel", "", "Hardware Acceleration Driver")
var crf = flag.Int("crf", 21, "CRF (Quality Factor)")
var codec = flag.String("codec", "libx265", "Video encoding codec")
var speed = flag.String("speed", "medium", "Encoder speed")
var pixFmt = flag.String("pix_fmt", "yuv420p", "Video color depth, dont go deeper than yuv420p if your encoding for a pi")

func main() {
	flag.Parse()

	outputDir := "/Volumes/downloads/tv/"
	tempDir := "/Volumes/TEMP_MEDIA/"
	recordings := "/Volumes/recordings/"

	if flag.NArg() > 3 {
		outputDir = flag.Arg(0)
		tempDir = flag.Arg(1)
		recordings = flag.Arg(2)
	}

	series, err := ioutil.ReadDir(recordings)
	handle(err)

	for _, seriesName := range series {
		if !seriesName.IsDir() {
			continue
		}

		seriesDir := filepath.Join(recordings, seriesName.Name())
		episodes, err := ioutil.ReadDir(seriesDir)
		handle(err)

		seriesMeta := readMetadata(seriesDir)
		for _, file := range episodes {
			if !file.IsDir() {
				episode := filepath.Join(seriesDir, file.Name())
				episodeMeta := episodeMeta(episode)
				if number, ok := episodeMeta["number"]; ok {
					if _, ok := seriesMeta[number]; !ok {
						transcodeMeta := transcode(episode, outputDir, tempDir, episodeMeta)
						if transcodeMeta != nil {
							seriesMeta[number] = transcodeMeta
							writeMetadata(seriesDir, seriesMeta)
						}
					}
				}
			}
		}
	}
}
