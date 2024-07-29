package controller

import (
	"fmt"
	"net/http"

	"github.com/CloudAceTW/go-gcs-signedurl/model"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

var (
	tracer = otel.Tracer("signurl-controller")
)

func Upload(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "Upload")
	defer span.End()

	signURL := model.NewSignURL(ctx)
	if len(signURL.Err) > 0 {
		span.SetStatus(codes.Error, fmt.Sprintf("err: %+v", signURL.Err))
		c.String(http.StatusInternalServerError, "Internal Error")
		return
	}

	span.SetStatus(codes.Ok, "OK")
	c.JSON(http.StatusOK, &signURL)
}
