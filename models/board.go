package models

import (
	"fmt"

	"github.com/ciiim/cloudborad/internal/database"
)

type Board struct {
	Space        string `json:"space" db:"space"`
	Name         string `json:"space_nickname" db:"space_nickname"`
	Owner        int64  `json:"owner_uid" db:"owner_uid"`
	Passwd       string `json:"passwd" db:"passwd"`
	PasswdStatus int    `json:"passwd_status" db:"passwd_status"` // 0: no passwd, 1: passwd
	Capacity     int64  `json:"capacity" db:"capacity"`
	Occupied     int64  `json:"occupied" db:"occupied"`
	Visible      int    `json:"visible" db:"visible"` // 0: private, 1: public
	Pinned       int    `json:"pinned" db:"pinned"`   // 0: not pinned, 1: pinned
}

type Board_basic struct {
	Space  string `json:"space" db:"space"`
	Owner  int64  `json:"owner_uid" db:"owner_uid"`
	Name   string `json:"space_nickname" db:"space_nickname"`
	Pinned int    `json:"pinned" db:"pinned"` // 0: not pinned, 1: pinned
}

func NewBoard(space, name string, ownerUID int64, passwd string, capacity int64, visible int, pinned int) *Board {
	hasPasswd := 0
	if passwd != "" {
		hasPasswd = 1
	}
	return &Board{
		Space:        space,
		Name:         name,
		Owner:        ownerUID,
		Passwd:       passwd,
		PasswdStatus: hasPasswd,
		Capacity:     capacity,
		Occupied:     0,
		Visible:      visible,
		Pinned:       pinned,
	}

}

func (b *Board) Insert() error {
	basic := &Board_basic{
		Space:  b.Space,
		Owner:  b.Owner,
		Name:   b.Name,
		Pinned: b.Pinned,
	}
	err := insertBoardBasic(basic)
	if err != nil {
		return err
	}

	statment := database.InsertStringMysql(b)
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(b.Space, b.Name, b.Owner, b.Passwd, b.PasswdStatus, b.Capacity, b.Occupied, b.Visible, b.Pinned)
	return err
}

func (b *Board) Update() error {
	basic := &Board_basic{
		Space:  b.Space,
		Owner:  b.Owner,
		Name:   b.Name,
		Pinned: b.Pinned,
	}
	err := updateBoardBasic(basic)
	if err != nil {
		return err
	}
	statment := database.UpdateStringMysql(b, "space = ? and owner_uid = ?")
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(b.Space, b.Name, b.Owner, b.Passwd, b.PasswdStatus, b.Capacity, b.Occupied, b.Visible, b.Pinned, b.Space, b.Owner)
	return err
}

func (b *Board) Delete() error {
	err := deleteBoardBasic(b.Space, b.Owner)
	if err != nil {
		return err
	}
	statment := database.DeleteStringMysql("board", "space = ?  and owner_uid = ?")
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(b.Space, b.Owner)
	return err
}

func insertBoardBasic(newBoard *Board_basic) error {
	statment := database.InsertStringMysql(newBoard)
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(newBoard.Space, newBoard.Owner, newBoard.Name, newBoard.Pinned)
	return err
}

func updateBoardBasic(newBoard *Board_basic) error {
	statment := database.UpdateStringMysql(newBoard, "space = ? and owner_uid = ?")
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(newBoard.Space, newBoard.Name, newBoard.Owner, newBoard.Pinned, newBoard.Space, newBoard.Owner)
	return err
}

func deleteBoardBasic(space string, ownerUID int64) error {
	statment := database.DeleteStringMysql("board_basic", "space = ?  and owner_uid = ?")
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(space, ownerUID)
	return err
}

/*
返回Board_basic数组
*/
func QueryAllBoardBasic(uid int64) ([]Board_basic, error) {
	var boards []Board_basic
	statment := database.SelectStringMysql(&Board_basic{}, fmt.Sprintf("owner_uid = %d", uid))
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var board Board_basic
		err = rows.Scan(&board.Space, &board.Owner, &board.Name, &board.Pinned)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, nil
}

func QueryBoardBasic(uid int64, space string) (*Board_basic, error) {
	var board Board_basic
	statment := database.SelectStringMysql(&Board_basic{}, fmt.Sprintf("owner_uid = %d and space = '%s'", uid, space))
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRow().Scan(&board.Space, &board.Name, &board.Owner, &board.Pinned)
	return &board, err
}

func QueryBoardByBasic(basic *Board_basic) (*Board, error) {
	var board Board
	statment := database.SelectStringMysql(&Board{}, fmt.Sprintf("owner_uid = %d and space = '%s'", basic.Owner, basic.Space))
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRow().Scan(&board.Space, &board.Name, &board.Owner, &board.Passwd, &board.PasswdStatus, &board.Capacity, &board.Occupied, &board.Visible, &board.Pinned)
	return &board, err
}

func QueryBoard(uid int64, space string) (*Board, error) {
	var board Board
	statment := database.SelectStringMysql(&Board{}, fmt.Sprintf("owner_uid = %d and space = '%s'", uid, space))
	stmt, err := database.MysqlDB.Prepare(statment)
	if err != nil {
		return nil, err
	}
	err = stmt.QueryRow().Scan(&board.Space, &board.Name, &board.Owner, &board.Capacity, &board.Occupied, &board.Visible, &board.Pinned)
	return &board, err
}
