package helper

import (
	"errors"
	"fmt"
	"github.com/gologme/log"
	"github.com/jonathanMelly/nomad/internal/pkg/data"
	"github.com/jonathanMelly/nomad/pkg/bytesize"
	"github.com/jonathanMelly/nomad/pkg/version"
	"io"
	"net/http"
	"os"
	"strings"
)

const USER_AGENT = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`

// GetVersion will return extracted text from a page at a URL
func GetVersion(url string, definition *data.AppDefinition, apiKey string, requestBody string) (*version.Version, error) {

	responseBody, err := sendRequest(url, apiKey, requestBody)
	if err != nil {
		return nil, err
	}

	foundVersion, err := version.FromStringCustom(responseBody, definition.VersionCheck.RegEx)
	if err != nil {
		return nil, fmt.Errorf("Could not find version on page:"+url+" | %w", err)
	}

	return foundVersion, nil
}

// TODO refactor with BuildAndDoHttp !!!
// sendRequest returns the request response body
func sendRequest(url string, apiKey string, requestBody string) (string, error) {

	var method string
	if requestBody != "" {
		method = "POST"
	} else {
		method = "GET"
	}

	r, err := http.NewRequest(method, url, strings.NewReader(requestBody))
	if err != nil {
		return "", err
	}

	if apiKey != "" && strings.Contains(url, "github") {
		r.Header.Add("Authorization", fmt.Sprint("Bearer ", apiKey))
	}
	r.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	r.Header.Add("User-Agent", USER_AGENT)

	log.Traceln("sending http request to", url, "with payload", requestBody)
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

	log.Traceln("received", bytesize.ByteSize(len(body) /*do not trust client.ContentLength...*/))
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

	client, err := BuildAndDoHttp(url, "GET")
	if err != nil {
		return -1, err
	}

	switch client.StatusCode {
	case 200:
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Errorln("Cannot close http body", err)
			}
		}(client.Body)
		n, err := io.Copy(out, client.Body)
		return n, err
	case 404:
		return -1, errors.New("URL not found")
	default:
		return -1, errors.New(fmt.Sprint("Bad http status: ", client.StatusCode))
	}
}

func BuildAndDoHttp(url string, method string) (*http.Response, error) {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	r.Header.Add("User-Agent", USER_AGENT)

	client, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	return client, nil
}
