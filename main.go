package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/wav"
)

func main() {

	fname := flag.String("f", "", "file path for audio file")
	flag.Int("m", 5, "number of minutes for each segment")
	flag.Parse()

	if *fname == "" {
		log.Fatal("no file path provided")
	}

	f, err := os.Open(*fname)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	count := 1

	for streamer.Position() < streamer.Len() {

		/** Encode snippet to file */
		snippeLen := format.SampleRate.N(time.Duration(1) * time.Hour)
		clip := beep.Take(snippeLen, streamer)
		oname := fmt.Sprintf("test_data/%s_%v.mp3", extractFilenameFromPath(*fname), count)
		out1, err := os.Create(oname)
		if err != nil {
			log.Fatal(err)
		}
		defer out1.Close()
		wav.Encode(out1, clip, format)

		/** Find position to rewind n seconds */
		currPos := streamer.Position()
		cross := format.SampleRate.N(time.Duration(3) * time.Second)
		newPos := currPos - cross
		if newPos < 0 {
			newPos = 0
		}

		/** Set the streamer to the rewound position */
		err = streamer.Seek(newPos)
		if err != nil {
			log.Fatal(err)
		}

		count++
	}

}

func extractFilenameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	filename := strings.TrimSuffix(base, ext)
	return filename
}
