package service

import (
	"fmt"
	"net/http"

	"github.com/ciiim/cloudborad/cmd/ring/ringapi"
	"github.com/gin-gonic/gin"
)

type BaseAndDir struct {
	Base string `json:"base"`
	Dir  string `json:"dir"`
}

/*
	 JSON Body
		{
			"base": "base",
			"dir": "dir"
		}
*/
func GetSpaceWithDir(ctx *gin.Context) {
	space := ctx.Param("space")
	var baseAndDir BaseAndDir
	if err := ctx.BindJSON(&baseAndDir); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": ErrParam})
		return
	}
	userDir, err := ringapi.Ring.SpaceWithDir(space, baseAndDir.Base, baseAndDir.Dir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"data": userDir,
	})
}

func CreateSpace(ctx *gin.Context) {
	space := ctx.PostForm("space")
	if err := ringapi.Ring.NewSpace(space, 0); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "success"})
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
}

func UploadFileTest(ctx *gin.Context) {
	space := ctx.Param("space")
	base := ctx.PostForm("base")
	file, err := ctx.FormFile("chunk0")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	if base == "" {
		base = "/"
	}
	fmt.Printf("space: %v, base: %v\n", space, base)
	fmt.Printf("file: %v\n", file.Size)
}
