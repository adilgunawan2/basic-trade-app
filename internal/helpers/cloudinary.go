package helpers

import (
	"basic-trade-app/config"
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"golang.org/x/net/context"
)

// define allowed file extensions and max file size to upload
var allowedFileExt = map[string]bool {
	".jpg": true,
	".jpeg": true,
	".png": true,
	".svg": true,
}

const maxFileSize = 5 * 1024 * 1024

func validateFile(fileHeader *multipart.FileHeader) error {
	// check file zise
	if fileHeader.Size > maxFileSize {
		return errors.New("file size could not be more than 5 MB") // still wont show up, given the server error instead when file size more than 5MB
	}

	// check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedFileExt[ext] {
		return errors.New("file type is not allowed")
	}

	return nil
}

func UploadFile(fileHeader *multipart.FileHeader, fileName string) (string, error) {
	// validate file type and size
	if err := validateFile(fileHeader)
	err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// add cloudinary product env credentials
	cld, err := cloudinary.NewFromParams(config.EnvCloudName(), config.EnvCloudAPIKey(), config.EnvCloudAPISecret()) 
	if err != nil {
		return "", err
	}

	// convert file
	fileReader, err := convertFile(fileHeader)
	if err != nil {
		return "", err
	}

	// upload file
	uploadParam, err := cld.Upload.Upload(ctx, fileReader, uploader.UploadParams{
		PublicID: fileName,
		Folder: config.EnvCloudUploadFolder(),
	})
	if err != nil {
		return "", err
	}

	return uploadParam.SecureURL, nil
}

func convertFile(fileHeader *multipart.FileHeader) (*bytes.Reader, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read the file content into an in-memory buffer
	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, file); err != nil {
		return nil, err
	}

	// create a bytes.Reader from the buffer
	fileReader := bytes.NewReader(buffer.Bytes())
	return fileReader, nil
}

func RemoveExtension(filename string) string {
	return path.Base(filename[:len(filename)-len(path.Ext(filename))])
}

