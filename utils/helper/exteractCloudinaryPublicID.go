package helper

import (
	"fmt"
	"path/filepath"
	"strings"
)

func ExtractCloudinaryPublicID(imageURL string) (string, error) {
    urlParts := strings.Split(imageURL, "/")
    if len(urlParts) < 2 {
        return "", fmt.Errorf("invalid Cloudinary URL")
    }
    
    uploadIndex := -1
    for i, part := range urlParts {
        if part == "upload" {
            uploadIndex = i
            break
        }
    }
    
    if uploadIndex == -1 || uploadIndex+2 >= len(urlParts) {
        return "", fmt.Errorf("invalid Cloudinary URL structure")
    }
    
    // Combine folder path and filename without extension
    fullPublicID := strings.Join(urlParts[uploadIndex+2:], "/")
    publicID := strings.TrimSuffix(fullPublicID, filepath.Ext(fullPublicID))
    
    return publicID, nil
}