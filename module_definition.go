package model

// ModuleDefinition is the one declaration of a domain module: its name and the entities it serves. definitionc generates
// the module's operations from it; nothing else lists the module's entities.
//
//	var Definition = model.ModuleDefinition{Name: "patient_directory", Entities: []model.Definition{PatientModel}}
type ModuleDefinition struct {
	Name     string       // the module identity, e.g. "patient_directory"
	Entities []Definition // every entity of the module, in declaration order
}
