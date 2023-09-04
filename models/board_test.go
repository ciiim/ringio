package models

import (
	"testing"

	"github.com/ciiim/cloudborad/internal/database"
)

func TestQueryAllBoard(t *testing.T) {
	database.InitMysql("ciiim:@Cwqqq222@tcp(124.70.14.215:3506)/cloudboard?charset=utf8")
	boards, err := QueryAllBoardBasic(4)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(boards)
}
