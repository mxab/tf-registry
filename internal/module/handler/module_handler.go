package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mxab/tf-registry/internal/module/service"
	"github.com/samber/lo"
)

// //https://developer.hashicorp.com/terraform/internals/module-registry-protocol
type (
	ListRequest struct {
		Limit     int    `query:"limit" validate:"gte=0,lte=100"`
		Offset    int    `query:"offset" validate:"gte=0"`
		Provider  string `query:"provider"`
		Namespace string `param:"namespace"`
	}
	SearchRequest struct {
		Limit     int    `query:"limit" validate:"gte=0,lte=100"`
		Offset    int    `query:"offset" validate:"gte=0"`
		Query     string `query:"q" validate:"required"`
		Namespace string `param:"namespace"`
		Provider  string `query:"provider"`
	}
	ListModuleVersionsRequest struct {
		Namespace string `param:"namespace"`
		Name      string `param:"name"`
		System    string `param:"system"`
	}
	DownloadModuleRequest struct {
		Namespace string `param:"namespace"`
		Name      string `param:"name"`
		System    string `param:"system"`
		Version   string `param:"version"`
	}
	Module struct {
		Id          string `json:"id"`
		Owner       string `json:"owner"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name"`
		Version     string `json:"version"`
		Provider    string `json:"provider"`
		Description string `json:"description"`
		Source      string `json:"source"`
		PublishedAt string `json:"published_at"`
	}
	ModuleResultMeta struct {
		Limit         int  `json:"limit"`
		CurrentOffset int  `json:"current_offset"`
		NextOffset    *int `json:"next_offset,omitempty"`
		PrevOffset    *int `json:"prev_offset,omitempty"`
	}
	ModuleResult struct {
		Meta    ModuleResultMeta `json:"meta"`
		Modules []Module         `json:"modules"`
	}

	ModuleVersionsResponse struct {
		Modules []ModuleVersions `json:"modules"`
	}
	ModuleVersions struct {
		Versions []ModuleVersion `json:"versions"`
	}
	ModuleVersion struct {
		Version string `json:"version"`
	}
	Controller struct {
		ModuleService service.ModuleService
	}
)

func (ctrl *Controller) ListModules(c echo.Context) (err error) {

	listRequest := &ListRequest{
		Limit:     10,
		Offset:    0,
		Provider:  "",
		Namespace: "",
	}
	if err = c.Bind(listRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = c.Validate(listRequest); err != nil {
		return err
	}
	data, err := ctrl.ModuleService.List(service.ListParams{
		Limit:     listRequest.Limit,
		Offset:    listRequest.Offset,
		Provider:  listRequest.Provider,
		Namespace: listRequest.Namespace,
	})
	if err != nil {
		c.Logger().Warn(err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, ModuleResult{
		Meta:    convertMeta(data.Meta),
		Modules: lo.Map(data.Modules, convertModule),
	})
}

func (ctrl *Controller) SearchModules(c echo.Context) (err error) {
	searchRequest := &SearchRequest{
		Limit:     10,
		Offset:    0,
		Query:     "",
		Namespace: "",
		Provider:  "",
	}

	if err = c.Bind(searchRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = c.Validate(searchRequest); err != nil {
		return err
	}

	data, err := ctrl.ModuleService.Search(service.SearchParams{Query: searchRequest.Query,
		Limit:     searchRequest.Limit,
		Offset:    searchRequest.Offset,
		Provider:  searchRequest.Provider,
		Namespace: searchRequest.Namespace,
	})
	if err != nil {
		c.Logger().Warn(err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, ModuleResult{
		Meta:    convertMeta(data.Meta),
		Modules: lo.Map(data.Modules, convertModule),
	})

}
func convertMeta(meta service.ModuleResultMeta) ModuleResultMeta {
	resultMeta := ModuleResultMeta{
		Limit:         meta.Limit,
		CurrentOffset: meta.CurrentOffset,
	}

	if meta.NextOffset != 0 {
		resultMeta.NextOffset = &meta.NextOffset
	}
	if meta.PrevOffset != 0 {
		resultMeta.PrevOffset = &meta.PrevOffset
	}
	return resultMeta
}
func convertModule(m service.Module, _ int) Module {
	return Module{
		Id:          m.Id,
		Owner:       m.Owner,
		Namespace:   m.Namespace,
		Name:        m.Name,
		Version:     m.Version,
		Provider:    m.Provider,
		Description: m.Description,
		Source:      m.Source,
		PublishedAt: m.PublishedAt,
	}
}
func (ctrl *Controller) ListModuleVersions(c echo.Context) (err error) {

	request := new(ListModuleVersionsRequest)

	if err = c.Bind(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = c.Validate(request); err != nil {
		return err
	}
	result, err := ctrl.ModuleService.Versions(service.ModuleDescriptor{
		Namespace: request.Namespace,
		Name:      request.Name,
		System:    request.System,
	})
	if err != nil {
		c.Logger().Warn(err)
		return echo.ErrInternalServerError
	}
	return c.JSON(http.StatusOK, ModuleVersionsResponse{
		Modules: []ModuleVersions{
			{
				Versions: lo.Map(result, func(v string, _ int) ModuleVersion {
					return ModuleVersion{
						Version: v,
					}
				}),
			},
		},
	})
}

// DownloadModule
func (ctrl *Controller) DownloadModule(c echo.Context) (err error) {
	request := new(DownloadModuleRequest)

	if err = c.Bind(request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err = c.Validate(request); err != nil {
		return err
	}

	url, err := ctrl.ModuleService.DownloadUrl(service.ModuleDescriptor{
		Namespace: request.Namespace,
		Name:      request.Name,
		System:    request.System,
	}, request.Version)
	if err != nil {
		c.Logger().Warn(err)
		return echo.ErrInternalServerError
	}
	c.Response().Header().Set("X-Terraform-Get", url)
	return c.NoContent(http.StatusNoContent)
}

func RegisterModuleControllerGroup(g *echo.Group, moduleService service.ModuleService) {
	ctrl := &Controller{ModuleService: moduleService}
	g.GET("", ctrl.ListModules)
	g.GET("/search", ctrl.SearchModules)
	g.GET("/:namespace/:name/:system/:version", ctrl.ListModuleVersions)
	g.GET("/:namespace/:name/:system/:version/download", ctrl.DownloadModule)
}
