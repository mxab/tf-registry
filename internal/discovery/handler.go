package discovery

import (
	"net/http"

	"github.com/labstack/echo"
)

type (
	Controller struct {
		discovery DiscoveryResponse
	}
	DiscoveryResponse struct {
		ModulesV1 string `json:"modules.v1"`
	}
)

func NewController(discovery DiscoveryResponse) *Controller {
	return &Controller{discovery: discovery}
}

// Discovery
func (ctrl *Controller) Discovery(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, ctrl.discovery)
}
