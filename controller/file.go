package controller

import (
	"net/http"
	"regexp"
	"time"

	"github.com/CloudAceTW/go-gcs-signedurl/model"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
)

const FilestoreIdReg = `^[0-9a-zA-Z]{7}$`

func File(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "File")
	defer span.End()

	id, _ := c.Params.Get("id")
	reg := regexp.MustCompile(FilestoreIdReg)
	if !reg.MatchString(id) {
		span.SetStatus(codes.Ok, "regex error")
		c.String(http.StatusNotFound, "not found")
		return
	}
	signURL := model.SignURL{
		Id:  id,
		Ctx: ctx,
	}
	signURL.GetGCSPathFromFirestore(ctx)
	if signURL.FileName == "" {
		span.SetStatus(codes.Ok, "doc not found")
		c.String(http.StatusNotFound, "not found")
		return
	}
	err := signURL.MakeGcsSignURL("GET", 15*time.Minute)
	if err != nil {
		span.SetStatus(codes.Error, "make sign url error")
		c.String(http.StatusNotFound, "not found")
		return
	}

	span.SetStatus(codes.Ok, "OK")
	c.JSON(http.StatusOK, &signURL)
}
