package test

import (
	"github.com/mxab/tf-registry/internal/module/service"
	"github.com/samber/lo"
)

type MockModuleService struct {
	modules []service.Module
}

func NewMockModuleService() *MockModuleService {
	return &MockModuleService{
		modules: []service.Module{

			{
				Id:          "GoogleCloudPlatform/lb-http/google/1.0.4",
				Owner:       "",
				Namespace:   "GoogleCloudPlatform",
				Name:        "lb-http",
				Version:     "1.0.4",
				Provider:    "google",
				Description: "Modular Global HTTP Load Balancer for GCE using forwarding rules.",
				Source:      "https://github.com/GoogleCloudPlatform/terraform-google-lb-http",
				PublishedAt: "2017-10-17T01:22:17.792066Z",
			},
			{
				Id:          "terraform-aws-modules/vpc/aws/1.5.1",
				Owner:       "",
				Namespace:   "terraform-aws-modules",
				Name:        "vpc",
				Version:     "1.5.1",
				Provider:    "aws",
				Description: "Terraform module which creates VPC resources on AWS",
				Source:      "https://github.com/terraform-aws-modules/terraform-aws-vpc",
				PublishedAt: "2017-11-23T10:48:09.400166Z",
			},
			{
				Id:          "zoitech/network/aws/0.0.3",
				Owner:       "",
				Namespace:   "zoitech",
				Name:        "network",
				Version:     "0.0.3",
				Provider:    "aws",
				Description: "This module is intended to be used for configuring an AWS network.",
				Source:      "https://github.com/zoitech/terraform-aws-network",
				PublishedAt: "2017-11-23T15:12:06.620059Z",
			},
			{
				Id:          "Azure/network/azurerm/1.1.1",
				Owner:       "",
				Namespace:   "Azure",
				Name:        "network",
				Version:     "1.1.1",
				Provider:    "azurerm",
				Description: "Terraform Azure RM Module for Network",
				Source:      "https://github.com/Azure/terraform-azurerm-network",
				PublishedAt: "2017-11-22T17:15:34.325436Z",
			},
		},
	}

}

// implment ModuleService
// list, takes a ListParams and returns a ModuleResult
func (m *MockModuleService) List(params service.ListParams) (service.ModuleResult, error) {

	//set limit to 10 if smaller 0 or bigger than 10
	limit := lo.Clamp(params.Limit, 0, 10)
	offset := lo.Clamp(params.Offset, 0, len(m.modules))

	filteredModules := lo.Filter(m.modules, func(module service.Module, index int) bool {
		return (params.Provider != "" && module.Provider == params.Provider) ||
			(params.Namespace != "" && module.Namespace == params.Namespace) ||
			(params.Namespace == "" && params.Provider == "")
	})
	filteredModules = lo.Subset(filteredModules, offset, uint(limit))

	nextOffset := lo.Clamp(offset+limit, 0, len(m.modules))
	prevOffset := lo.Clamp(offset-limit, 0, len(m.modules))

	return service.ModuleResult{
		Meta: service.ModuleResultMeta{
			Limit:         limit,
			CurrentOffset: offset,
			NextOffset:    nextOffset,
			PrevOffset:    prevOffset,
		},
		Modules: filteredModules,
	}, nil
}

// search
func (m *MockModuleService) Search(params service.SearchParams) (service.ModuleResult, error) {

	//set limit to 10 if smaller 0 or bigger than 10
	limit := lo.Clamp(params.Limit, 0, 10)
	offset := lo.Clamp(params.Offset, 0, len(m.modules))

	filteredModules := lo.Filter(m.modules, func(module service.Module, index int) bool {
		return module.Id == params.Query ||
			module.Owner == params.Query ||
			module.Namespace == params.Query ||
			module.Name == params.Query ||
			module.Version == params.Query ||
			module.Provider == params.Query ||
			module.Description == params.Query ||
			module.Source == params.Query ||

			(params.Provider != "" && module.Provider == params.Provider) ||

			(params.Namespace != "" && module.Namespace == params.Namespace)

	})
	filteredModules = lo.Subset(filteredModules, offset, uint(limit))

	nextOffset := lo.Clamp(offset+limit, 0, len(m.modules))
	return service.ModuleResult{
		Meta: service.ModuleResultMeta{
			Limit:         limit,
			CurrentOffset: offset,
			NextOffset:    nextOffset,
		},
		Modules: filteredModules,
	}, nil
}

// Versions
func (m *MockModuleService) Versions(params service.ModuleDescriptor) ([]string, error) {

	modules := lo.Filter(m.modules, func(module service.Module, index int) bool {
		return module.Namespace == params.Namespace &&
			module.Name == params.Name &&
			module.Provider == params.System
	})

	return lo.Map(modules, func(module service.Module, _ int) string {
		return module.Version
	}), nil

}
