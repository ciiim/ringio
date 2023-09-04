package router

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ciiim/cloudborad/models"
	"github.com/gin-gonic/gin"
)

const (
	Board_Success = 2000
	Board_Failed  = 2001
	Board_FormErr = 2002
)

type newBoardForm struct {
	Name      string `json:"name" form:"name" binding:"required"`
	Owner_uid int64  `json:"owner_uid" form:"owner_uid" binding:"required"`
	Passwd    string `json:"passwd" form:"passwd"`
	Visible   int    `json:"visible" form:"visible"`
	Pinned    int    `json:"pinned" form:"pinned"`
}

func (a *ApiServer) NewBoard(c *gin.Context) {
	newBoard := &newBoardForm{}
	if err := c.ShouldBindJSON(newBoard); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Board_FormErr, false, err.Error(), nil))
		return
	}

	if err := a.service.NewBoard(a.service.ToBoard(newBoard.Owner_uid, newBoard.Name, newBoard.Passwd, newBoard.Visible, newBoard.Pinned)); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"board": newBoard,
		},
	))

}

func (a *ApiServer) UpdateBoard(c *gin.Context) {
	board := &models.Board{}
	if err := c.ShouldBindJSON(board); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Board_FormErr, false, err.Error(), nil))
		return
	}
	if err := a.service.UpdateBoard(board); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"board": board,
		},
	))

}

func (a *ApiServer) DeleteBoard(c *gin.Context) {
	boardBasic := &models.Board_basic{}
	if err := c.ShouldBindJSON(boardBasic); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Board_FormErr, false, err.Error(), nil))
		return
	}
	if err := a.service.DeleteBoard(boardBasic); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"board": boardBasic,
		},
	))
}

func (a *ApiServer) GetAllBoardBasic(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Query("uid"))
	boards, err := a.service.GetAllBoard(int64(uid))
	if err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"boards": boards,
		},
	))
}

func (a *ApiServer) GetBoardSub(c *gin.Context) {
	space := c.Query("space")
	baseDir := c.Query("baseDir")
	nowDir := c.Query("nowDir")
	subInfos, err := a.service.GetBoardSub(space, baseDir, nowDir)
	if err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"num":      len(subInfos),
			"subInfos": subInfos,
		},
	))
}

func (a *ApiServer) MakeDir(c *gin.Context) {
	baseDir := c.Query("baseDir")
	newDir := c.Query("newDir")
	space := c.Param("space")
	if err := a.service.MakeDir(space, baseDir, newDir); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"path": filepath.Join(baseDir, newDir),
		},
	))
}

func (a *ApiServer) RenameDir(c *gin.Context) {
	baseDir := c.Query("baseDir")
	oldName := c.Query("oldName")
	newName := c.Query("newName")
	space := c.Param("space")
	if err := a.service.RenameDir(space, baseDir, oldName, newName); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"path": filepath.Join(baseDir, newName),
		},
	))
}

func (a *ApiServer) DeleteDir(c *gin.Context) {
	baseDir := c.Query("baseDir")
	dirName := c.Query("dirName")
	space := c.Param("space")
	if err := a.service.DeleteDir(space, baseDir, dirName); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"path": filepath.Join(baseDir, dirName),
		},
	))
}
