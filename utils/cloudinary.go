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

func UploadImageToCloudinary(file multipart.File, fileHeader *multipart.FileHeader, cld *cloudinary.Cloudinary, folder string, imageURL string) (string, error) {
    ctx := context.Background()

    var uploadParams *uploader.UploadResult
    var err error

    if imageURL != "" {
        uploadParams, err = cld.Upload.Upload(ctx, imageURL, uploader.UploadParams{
            Folder: folder,
        })
    } else if file != nil && fileHeader != nil {
        uploadParams, err = cld.Upload.Upload(ctx, file, uploader.UploadParams{
            Folder: folder,
        })
    } else {
        return "", errors.New("invalid file or URL")
    }

    if err != nil {
        return "", errors.New("upload failed: " + err.Error())
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
