package toolkit

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const randomStringSource = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_+"

// Tools is the type used to instantiate this module
type Tools struct {
	AllowedFileTypes []string
	MaxFileSize      int
}

// RandomString returns a string of randam characters of length n
func (t *Tools) RandomString(n int) string {
	randomString := make([]rune, n)
	stringSource := []rune(randomStringSource)

	for idx := range randomString {
		primeNumber, _ := rand.Prime(rand.Reader, len(stringSource))
		x := primeNumber.Uint64()
		y := uint64(len(stringSource))

		randomString[idx] = stringSource[x%y]
	}

	return string(randomString)
}

// UploadedFile is a struct used to save information about an uploaded file
type UploadedFile struct {
	NewFileName      string
	OriginalFileName string
	FileSize         int64
}

func (t *Tools) UploadOneFile(r *http.Request, uploadDir string, rename ...bool) (*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	files, err := t.UploadFiles(r, uploadDir, renameFile)
	if err != nil {
		return nil, err
	}

	return files[0], nil
}

// UploadFiles uploads files from a request
func (t *Tools) UploadFiles(r *http.Request, uploadDir string, rename ...bool) ([]*UploadedFile, error) {
	renameFile := true
	if len(rename) > 0 {
		renameFile = rename[0]
	}

	// Create placeholder value for adding saved files to
	var uploadedFiles []*UploadedFile

	// Set a default max size if one wasn't configured
	if t.MaxFileSize == 0 {
		t.MaxFileSize = 1024 * 1024
	}

	err := r.ParseMultipartForm(int64(t.MaxFileSize))
	if err != nil {
		return nil, errors.New("the uploaded file was too big")
	}

	for _, fileHeaders := range r.MultipartForm.File {
		for _, header := range fileHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile

				infile, err := header.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				buff := make([]byte, 512)
				_, err = infile.Read(buff)
				if err != nil {
					return nil, err
				}

				allowed := false
				fileType := http.DetectContentType(buff)

				if len(t.AllowedFileTypes) > 0 {
					for _, x := range t.AllowedFileTypes {
						if strings.EqualFold(fileType, x) {
							allowed = true
						}
					}
				} else {
					allowed = true
				}

				if !allowed {
					return nil, errors.New("the uploaded file type is not permitted")
				}

				// File type is allowed, return to file start to read entire file
				_, err = infile.Seek(0, 0)
				if err != nil {
					return nil, err
				}

				uploadedFile.OriginalFileName = header.Filename
				if renameFile {
					uploadedFile.NewFileName = fmt.Sprintf("%s%s", t.RandomString(25), filepath.Ext(header.Filename))
				} else {
					uploadedFile.NewFileName = header.Filename
				}

				var outfile *os.File
				outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.NewFileName))
				if err != nil {
					return nil, err
				}

				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize

				uploadedFiles = append(uploadedFiles, &uploadedFile)
				return uploadedFiles, nil
			}(uploadedFiles)

			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
