/*cmus-status - Print the artist, song title, status, elapsed time

  Copyright (C) 2021 Brian C. Lane <bcl@brianlane.com>

  This program is free software; you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation; either version 2 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License along
  with this program; if not, write to the Free Software Foundation, Inc.,
  51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	PLAY  = "> "
	STOP  = "# "
	PAUSE = "||"
)

/* commandline flags */
type cmdlineArgs struct {
	Volume  bool // Show volume percentage 0-100
	Elapsed bool // Show the duration:elapsed time
	Width   int  // Maximum width
}

/* commandline defaults */
var cfg = cmdlineArgs{
	Volume:  false,
	Elapsed: false,
	Width:   60,
}

/* parseArgs handles parsing the cmdline args and setting values in the global cfg struct */
func parseArgs() {
	flag.BoolVar(&cfg.Volume, "volume", cfg.Volume, "Include the volume percentage 0-100")
	flag.BoolVar(&cfg.Elapsed, "elapsed", cfg.Elapsed, "Include the duration:elapsed time")
	flag.IntVar(&cfg.Width, "width", cfg.Width, "Maximum width of output")

	flag.Parse()
}

type status struct {
	status    string
	file      string
	duration  int
	position  int
	tags      map[string]string
	vol_left  int
	vol_right int
}

func parseCMUSStatus(response []byte) status {
	status := status{}
	status.tags = make(map[string]string)

	s := bufio.NewScanner(bytes.NewReader(response))
	for s.Scan() {
		f := strings.SplitN(s.Text(), " ", 2)
		if len(f) != 2 {
			continue
		}
		name := f[0]
		value := f[1]

		switch name {
		case "tag":
			t := strings.SplitN(value, " ", 2)
			if len(t) != 2 {
				continue
			}
			status.tags[t[0]] = t[1]
		case "set":
			t := strings.SplitN(value, " ", 2)
			if len(t) != 2 {
				continue
			}
			switch t[0] {
			case "vol_left":
				v, err := strconv.Atoi(t[1])
				if err == nil {
					status.vol_left = v
				}
			case "vol_right":
				v, err := strconv.Atoi(t[1])
				if err == nil {
					status.vol_right = v
				}
			}
		case "status":
			status.status = value
		case "file":
			status.file = value
		case "duration":
			v, err := strconv.Atoi(value)
			if err == nil {
				status.duration = v
			}
		case "position":
			v, err := strconv.Atoi(value)
			if err == nil {
				status.position = v
			}
		}
	}

	return status
}

func (s status) Status() string {
	return s.status
}

func (s status) Title() string {
	song, ok := s.tags["title"]
	if !ok {
		return "Unknown"
	}
	return song
}

func (s status) Artist() string {
	artist, ok := s.tags["artist"]
	if !ok {
		return "Unknown"
	}
	return artist
}

func (s status) Album() string {
	album, ok := s.tags["album"]
	if !ok {
		return "Unknown"
	}
	return album
}

func (s status) Volume() int {
	// Assume volume is balanced
	return s.vol_right
}

func (s status) Duration() string {
	return fmt.Sprintf("%ds", s.duration)
}

func (s status) Position() string {
	return fmt.Sprintf("%ds", s.position)
}

func main() {
	parseArgs()

	// Run 'cmus-remote -Q' and parse the results
	out, err := exec.Command("cmus-remote", "-Q").Output()
	if err != nil {
		log.Fatal(err)
	}
	cmus := parseCMUSStatus(out)

	// Build the final output string
	var s strings.Builder
	switch cmus.Status() {
	case "playing":
		s.WriteString(PLAY)
	case "stopped":
		s.WriteString(STOP)
	case "paused":
		s.WriteString(PAUSE)
	default:
		s.WriteString("  ")
	}

	// Build the optional parts of the output
	if cfg.Volume {
		s.WriteString(fmt.Sprintf("%d%% ", cmus.Volume()))
	}
	if cfg.Elapsed {
		duration, _ := time.ParseDuration(cmus.Duration())
		elapsed, _ := time.ParseDuration(cmus.Position())
		s.WriteString(fmt.Sprintf("%s/%s ", elapsed.Truncate(time.Second), duration.Truncate(time.Second)))
	}

	// Build the Artist + title part (do I want to make artist optional? album?)
	songStr := fmt.Sprintf("%s - %s", cmus.Artist(), cmus.Title())

	// Calculate how much title to trim
	trim := cfg.Width - utf8.RuneCountInString(s.String())
	trim = utf8.RuneCountInString(songStr) - trim
	if trim < 0 {
		trim = 0
	} else if trim > utf8.RuneCountInString(songStr) {
		trim = utf8.RuneCountInString(songStr)
	}
	s.WriteString(songStr[trim:])

	fmt.Printf("%s", s.String())
}
