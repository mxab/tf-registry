package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	tft "github.com/mxab/tf-registry/test"

	"github.com/kinbiko/jsonassert"
	"github.com/labstack/echo/v4"
	tfv "github.com/mxab/tf-registry/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ()

func TestListModule(t *testing.T) {
	// Setup
	ja := jsonassert.New(t)
	e := echo.New()
	e.Validator = tfv.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/v1/modules")
	controller := &Controller{
		ModuleService: tft.NewMockModuleService(),
	}

	// Assertions
	if assert.NoError(t, controller.ListModules(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		ja.Assertf(rec.Body.String(), buildExpectedModulesJson(t, map[string]any{"next_offset": "<<PRESENCE>>"}, map[string]any{}, map[string]any{}, map[string]any{}, map[string]any{}))
	}
}
func TestSearchModules(t *testing.T) {

	// Testing Table that sets query parameters
	table := []struct {
		name        string
		queryParams map[string]string

		expectedCode int
		expectedJSON string
	}{
		{
			name:         "no query",
			queryParams:  map[string]string{},
			expectedCode: http.StatusBadRequest,
			expectedJSON: "",
		},
		{
			name: "only q",
			queryParams: map[string]string{
				"q": "network",
			},
			expectedCode: http.StatusOK,
			expectedJSON: buildExpectedModulesJson(t, map[string]any{"next_offset": "<<PRESENCE>>"}, map[string]any{"name": "network"}, map[string]any{"name": "network"}),
		},
		{
			name: "with limit",
			queryParams: map[string]string{
				"q":     "network",
				"limit": "1",
			},
			expectedCode: http.StatusOK,
			expectedJSON: buildExpectedModulesJson(t, map[string]any{"next_offset": "<<PRESENCE>>"}, map[string]any{"name": "network"}),
		},
		{
			name: "with limit, offset",
			queryParams: map[string]string{
				"q":      "network",
				"limit":  "1",
				"offset": "1",
			},

			expectedCode: http.StatusOK,
			expectedJSON: buildExpectedModulesJson(t, map[string]any{"next_offset": "<<PRESENCE>>"}, map[string]any{"name": "network"}),
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {

			ja := jsonassert.New(t)
			require := require.New(t)
			assert := assert.New(t)
			e := echo.New()
			e.Validator = tfv.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/v1/modules/search")

			for key, value := range test.queryParams {
				c.QueryParams().Add(key, value)
			}

			controller := &Controller{
				ModuleService: tft.NewMockModuleService(),
			}

			err := controller.SearchModules(c)

			if test.expectedCode < 400 && assert.NoError(err) {
				require.Equal(test.expectedCode, rec.Code, fmt.Sprintf("Should have status code %d", test.expectedCode))
				body := rec.Body.String()
				ja.Assertf(body, test.expectedJSON)
			} else if test.expectedCode >= 400 && assert.Error(err) {

				require.Equal(test.expectedCode, err.(*echo.HTTPError).Code, fmt.Sprintf("Should have status code %d", test.expectedCode))
			}
		})
	}

	// Setup

}

func buildExpectedModulesJson(t *testing.T, meta map[string]any, modules ...map[string]any) string {
	var defaultMapForModuleFields = map[string]string{
		"id":           "<<PRESENCE>>",
		"owner":        "",
		"namespace":    "<<PRESENCE>>",
		"name":         "<<PRESENCE>>",
		"version":      "<<PRESENCE>>",
		"provider":     "<<PRESENCE>>",
		"description":  "<<PRESENCE>>",
		"source":       "<<PRESENCE>>",
		"published_at": "<<PRESENCE>>",
	}

	var defaultMapForMetaFields = map[string]string{
		"current_offset": "<<PRESENCE>>",
		"limit":          "<<PRESENCE>>",
	}
	// iterarte over modules and add default values, concat all to one string
	var modulesForJa []any = []any{"<<UNORDERED>>"}
	for _, module := range modules {
		for key, value := range defaultMapForModuleFields {
			if _, ok := module[key]; !ok {
				module[key] = value
			}
		}
		modulesForJa = append(modulesForJa, module)

	}
	// craete meta json, add default values
	for key, value := range defaultMapForMetaFields {
		if _, ok := meta[key]; !ok {
			meta[key] = value
		}
	}

	//join modulesJson to one string, comma separated and wrap in array
	expectedJson, err := json.Marshal(map[string]any{
		"meta":    meta,
		"modules": modulesForJa,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(expectedJson)

}

// /:namespace/:name/:provider/versions
// :namespace/:name/:system/versions
func TestListModuleVersions(t *testing.T) {
	// Setup
	ja := jsonassert.New(t)

	e := echo.New()
	e.Validator = tfv.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/v1/modules/:namespace/:name/:system/versions")
	c.SetParamNames("namespace", "name", "system")
	c.SetParamValues("Azure", "network", "azurerm")

	controller := &Controller{
		ModuleService: tft.NewMockModuleService(),
	}

	// Assertions
	if assert.NoError(t, controller.ListModuleVersions(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		json, err := json.Marshal(map[string]any{
			"modules": []any{
				map[string]any{
					"versions": []map[string]string{
						{"version": "1.1.1"},
					},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		ja.Assertf(rec.Body.String(), string(json))
	}
}
