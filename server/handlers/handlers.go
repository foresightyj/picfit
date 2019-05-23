package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mholt/binding"

	api "gopkg.in/fukata/golang-stats-api-handler.v1"

	"github.com/thoas/picfit/application"
	"github.com/thoas/picfit/constants"
	"github.com/thoas/picfit/errs"
	"github.com/thoas/picfit/payload"
	"github.com/thoas/picfit/storage"
)

func StatsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, api.GetStats())
}

// Healthcheck displays an ok response for healthcheck
func Healthcheck(startedAt time.Time) func(c *gin.Context) {
	return func(c *gin.Context) {
		now := time.Now().UTC()

		uptime := now.Sub(startedAt)

		c.JSON(http.StatusOK, gin.H{
			"started_at": startedAt.String(),
			"uptime":     uptime.String(),
			"status":     "Ok",
			"version":    constants.Version,
			"revision":   constants.Revision,
			"build_time": constants.BuildTime,
			"compiler":   constants.Compiler,
			"ip_address": c.ClientIP(),
		})
	}
}

// Display displays and image using resizing parameters
func Display(c *gin.Context) {
	file, err := application.ImageFileFromContext(c, true, true)
	if err != nil {
		errs.Handle(err, c.Writer)

		return
	}

	for k, v := range file.Headers {
		c.Header(k, v)
	}

	c.Data(http.StatusOK, file.ContentType(), file.Content())
}

// Upload uploads an image to the destination storage
func Upload(c *gin.Context) {
	multipartPayload := new(payload.MultipartPayload)
	errs := binding.Bind(c.Request, multipartPayload)
	if errs != nil {
		c.String(http.StatusBadRequest, errs.Error())
		return
	}

	file, err := multipartPayload.Upload(storage.DestinationFromContext(c))

	if err != nil {
		c.String(http.StatusBadRequest, errs.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": file.Filename(),
		"path":     file.Path(),
		"url":      file.URL(),
	})
}

// Delete deletes a file from storages
func Delete(c *gin.Context) {
	var (
		err         error
		path        = c.Param("parameters")
		key, exists = c.Get("key")
	)

	if path == "" && !exists {
		c.String(http.StatusUnprocessableEntity, "no path or key provided")
		return
	}

	if !exists {
		err = application.Delete(c, path[1:])
	} else {
		err = application.DeleteChild(c, key.(string))
	}

	if err != nil {
		errs.Handle(err, c.Writer)

		return
	}

	c.String(http.StatusOK, "Ok")
}

// Get generates an image synchronously and return its information from storages
func Get(c *gin.Context) {
	file, err := application.ImageFileFromContext(c, false, false)
	if err != nil {
		errs.Handle(err, c.Writer)

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": file.Filename(),
		"path":     file.Path(),
		"url":      file.URL(),
		"key":      file.Key,
	})
}

// Redirect redirects to the image using base url from storage
func Redirect(c *gin.Context) {
	file, err := application.ImageFileFromContext(c, false, false)
	if err != nil {
		errs.Handle(err, c.Writer)

		return
	}

	c.Redirect(http.StatusMovedPermanently, file.URL())
}

func Pprof(h http.HandlerFunc) gin.HandlerFunc {
	handler := http.HandlerFunc(h)
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
