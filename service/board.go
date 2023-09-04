package service

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ciiim/cloudborad/internal/fs"
	"github.com/ciiim/cloudborad/models"
)

const (
	DefaultBoardCapacity = fs.GB
)

func generateSpaceKey(uid int64, name string) string {
	sum := sha1.Sum([]byte(fmt.Sprintf("%d%s%d", uid, name, time.Now().UnixMilli())))
	return hex.EncodeToString(sum[:])
}

// 将前端传入的Board转换为数据库中的Board
func (s *Service) ToBoard(uid int64, name, passwd string, visible, pinned int) *models.Board {
	space := generateSpaceKey(uid, name)
	hasPasswd := 0
	if passwd != "" {
		passwd = encryptPasswd(passwd)
		hasPasswd = 1
	}
	return &models.Board{
		Space:        space,
		Owner:        uid,
		Name:         name,
		Passwd:       passwd,
		PasswdStatus: hasPasswd,
		Visible:      visible,
		Pinned:       pinned,
	}
}

func (s *Service) NewBoard(newBoard *models.Board) error {
	if newBoard == nil {
		return fmt.Errorf("board is nil")
	}
	newBoard.Capacity = DefaultBoardCapacity
	err := newBoard.Insert()
	if err != nil {
		return err
	}
	return s.fileServer.NewBoard(newBoard.Space, newBoard.Capacity)
}

func (s *Service) UpdateBoard(board *models.Board) error {
	if board == nil {
		return fmt.Errorf("board is nil")
	}
	return board.Update()
}

func (s *Service) DeleteBoard(boardBasic *models.Board_basic) error {
	if boardBasic == nil {
		return fmt.Errorf("board is nil")
	}
	board, err := models.QueryBoardByBasic(boardBasic)
	if err != nil {
		return err
	}
	return board.Delete()
}

func (s *Service) GetAllBoard(uid int64) ([]models.Board_basic, error) {
	return models.QueryAllBoardBasic(uid)
}

func (s *Service) GetBoardByBasic(basic *models.Board_basic) (*models.Board, error) {
	if basic == nil {
		return nil, fmt.Errorf("basic is nil")
	}
	return models.QueryBoardByBasic(basic)
}

/*
访问Board内文件夹
*/
func (s *Service) GetBoardSub(space string, baseDir string, nowDir string) ([]fs.SubInfo, error) {
	if space == "" {
		return nil, fmt.Errorf("space is nil")
	}
	return s.fileServer.GetDirSub(space, baseDir, nowDir)
}

func (s *Service) MakeDir(space string, baseDir string, dirName string) error {
	if space == "" {
		return fmt.Errorf("space is nil")
	}
	return s.fileServer.MakeDir(space, baseDir, dirName)
}

func (s *Service) RenameDir(space string, baseDir string, oldName string, newName string) error {
	if space == "" {
		return fmt.Errorf("space is nil")
	}
	return s.fileServer.RenameDir(space, baseDir, oldName, newName)
}

func (s *Service) DeleteDir(space string, baseDir string, dirName string) error {
	if space == "" {
		return fmt.Errorf("space is nil")
	}
	return s.fileServer.DeleteDir(space, baseDir, dirName)
}
