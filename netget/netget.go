package netget

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Getting data from GET-response and returns it, its content type and error, if any
func GetData(address string) ([]byte, string, error) {

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	res, err := client.Get(address)

	if err != nil {
		return nil, "", err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, "", fmt.Errorf("Respone status not OK (%d)", res.StatusCode)
	}

	data, err := ioutil.ReadAll(res.Body)

	return data, res.Header.Get("Content-Type"), err
}
