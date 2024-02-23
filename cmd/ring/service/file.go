package service

import (
	"io"
	"net/http"
	"path/filepath"

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
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	if _, err = io.Copy(ctx.Writer, file); err != nil {
		ctx.JSON(500, gin.H{
			"msg": err.Error(),
		})
		return
	}

}

func DeleteFile(ctx *gin.Context) {
	space := ctx.Param("space")
	if space == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "space is empty"})
		return
	}

	path := ctx.Param("path")

	base, fileName := filepath.Split(path)

	if err := ringapi.Ring.DeleteFile(space, base, fileName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
}

func UploadFile(ctx *gin.Context) {
	space := ctx.Param("space")
	if space == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "space is empty"})
		return

	}

	//存储路径
	base := ctx.PostForm("base")

	//文件hash
	fileHash := ctx.PostForm("file_hash")

	if fileHash == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "file hash is empty"})
		return
	}

	//文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	reader, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	defer reader.Close()

	fileSize := file.Size
	if err := ringapi.Ring.PutFile(space, base, file.Filename, []byte(fileHash), fileSize, reader); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "success"})
}
