package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CloudAceTW/go-gcs-signedurl/model"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/codes"
)

func Done(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "upload/done")
	defer span.End()

	var signURL model.SignURL
	err := c.BindJSON(&signURL)
	if err != nil {
		log.Printf("c.BindJSON err: %v", err)
		span.SetStatus(codes.Error, err.Error())
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}
	signURL.Ctx = ctx

	signURL.GetFileNameFromRedis(ctx)
	if len(signURL.Err) > 0 {
		span.SetStatus(codes.Error, fmt.Sprintf("err: %+v", signURL.Err))
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}

	signURL.SetGCSPathToFirestore(ctx)
	if len(signURL.Err) > 0 {
		span.SetStatus(codes.Error, fmt.Sprintf("err: %+v", signURL.Err))
		c.String(http.StatusBadRequest, "Bad Request")
		return
	}

	span.SetStatus(codes.Ok, "OK")
	c.JSON(http.StatusOK, &signURL)
}
