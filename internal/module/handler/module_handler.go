package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mxab/tf-registry/internal/module/service"
	"github.com/samber/lo"
)

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
