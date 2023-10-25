package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	Qwerasd205ClassChangeURL = "https://qwerasd205.github.io/DiscordClassChanges/differences.csv"
	SyndiShanXClassChangeURL = "https://raw.githubusercontent.com/SyndiShanX/Update-Classes/main/Changes.txt"
	TargetURL                = "/mnt/f/Documents/BetterDiscord/YoRHA-UI-BetterDiscord/src"
	TmpFilename              = "replace.txt"
)

func main() {
	if err := downloadFile(TmpFilename, SyndiShanXClassChangeURL); err != nil {
		os.Remove(TmpFilename)
		panic(err)
	}

	f, err := os.Open(TmpFilename)
	if err != nil {
		os.Remove(TmpFilename)
		panic(err)
	}
	defer f.Close()

	if err := SaveSyndiShanX(f); err != nil {
		panic(err)
	}

	os.Remove(TmpFilename)
	fmt.Println("Successfully modified all files! Closing...")
}

func SaveSyndiShanX(f *os.File) error {
	raw, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	repl := strings.Split(string(raw), "\n")
	if len(repl)%2 != 0 {
		return fmt.Errorf("non even number of lines in file")
	}
	fmt.Println(len(repl))

	var wg sync.WaitGroup
	filepath.Walk(TargetURL, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			dat, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}

			datStr := string(dat)
			for i := 0; i < len(repl); i += 2 {
				datStr = strings.ReplaceAll(datStr, repl[i], repl[i+1])
			}

			f, err := os.Create(path)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			f.WriteString(datStr)

			fmt.Printf("finished %s\n", path)
		}(path)

		return nil
	})
	wg.Wait()
	os.Remove(TmpFilename)

	return nil
}

func SaveQwerasd205(f *os.File) {
	s := bufio.NewScanner(f)
	legend := make(map[string]string)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		names := strings.Split(s.Text(), ",")
		// Set all names in comma list to last name (newest)
		for i, name := range names {
			if i == len(names)-1 {
				continue
			}
			legend[name] = names[len(names)-1]
		}
	}

	var wg sync.WaitGroup
	filepath.Walk(TargetURL, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			dat, err := os.ReadFile(path)
			if err != nil {
				panic(err)
			}

			names := string(dat)
			for k, v := range legend {
				names = strings.ReplaceAll(names, k, v)
			}

			f, err := os.Create(path)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			f.WriteString(names)

			fmt.Printf("finished %s\n", path)
		}()

		return nil
	})

	wg.Wait()

	os.Remove(TmpFilename)
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
