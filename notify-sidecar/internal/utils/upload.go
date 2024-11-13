package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func UploadToS3(url string, file io.Reader) error {

	buf := &bytes.Buffer{}
	buf.ReadFrom(file)

	req, err := http.NewRequest("PUT", url, buf)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating request %s: %s", url, err.Error()))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error making request: %s", err.Error()))
	}
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		return errors.New(fmt.Sprintf("AWS Api responded with status %d : %s", resp.StatusCode, b))
	}
	return nil
}
