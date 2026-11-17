package model

// ModuleDefinition is the one declaration of a domain module: its name, entities, supported industries, icon and uses.
// definitionc generates the module's operations from it; nothing else lists the module's entities.
//
//	var Definition = model.ModuleDefinition{Name: "client_directory", Entities: []model.Definition{ClientModel}}
type ModuleDefinition struct {
	Name       string         // the module identity, e.g. "client_directory"
	Entities   []Definition   // every entity of the module, in declaration order
	Industries []Industry     // supported industries (model.AnyIndustry() for generic modules)
	Icon       string         // default sprite icon name
	IconFor    []IndustryIcon // optional icon overrides by industry
	Uses       []ModuleUse    // supported module uses (1-4)
}

// IconForIndustry returns the icon name for ind, falling back to the default Icon.
func (d ModuleDefinition) IconForIndustry(ind Industry) string {
	for _, ov := range d.IconFor {
		if ov.Industry == ind && ov.Icon != "" {
			return ov.Icon
		}
	}
	return d.Icon
}
