package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/wav"
)

func main() {

	// Parse and validate flags
	fname := flag.String("f", "", "file path for audio file")
	mins := flag.Int("m", 5, "number of minutes for each segment")
	flag.Parse()

	if *fname == "" {
		log.Fatal("no file path provided")
	}

	// Get file reference
	f, err := os.Open(*fname)
	if err != nil {
		log.Fatal(err)
	}

	// Decode mp3 to get sample rate from format object
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	iterations := int(math.Ceil(float64(streamer.Len()) / float64(format.SampleRate.N(time.Duration(*mins)*time.Minute))))

	// Instance a wait group
	var wg sync.WaitGroup
	wg.Add(iterations)

	for i := 0; i < iterations; i++ {

		duration := format.SampleRate.N(time.Duration(*mins) * time.Minute)
		startPos := duration*i - format.SampleRate.N(time.Duration(3)*time.Second)
		if startPos < 0 {
			startPos = 0
		}

		go func(c int, s int, d int) {
			defer wg.Done()

			// Get file reference
			f, err := os.Open(*fname)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("Started processing snipped %v from %v to %v\n", c, s, d)

			/** Create a streamer from file reference */
			streamer, _, err := mp3.Decode(f)
			if err != nil {
				log.Fatal(err)
			}
			defer streamer.Close()

			/** Set start position of streamer */
			err = streamer.Seek(s)
			if err != nil {
				log.Fatal(err)
			}

			/** Encode snippet to file */
			snippeLen := format.SampleRate.N(time.Duration(*mins) * time.Minute)
			clip := beep.Take(snippeLen, streamer)
			base := fmt.Sprintf("%s/split/", extractBaseFromPath(*fname))
			createDir(base)
			oname := fmt.Sprintf("%s/%03d_%s.mp3", base, c, extractFilenameFromPath(*fname))

			fmt.Printf("Encoding snipped %s\n", oname)

			out1, err := os.Create(oname)
			if err != nil {
				log.Fatal(err)
			}
			defer out1.Close()
			wav.Encode(out1, clip, format)

		}(i+1, startPos, duration)

	}

	wg.Wait()
}

func extractFilenameFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	filename := strings.TrimSuffix(base, ext)
	return filename
}

func extractBaseFromPath(path string) string {
	return filepath.Dir(path)
}

func createDir(dirName string) {
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}
}
