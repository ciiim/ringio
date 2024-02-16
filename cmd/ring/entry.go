package ring

import (
	"github.com/ciiim/cloudborad/cmd/ring/router"
)

func init() {
	api := router.Router()
	api.Run(":8080")
}
