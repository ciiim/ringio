package router

import (
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ciiim/cloudborad/models"
	"github.com/ciiim/cloudborad/service"
	"github.com/gin-gonic/gin"
)

const (
	Board_Success = 2000
	Board_Failed  = 2001
	Board_FormErr = 2002

	Upload_Success = 3000
	Upload_Failed  = 3001
	Upload_FormErr = 3002

	Download_Success = 3100
	Download_Failed  = 3101
	Download_FormErr = 3102
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
	makeDirJSON := struct {
		Space   string `json:"space"`
		BaseDir string `json:"baseDir"`
		NewDir  string `json:"newDir"`
	}{}
	if err := c.ShouldBindJSON(&makeDirJSON); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Board_FormErr, false, err.Error(), nil))
		return
	}
	if err := a.service.MakeDir(makeDirJSON.Space, makeDirJSON.BaseDir, makeDirJSON.NewDir); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"path": filepath.Join(makeDirJSON.BaseDir, makeDirJSON.NewDir),
		},
	))
}

func (a *ApiServer) RenameDir(c *gin.Context) {
	renameDirJSON := struct {
		Space   string `json:"space"`
		BaseDir string `json:"baseDir"`
		OldName string `json:"oldName"`
		NewName string `json:"newName"`
	}{}
	if err := a.service.RenameDir(renameDirJSON.Space, renameDirJSON.BaseDir, renameDirJSON.OldName, renameDirJSON.NewName); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"path": filepath.Join(renameDirJSON.BaseDir, renameDirJSON.NewName),
		},
	))
}

func (a *ApiServer) DeleteDir(c *gin.Context) {
	deleteDirJSON := struct {
		Space   string `json:"space"`
		BaseDir string `json:"baseDir"`
		DirName string `json:"dirName"`
	}{}
	if err := a.service.DeleteDir(deleteDirJSON.Space, deleteDirJSON.BaseDir, deleteDirJSON.DirName); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Board_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(
		Board_Success,
		true,
		"success",
		gin.H{
			"path": filepath.Join(deleteDirJSON.BaseDir, deleteDirJSON.DirName),
		},
	))
}

/*
下载前准备
获取storeID用于识别分块文件
*/
func (a *ApiServer) PreUploadFile(c *gin.Context) {
	log.Println("pre upload file")
	preUploadForm := &service.PreUploadForm{}
	if err := c.ShouldBindJSON(preUploadForm); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Upload_FormErr, false, err.Error(), nil))
		return
	}
	log.Println(preUploadForm)
	storeID, err := a.service.PreUploadFile(preUploadForm)
	if err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Upload_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(Upload_Success, true, "success", gin.H{
		"storeID": storeID,
	}))
}

func (a *ApiServer) UploadFile(c *gin.Context) {
	log.Println("upload file")
	uploadForm := &service.UploadForm{}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Upload_FormErr, false, err.Error(), nil))
	}
	uploadForm.Space = c.PostForm("space")
	uploadForm.BaseDir = c.PostForm("base")
	uploadForm.StoreID = c.PostForm("identifier")
	uploadForm.ChunkHash = c.PostForm("chunkHash")

	uploadForm.File = file
	uploadForm.ChunkNumber, _ = strconv.Atoi(c.PostForm("chunkNumber"))
	uploadForm.ChunkSize, _ = strconv.ParseInt(c.PostForm("chunkSize"), 10, 64)
	uploadForm.CurrentChunkSize, _ = strconv.ParseInt(c.PostForm("currentChunkSize"), 10, 64)

	uploadForm.TotalSize, _ = strconv.ParseInt(c.PostForm("totalSize"), 10, 64)

	uploadForm.Filename = c.PostForm("filename")
	uploadForm.RelativeDir = c.PostForm("relativePath")
	uploadForm.TotalChunks, _ = strconv.Atoi(c.PostForm("totalChunks"))
	status, err := a.service.UploadFile(uploadForm)
	if err != nil || status == -1 {
		c.JSON(http.StatusOK, JSON_RETURN(Upload_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(Upload_Success, true, "success", nil))
}

func (a *ApiServer) PreDownloadFile(c *gin.Context) {
	log.Println("pre download file")
	preDownloadForm := struct {
		Space    string `form:"space"`
		BaseDir  string `form:"base"`
		Filename string `form:"fileName"`
	}{}
	if err := c.ShouldBindQuery(&preDownloadForm); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Download_FormErr, false, err.Error(), nil))
		return
	}
	log.Println(preDownloadForm)
	id, size, num, err := a.service.PreDownloadFile(preDownloadForm.Space, preDownloadForm.BaseDir, preDownloadForm.Filename)
	log.Printf("[PreDownload]id:%s, size:%d, num:%d", id, size, num)
	if err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Download_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(Download_Success, true, "success", gin.H{
		"downloadID": id,
		"chunkNum":   num,
		"fileSize":   size,
	}))
}

func (a *ApiServer) DownloadChunk(c *gin.Context) {
	log.Println("download chunk")
	downloadForm := struct {
		DownloadID string `form:"downloadID"`
		ChunkIndex int    `form:"chunkIndex"`
	}{}

	if err := c.ShouldBindQuery(&downloadForm); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Download_FormErr, false, err.Error(), nil))
		return
	}
	data, err := a.service.DownloadChunk(downloadForm.DownloadID, downloadForm.ChunkIndex)
	if err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Download_Failed, false, err.Error(), nil))
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// use range header
func (a *ApiServer) DownloadChunkRange(c *gin.Context) {
	log.Println("download chunk use range")
	downloadID := c.Param("downloadID")
	rangeHeader := c.GetHeader("Range")
	chunks, start, end, total, err := a.service.DownloadChunks(downloadID, rangeHeader)
	if err != nil {
		c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
		c.Header("Accept-Ranges", "bytes")
		c.Header("Content-Type", "application/octet-stream")
		c.JSON(http.StatusOK, JSON_RETURN(Download_Failed, false, err.Error(), nil))
		return
	}
	c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(total, 10))
	c.Data(http.StatusOK, "application/octet-stream", chunks)
}

func (a *ApiServer) DownloadFileDone(c *gin.Context) {
	log.Println("upload file done")
	uploadDoneForm := struct {
		DownloadID string `json:"downloadID"`
	}{}
	if err := c.ShouldBindJSON(&uploadDoneForm); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Download_FormErr, false, err.Error(), nil))
		return
	}
	if err := a.service.DownloadDone(uploadDoneForm.DownloadID); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Download_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(Download_Success, true, "success", nil))
}

func (a *ApiServer) DeleteFile(c *gin.Context) {
	deleteFileForm := struct {
		Space    string `json:"space"`
		BaseDir  string `json:"base"`
		Filename string `json:"fileName"`
	}{}
	if err := c.ShouldBindJSON(&deleteFileForm); err != nil {
		c.JSON(http.StatusBadRequest, JSON_RETURN(Download_FormErr, false, err.Error(), nil))
		return
	}
	if err := a.service.DeleteFile(deleteFileForm.Space, deleteFileForm.BaseDir, deleteFileForm.Filename); err != nil {
		c.JSON(http.StatusOK, JSON_RETURN(Download_Failed, false, err.Error(), nil))
		return
	}
	c.JSON(http.StatusOK, JSON_RETURN(Download_Success, true, "success", nil))
}
