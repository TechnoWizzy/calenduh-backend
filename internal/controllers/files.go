package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/matoous/go-nanoid/v2"
	"net/http"
)

func UploadFile(c *gin.Context) {
	user := *ParseUser(c)
	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	if file.Size > 10<<20 {
		c.AbortWithStatus(http.StatusRequestEntityTooLarge)
		return
	}

	if !isValidFileType(file.Header.Get("Content-Type")) {
		message := gin.H{"message": "File type must be JPEG, PNG, WEBP"}
		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, message)
		return
	}

	client, err := GetS3Client()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	key := c.Query("key")
	if key == "" {
		key, err = gonanoid.New()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if *user.ProfilePicture != key {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key does not match"})
			return
		}
	}

	buffer, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	_, err = client.PutObject(&s3.PutObjectInput{
		Body:         buffer,
		Bucket:       aws.String(util.GetEnv("AWS_BUCKET")),
		Key:          aws.String(key),
		StorageClass: aws.String("GLACIER_IR"),
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.PureJSON(http.StatusOK, key)
	return
}

func DeleteFile(c *gin.Context) {
	user := *ParseUser(c)
	key := c.Param("key")

	if key == "" {
		message := gin.H{"message": "file key is required"}
		c.AbortWithStatusJSON(http.StatusBadRequest, message)
		return
	}

	if *user.ProfilePicture != key {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "key does not match"})
		return
	}

	if err := deleteFile(key); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	if err := database.Db.Queries.DeleteUserProfilePicture(c, &key); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Status(http.StatusOK)
	return
}

func GetS3Client() (*s3.S3, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	client := s3.New(sess)
	return client, nil
}

func deleteFile(key string) error {
	client, err := GetS3Client()
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(util.GetEnv("AWS_BUCKET")),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}

func isValidFileType(fileType string) bool {
	validFileTypes := []string{
		"image/jpeg",
		"image/png",
		"image/webp",
	}

	for _, validFileType := range validFileTypes {
		if fileType == validFileType {
			return true
		}
	}

	return false
}
