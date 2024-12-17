package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dnsbpm/pkg/bpm"
	"github.com/miekg/dns"
)

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)

	for _, q := range r.Question {
		if q.Qtype != dns.TypeTXT {
			continue
		}

		// Decode Base64 audio file from the DNS query name
		encodedAudio := strings.TrimSuffix(q.Name, ".")
		audioData, err := base64.RawURLEncoding.DecodeString(encodedAudio)
		if err != nil {
			log.Printf("Failed to decode Base64: %v", err)
			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
				Txt: []string{"ERR: Invalid audio data"},
			})
			continue
		}

		// Save the audio data to a temporary file using os.CreateTemp
		tempFile, err := os.CreateTemp("", "audio-*.wav")
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
				Txt: []string{"ERR: Temp file error"},
			})
			continue
		}
		defer os.Remove(tempFile.Name())

		// Write audio data to the temporary file
		if _, err := tempFile.Write(audioData); err != nil {
			log.Printf("Failed to write to temp file: %v", err)
			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
				Txt: []string{"ERR: Temp write error"},
			})
			continue
		}
		tempFile.Close()

		// Calculate BPM using the provided BPM logic
		bpmValue, err := bpm.AnalyzeBPM(filepath.Base(tempFile.Name()))
		if err != nil {
			log.Printf("Failed to calculate BPM: %v", err)
			msg.Answer = append(msg.Answer, &dns.TXT{
				Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
				Txt: []string{"ERR: BPM calculation failed"},
			})
			continue
		}

		// Respond with the BPM
		msg.Answer = append(msg.Answer, &dns.TXT{
			Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
			Txt: []string{fmt.Sprintf("%d BPM", bpmValue)},
		})
	}

	w.WriteMsg(msg)
}

func main() {
	server := &dns.Server{Addr: ":5353", Net: "udp"}

	dns.HandleFunc(".", handleDNSRequest)

	log.Println("Starting DNS BPM server on :5353...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}
