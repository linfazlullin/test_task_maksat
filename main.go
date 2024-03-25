package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	url                         = "http://linfazlullin.ru/mp3/"
	successfulFile              = "logs/successful.txt"
	notSuccessfulFile           = "logs/not-successful.txt"
	downloadRetryLimit          = 12
	downloadRetryOneMinuteLimit = 9
)

func main() {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting page:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading page:", err)
		return
	}

	re := regexp.MustCompile(`href="([^"]+\.mp3)"`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	files := make([]string, len(matches))
	for i, match := range matches {
		files[i] = match[1]
	}

	successfulFile, err := os.OpenFile(successfulFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening successful file:", err)
		return
	}
	defer successfulFile.Close()

	notSuccessfulFile, err := os.OpenFile(notSuccessfulFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error opening not successful file:", err)
		return
	}
	defer notSuccessfulFile.Close()

	for _, file := range files {
		saveUrl := url + file
		out, err := os.Create("mp3s/" + file)
		if err != nil {
			fmt.Println("Error creating file:", err)
			continue
		}
		defer out.Close()

		for retryCounter := 0; retryCounter <= downloadRetryOneMinuteLimit+downloadRetryLimit; retryCounter++ {
			resp, err := http.Get(saveUrl)
			if err != nil {
				fmt.Println("Error downloading file:", err)

				if retryCounter < downloadRetryLimit {
					fmt.Println("Retrying in 5 seconds...")
					time.Sleep(5 * time.Second)
					continue
				} else if retryCounter < downloadRetryOneMinuteLimit+downloadRetryLimit {
					fmt.Println("Reached retry limit. Pausing 1 minute...")
					time.Sleep(1 * time.Minute)
					continue
				} else {
					fmt.Println("Failed to download file.")
					fmt.Fprintln(notSuccessfulFile, file)
					os.Remove("mp3s/" + file)
					break
				}
			}

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				fmt.Println("Error writing file:", err)
			} else {
				fmt.Fprintln(successfulFile, file)
			}

			resp.Body.Close()
			break
		}
	}
}
