package router

import "github.com/gin-gonic/gin"

const (
	JoinCluster_Succeed = 9000
	JoinCluster_Failed  = 9001

	QuitCluster_Succeed = 9002
	QuitCluster_Failed  = 9003
)

func (a *ApiServer) JoinCluster(c *gin.Context) {
	name := c.Query("name")
	addr := c.Query("addr")
	err := a.service.JoinCluster(name, addr)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":     err.Error(),
			"success": false,
			"code":    JoinCluster_Failed,
			"data":    nil,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":     "加入集群成功",
		"success": true,
		"code":    JoinCluster_Succeed,
		"data":    nil,
	})
}

func (a *ApiServer) GetCluster(c *gin.Context) {
	//cluster := a.service.GetCluster()
	c.JSON(200, gin.H{
		"msg":     "testing",
		"success": true,
		"code":    JoinCluster_Succeed,
		"data":    nil,
	})
}

func (a *ApiServer) QuitCluster(c *gin.Context) {
	//err := a.service.QuitCluster()
	c.JSON(200, JSON_RETURN(QuitCluster_Succeed, true, "test", nil))

}
