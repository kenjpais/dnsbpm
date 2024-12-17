# DNS BPM Calculator

This project implements a DNS-based system for calculating the Beats Per Minute (BPM) of audio files. It allows users to query a DNS server and get the BPM of an audio file as a response. The goal is to provide a fast and efficient way to get BPM data over the network.

## Features
- Calculate BPM for MP3 and WAV audio files.
- Query the BPM over a custom DNS server.
- Optimized for performance, handling audio analysis efficiently.

## Setup

### Prerequisites
- Go 1.18+ installed on your machine.
- A running DNS server setup that supports this project (optional, you can set up your own or use a local server).

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/kenjpais/dnsbpm.git
   cd dnsbpm

2. Clone the repository:
   ```bash
   go mod tidy

### Usage
- Sending Audio File Over the Network Using dig
- To query the BPM of an audio file, we are using the DNS protocol. The audio file is encoded and sent as part of the DNS query, where the file's data is sent as the query string. In this setup, the DNS server is configured to extract the audio 
- data and process it to calculate the BPM.

    ```bash
    $ dig @localhost sample_audio.wav BPM

# Respponse
- The DNS server will return the BPM as a response:
- 128