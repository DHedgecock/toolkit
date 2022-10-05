package toolkit

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)

	if len(s) != 10 {
		t.Error("Wrong length random string")
	}
}

var uploadTests = []struct {
	name         string
	allowedTypes []string
	renameFile   bool
	expectError  bool
}{
	{name: "allowed no rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: false, expectError: false},
	{name: "allowed rename", allowedTypes: []string{"image/jpeg", "image/png"}, renameFile: true, expectError: false},
	{name: "not allowed", allowedTypes: []string{"image/jpeg"}, renameFile: false, expectError: true},
}

func TestTools_UploadFiles(t *testing.T) {
	for _, e := range uploadTests {
		// setup a pipe to avoid buffering
		pipeReader, pipeWriter := io.Pipe()
		writer := multipart.NewWriter(pipeWriter)
		wg := sync.WaitGroup{}
		wg.Add(1)

		go func() {
			defer writer.Close()
			defer wg.Done()

			// create the form data field 'file
			part, err := writer.CreateFormFile("file", "./testdata/img.png")
			if err != nil {
				t.Error(err)
			}

			f, err := os.Open("./testdata/img.png")
			if err != nil {
				t.Error(err)
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				t.Error("error decoding image", err)
			}

			err = png.Encode(part, img)
			if err != nil {
				t.Error(err)
			}
		}()

		// read from the pipe which receives data
		request := httptest.NewRequest("POST", "/", pipeReader)
		request.Header.Add("Content-Type", writer.FormDataContentType())

		var testTools Tools
		testTools.AllowedFileTypes = e.allowedTypes

		uploadedFiles, err := testTools.UploadFiles(request, "./testdata/uploads/", e.renameFile)
		if err != nil && !e.expectError {
			t.Error(err)
		}

		if !e.expectError {
			if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName)); os.IsNotExist(err) {
				t.Errorf("%s: expected file to exist: %s", e.name, err.Error())
			}

			// clean up
			os.Remove(fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].NewFileName))
		}

		if e.expectError && err == nil {
			t.Errorf("%s: error expected but none received", e.name)
		}

		wg.Wait()
	}
}

func TestTools_CreateDirIfNotExists(t *testing.T) {
	var testTools Tools

	err := testTools.CreateDirIfNotExists("./testdata/create_dir_test")
	if err != nil {
		t.Errorf("error creating new directory: %s", err.Error())
	}

	err = testTools.CreateDirIfNotExists("./testdata/create_dir_test")
	if err != nil {
		t.Errorf("error creating dir that already exists: %s", err.Error())
	}

	if _, err := os.Stat("./testdata/create_dir_test"); os.IsNotExist(err) {
		t.Error("directory not actually created")
	}

	os.Remove("./testdata/create_dir_test")
}

func TestTools_Slugify(t *testing.T) {
	var testTools Tools

	slug, err := testTools.Slugify("!OH_MY*lanta")
	if err != nil {
		t.Errorf("error creating slug: %s", err.Error())
	}

	if slug != "oh-my-lanta" {
		t.Errorf("incorrect slug of: %s", slug)
	}
}
