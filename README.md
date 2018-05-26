## go-tv-transcoder

Process a tvheadend recording directory and produce HEVC compressed recodings.  

## Requirements
 - Must have ffmpeg on your path
 - Must have comskip on your path (or build locally)

## Build and Execute
 - This library has [Comskip](https://github.com/erikkaashoek/Comskip/) as a dependency.  It must either be installed in the OS path or built locally.
 - If your building from source, please see the latest build requirements for comskip
 - Once comskip is built, execute `go build`
 - Execution is as simple as `./go-tv-transcoder <SRC_DIR> <DEST_DIR>`
 - More advanced command options like encoder, speed, and format are available

## Mac Build (working 25 May 2018)
 - Install xcode and brew before you build comskip
 - `$ xcode-select --install`
 - `$ /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
 - `$ brew install autoconf automake libtool pkgconfig argtable ffmpeg sdl`
 - `$ cd Comskip`
 - `$ ./autogen.sh`
 - `$ ./configure`
 - `$ make`
 - `$ cd ../`
 - `go build`
