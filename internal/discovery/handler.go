package discovery

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type (
	Controller struct {
		discovery DiscoveryResponse
	}
	DiscoveryResponse struct {
		ModulesV1 string `json:"modules.v1"`
	}
)

func NewController(e *echo.Echo, discovery DiscoveryResponse) *Controller {
	controller := &Controller{discovery: discovery}
	e.GET("/.well-known/terraform.json", controller.Discovery)
	return controller
}

// Discovery
func (ctrl *Controller) Discovery(c echo.Context) (err error) {
	return c.JSON(http.StatusOK, ctrl.discovery)
}
