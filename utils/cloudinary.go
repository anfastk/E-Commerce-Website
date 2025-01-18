package utils

import (
	"context"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"mime/multipart"
)

func UploadImageToCloudinary(file multipart.File, fileHeader *multipart.FileHeader, cld *cloudinary.Cloudinary,folder string) (string, error) {
	ctx := context.Background()

	uploadParams, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return "", err
	}
	return uploadParams.SecureURL, nil
}
