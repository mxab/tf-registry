package upload

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// create archtive and upload to registry host
func UploadDir(dir, host, namespace, name, system, version string) error {
	file, cleanup, err := Zip(dir)
	if err != nil {
		return err
	}
	defer cleanup()

	err = Upload(file, host, namespace, name, system, version)
	if err != nil {
		return err
	}

	return nil
}

// upload to registry host with the /modules/namespaces/name/provider/version/upload endpoint
func Upload(file *os.File, host string, namespace, name, system, version string) error {
	uploadUrl := fmt.Sprintf("%s/v1/modules/%s/%s/%s/%s/upload", host, namespace, name, system, version)

	//open the file
	file, err := os.Open(file.Name())
	if err != nil {
		return err
	}
	res, err := http.Post(uploadUrl, "application/zip", file)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)

		if err != nil {
			fmt.Printf("Failed to parse response body, %v", err)
		}
		fmt.Printf("failed to upload module, %s\n%s\n", res.Status, string(body))
		return fmt.Errorf("failed to upload module, %s", res.Status)
	}

	return nil
}

// Upload uploads a file to S3.
// takes the the dir path, creates a tar zip and uploads to a server with the /.../upload endpoint
func Zip(dir string) (*os.File, func() error, error) {
	file, err := os.CreateTemp("", "*.tf-registry-module.zip")

	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	// new zip writer
	w := zip.NewWriter(file)
	defer w.Close()
	// walk the dir and add files to the zip
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// open the file
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		// create a new file in the zip
		fw, err := w.Create(path)
		if err != nil {
			return err
		}
		// copy the file to the zip
		_, err = io.Copy(fw, f)
		if err != nil {
			return err
		}
		return nil

	}
	err = filepath.Walk(dir, walker)
	if err != nil {
		return nil, nil, err
	}

	return file, func() error {
		return os.Remove(file.Name())
	}, nil

}
