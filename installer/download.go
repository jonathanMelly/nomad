package installer

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// getRequestBody returns the page HTML
func getRequestBody(url string) (string, error) {
	r, err := http.NewRequest("GET", url, nil)

	r.Header.Add("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	r.Header.Add("User-Agent", `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11`)

	if err != nil {
		return "", err
	}

	client, err := http.DefaultClient.Do(r)
	defer client.Body.Close()

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(client.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// downloadFile downloads a file from a URL
func downloadFile(url string, fileName string) (int64, error) {
	out, err := os.Create(fileName)
	defer out.Close()
	if err != nil {
		return 0, err
	}

	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	n, err := io.Copy(out, resp.Body)

	return n, err
}
