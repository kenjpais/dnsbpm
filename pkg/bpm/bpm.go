package bpm

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
)

type Peak struct {
	volume   float64
	position int
}

type Interval struct {
	tempo float64
	count int
}

func getPeaks(data [][]float64) []Peak {
	const partSize = 22050
	parts := len(data[0]) / partSize
	var peaks []Peak

	for i := 0; i < parts; i++ {
		max := Peak{volume: 0, position: 0}
		for j := i * partSize; j < (i+1)*partSize; j++ {
			volume := math.Max(math.Abs(data[0][j]), math.Abs(data[1][j]))
			if volume > max.volume {
				max = Peak{position: j, volume: volume}
			}
		}
		peaks = append(peaks, max)
	}

	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].volume > peaks[j].volume
	})

	peaks = peaks[:len(peaks)/2]
	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].position < peaks[j].position
	})

	return peaks
}

func getIntervals(peaks []Peak) []Interval {
	var groups []Interval

	for index, peak := range peaks {
		for i := 1; index+i < len(peaks) && i < 10; i++ {
			group := Interval{
				tempo: (60 * 44100) / float64(peaks[index+i].position-peak.position),
				count: 1,
			}

			for group.tempo < 90 {
				group.tempo *= 2
			}

			for group.tempo > 180 {
				group.tempo /= 2
			}

			group.tempo = math.Round(group.tempo)

			found := false
			for j := range groups {
				if groups[j].tempo == group.tempo {
					groups[j].count++
					found = true
					break
				}
			}

			if !found {
				groups = append(groups, group)
			}
		}
	}

	return groups
}

func AnalyzeBPM(relFilePath string)(byte, error){
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v", err)
		return 0, fmt.Errorf("Error getting current directory: %v", err)
	}

	filePath := filepath.Join(dir, relFilePath)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("File Open Error: %v", err)
		return 0, fmt.Errorf("File Open Error: %v", err)
	}
	
	defer file.Close()

	var channels [][]float64

	if strings.HasSuffix(filePath, ".mp3") {
		decoder, err := mp3.NewDecoder(file)
		if err != nil {
			fmt.Printf("Error decoding MP3 file: %v", err)
			return 0, fmt.Errorf("Error decoding MP3 file: %v", err)
		}

		var samples []float64
		buf := make([]byte, 1024)
		for {
			n, err := decoder.Read(buf)
			if err != nil {
				break
			}
			for i := 0; i < n; i += 2 {
				sample := float64(int16(buf[i]) | int16(buf[i+1])<<8)
				samples = append(samples, sample)
			}
		}

		// Assuming stereo audio
		channels = [][]float64{
			make([]float64, len(samples)/2),
			make([]float64, len(samples)/2),
		}
		for i := 0; i < len(samples); i += 2 {
			channels[0][i/2] = samples[i]
			channels[1][i/2] = samples[i+1]
		}
	} else if strings.HasSuffix(filePath, ".wav") {
		decoder := wav.NewDecoder(file)
		audioBuffer, err := decoder.FullPCMBuffer()
		if err != nil {
			fmt.Printf("Error decoding WAV file: %v", err)
			return 0, fmt.Errorf("Error decoding WAV file: %v", err)
		}

		channels = make([][]float64, audioBuffer.Format.NumChannels)
		for i := range channels {
			channels[i] = make([]float64, len(audioBuffer.Data))
		}

		for i, sample := range audioBuffer.Data {
			for ch := 0; ch < audioBuffer.Format.NumChannels; ch++ {
				channels[ch][i] = float64(sample)
			}
		}
	} else {
		ext := filepath.Ext(filePath)
		fmt.Printf("Unsupported file format: %s", ext)
		return 0, fmt.Errorf("Unsupported file format: %s", ext)
	}

	groups := getIntervals(getPeaks(channels))

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].count > groups[j].count
	})
	top := groups[:5]

	fmt.Println("Top BPM guesses:")
	for _, group := range top {
		fmt.Printf("%.0f BPM (%d samples)\n", group.tempo, group.count)
	}

	result := byte(top[len(top)-1].tempo)

	fmt.Printf("BPM: %d\n", result)

	if result < 39 || result > 250 {
		fmt.Printf("Error: BPM out of range")
		return 0, fmt.Errorf("Error: BPM out of range")
	}

	return result, nil
}