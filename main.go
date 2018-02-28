package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"image"
	"log"
	"net/http"
	"strconv"
	"sync"

	"image/jpeg"
	_ "image/png"
)

func main() {
	raw, err := base64.StdEncoding.DecodeString(src64)
	if err != nil {
		log.Fatalf("well that didn't go well: %v", err)
	}

	src, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		log.Fatalf("you should have done better: %v", err)
	}

	jpg := bytes.NewBuffer(make([]byte, 0, 64*1024))
	if err := jpeg.Encode(jpg, src, &jpeg.Options{Quality: 100}); err != nil {
		log.Fatalf("we made it so far: %v", err)
	}

	jpgImg, _, err := image.Decode(bytes.NewReader(jpg.Bytes()))
	if err != nil {
		log.Fatalf("your failures are so mundane: %v", err)
	}

	jpgMu := &sync.Mutex{}
	q := 901
	writeJpg := func() ([]byte, error) {
		out := make([]byte, 64*1024)
		buf := bytes.NewBuffer(make([]byte, 0, 64*1024))

		jpgMu.Lock()
		defer jpgMu.Unlock()

		// Re-encode the jpg
		q--
		if q < 10 {
			q = 999
		}

		if err := jpeg.Encode(buf, jpgImg, &jpeg.Options{Quality: int(q / 10)}); err != nil {
			return nil, err
		}

		// Grab a reference to the new img slice
		raw := buf.Bytes()

		// Copy new img to the scratch buffer
		if len(raw) > len(out) {
			return nil, fmt.Errorf("raw: %d  out: %d", len(raw), len(out))
		}
		n := copy(out, raw)
		out = out[:n]

		// Re-decode the image
		newImg, _, err := image.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, err
		}

		// Overwrite the global image with the worse version
		jpgImg = newImg
		return out, nil
	}

	http.HandleFunc("/ohno/ohno.jpg", func(w http.ResponseWriter, r *http.Request) {
		buf, err := writeJpg()
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(500)
			w.Write([]byte("sometimes bad things happen to good people.\nwe can't be sure that's what happened here, but here's an error message:\n\n" + err.Error()))
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
		w.WriteHeader(200)
		if _, err := w.Write(buf); err == nil {
			h := crc32.ChecksumIEEE(buf)
			log.Printf("%s received %0.8x", r.RemoteAddr, h)
		}
	})

	http.HandleFunc("/ohno", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
		<center>
		<img src="/ohno/ohno.jpg"><br>
		<small>this image gets re-jpg encoded with each visitor.<br>
		every time you visit this page you're helping destroy something beautiful.<br>
		<a href="https://github.com/schmichael.com/ohno">oh no</a>
		</small>
		</center>`))
	})

	log.Printf("serving on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
