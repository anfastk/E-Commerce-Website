package utils

import (
	"context"
	"errors"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"mime/multipart"
)

var (
	ErrInvalidFile  = errors.New("invalid file")
	ErrUploadFailed = errors.New("upload to Cloudinary failed")
	ErrDeleteFailed = errors.New("delete from Cloudinary failed")
)

func UploadImageToCloudinary(file multipart.File, fileHeader *multipart.FileHeader, cld *cloudinary.Cloudinary, folder string) (string, error) {
	if file == nil || fileHeader == nil || cld == nil {
		return "", ErrInvalidFile
	}
	ctx := context.Background()

	uploadParams, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return "", errors.Join(ErrUploadFailed, err)
	}
	return uploadParams.SecureURL, nil
}

func DeleteCloudinaryImage(cld *cloudinary.Cloudinary, publicID string, c *gin.Context) error {
	if cld == nil || publicID == "" {
		return ErrInvalidFile
	}

	ctx := context.Background()
	_, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		return errors.Join(ErrDeleteFailed, err)
	}

	return nil
}
