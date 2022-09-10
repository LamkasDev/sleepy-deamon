package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const (
	UploadFileDataBackupDatabase string = "BACKUP_DATABASE"
)

type UploadFileBackupDatabaseData struct {
	Type     string `json:"type"`
	Database string `json:"database"`
	Task     string `json:"task"`
}

func UploadFile(handler *Handler, path string, data any) error {
	dataRaw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"data": bytes.NewReader(dataRaw),
		"file": mustOpen(path),
	}

	var url string
	if handler.Config.Https {
		url = fmt.Sprintf("https://%s/v1/daemon/file/upload", handler.Config.APIHost)
	} else {
		url = fmt.Sprintf("http://%s/v1/daemon/file/upload", handler.Config.APIHost)
	}

	err = Upload(handler, url, values)
	if err != nil {
		return err
	}

	return nil
}

func Upload(handler *Handler, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Cookie", fmt.Sprintf("Token=%s", handler.Config.Token))
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}
