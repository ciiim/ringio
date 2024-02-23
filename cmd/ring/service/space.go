package service

import (
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
