package main

import (
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Info struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	UploadDate string `json:"upload_date"`
}

func usage() {
	log.Fatalln("usage: yt-dlp-renamer /path/to/videos")
}

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()

	dir := flag.Arg(0)
	if dir == "" {
		usage()
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}

	infos, err := buildInfo(dir, entries)
	if err != nil {
		log.Fatalln(err)
	}

	matches, unmatched := match(dir, entries, infos)
	err = rename(dir, matches)
	if err != nil {
		log.Fatalln(err)
	}

	// Print out any unmatched files.
	if len(unmatched) > 0 {
		log.Println("unmatched files:")
		for _, u := range unmatched {
			log.Println(u)
		}
	}
}

func buildInfo(dir string, entries []fs.DirEntry) ([]Info, error) {
	var infos []Info
	for _, e := range entries {
		// Loop over info json files.
		if e.IsDir() {
			continue
		}

		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}

		var i Info
		err = json.Unmarshal(data, &i)
		if err != nil {
			return nil, err
		}
		infos = append(infos, i)
	}
	return infos, nil
}

func match(dir string, entries []fs.DirEntry, infos []Info) (
	map[string]Info,
	[]string,
) {
	matches := make(map[string]Info)
	var unmatched []string
	for _, e := range entries {
		// Loop over videos.
		if strings.HasSuffix(e.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, e.Name())
		title := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		for _, i := range infos {
			// Loop over all infos for each video and try to find a match.
			if strings.Contains(title, i.ID) {
				matches[path] = i
			}

			if strings.Contains(title, i.Title) {
				matches[path] = i
			}

			unmatched = append(unmatched, path)
		}
	}
	return matches, unmatched
}

func rename(dir string, matches map[string]Info) error {
	for p, i := range matches {
		ext := filepath.Ext(p)
		title := strings.ReplaceAll(i.Title, "/", "_")
		name := i.UploadDate + " - " + title + ext
		err := os.Rename(p, filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
