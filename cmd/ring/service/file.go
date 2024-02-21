package service

import (
	"io"
	"path/filepath"
	"strconv"

	"github.com/ciiim/cloudborad/cmd/ring/ringapi"
	"github.com/gin-gonic/gin"
)

func GetFileContent(ctx *gin.Context) {
	space := ctx.Param("space")

	path := ctx.Param("path")

	base, fileName := filepath.Split(path)

	file, err := ringapi.Ring.GetFile(space, base, fileName)
	if err != nil {
		ctx.JSON(500, gin.H{
			"msg": err.Error(),
		})
		return
	}
	defer file.Close()

	ctx.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	ctx.Writer.Header().Set("Content-Length", strconv.FormatInt(file.FileSize, 10))

	if _, err = io.Copy(ctx.Writer, file); err != nil {
		ctx.JSON(500, gin.H{
			"msg": err.Error(),
		})
		return
	}

}
