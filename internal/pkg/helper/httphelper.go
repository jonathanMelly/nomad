package helper

import (
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/pkg/version"
	"io"
	"net/http"
	"os"
	"strings"
)

// ExtractFromRequest will return extracted text from a page at a URL
func ExtractFromRequest(url string, regex string, apiKey string) (*version.Version, error) {
	body, err := getRequestBody(url, apiKey)
	if err != nil {
		return nil, err
	}

	foundVersion, err := version.FromStringCustom(body, regex)
	if err != nil {
		return nil, fmt.Errorf("Could not find version on page:"+url+" | %w", err)
	}

	return foundVersion, nil
}

// getRequestBody returns the page HTML
func getRequestBody(url string, apiKey string) (string, error) {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	if apiKey != "" && strings.Contains(url, "github") {
		r.Header.Add("Authorization", fmt.Sprint("Bearer ", apiKey))
	}
	r.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	r.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)

	client, err := http.DefaultClient.Do(r)
	if err != nil {
		//avoid too much visibility even if secret is not encrypted in binary...
		if r.Header.Get("Authorization") != "" {
			r.Header.Set("Authorization", "*****")
		}
		return "", errors.New(fmt.Sprintln("HTTP error:", r, "|", err))
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Warn("Cannot close http request body", err)
		}
	}(client.Body)

	body, err := io.ReadAll(client.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// DownloadFile downloads a file from a URL
func DownloadFile(url string, fileName string) (int64, error) {
	out, err := os.Create(fileName)
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Errorln("Cannot close", fileName, "|", err)
		}
	}(out)
	if err != nil {
		return 0, err
	}

	r, err := http.NewRequest("GET", url, nil)

	r.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	r.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)

	if err != nil {
		return -1, err
	}

	client, err := http.DefaultClient.Do(r)
	if err != nil {
		return -1, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorln("Cannot close http body", err)
		}
	}(client.Body)

	n, err := io.Copy(out, client.Body)

	return n, err
}
