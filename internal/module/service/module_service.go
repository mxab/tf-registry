package service

// go interface called ModuleService, has a search method that takes a query string, limit int, offset int, provider string, namespace string, verified bool and returns a ModuleResult, and a Get method that takes an id string and returns a Module

type (
	Module struct {
		Id          string
		Owner       string
		Namespace   string
		Name        string
		Version     string
		Provider    string
		Description string
		Source      string
		PublishedAt string
	}
	ModuleResultMeta struct {
		Limit         int
		CurrentOffset int
		NextOffset    int
		PrevOffset    int
	}
	ModuleResult struct {
		Meta    ModuleResultMeta
		Modules []Module
	}
	ListParams struct {
		Limit     int
		Offset    int
		Provider  string
		Namespace string
	}
	SearchParams struct {
		Query     string
		Limit     int
		Offset    int
		Provider  string
		Namespace string
	}

	ModuleDescriptor struct {
		Namespace string
		Name      string
		System    string
	}
)
type ModuleService interface {
	Search(params SearchParams) (ModuleResult, error)
	List(params ListParams) (ModuleResult, error)
	Versions(modul ModuleDescriptor) ([]string, error)
}
