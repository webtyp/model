package model

// Industry identifies the business domain of an application so a module can adapt its copy, icon and availability.
type Industry string

const (
	IndustryHumanHealth     Industry = "human_health"
	IndustryAnimalCare      Industry = "animal_care"
	IndustryEducation       Industry = "education"
	IndustryGeneralBusiness Industry = "general_business"
)

// AnyIndustry returns a fresh slice with every supported industry. Generic modules declare Industries: model.AnyIndustry().
func AnyIndustry() []Industry {
	return []Industry{
		IndustryHumanHealth,
		IndustryAnimalCare,
		IndustryEducation,
		IndustryGeneralBusiness,
	}
}

// IndustryIcon pairs an industry with the sprite icon name a module uses for it.
type IndustryIcon struct {
	Industry Industry
	Icon     string
}

// SupportsIndustry reports whether def runs in ind. An empty ind or an empty Industries list is treated as compatible.
func SupportsIndustry(def ModuleDefinition, ind Industry) bool {
	if ind == "" || len(def.Industries) == 0 {
		return true
	}
	for _, supported := range def.Industries {
		if supported == ind {
			return true
		}
	}
	return false
}

// ModuleUse is how a role uses a module; each use translates to RBAC grants and UI behavior.
type ModuleUse uint8

const (
	ManagesEverything    ModuleUse = iota + 1 // administrator: full CRUD
	WorksWithAllRecords                       // staff/reception: create, read and update all records
	SeesOwnRecords                            // client/patient: create, read and update only own linked records
	VisitsWithoutAccount                      // visitor: read-only without an account
)

// ModuleUseOverride sets a different ModuleUse for one named module within a Role.
type ModuleUseOverride struct {
	Module string
	Use    ModuleUse
}

// Role declares an application role by code, display name, default module use and per-module overrides.
type Role struct {
	Code      RoleCode
	Name      string
	Use       ModuleUse
	Overrides []ModuleUseOverride
}

// UseFor returns the ModuleUse of r for module, checking Overrides before falling back to r.Use.
func (r Role) UseFor(module string) ModuleUse {
	for _, ov := range r.Overrides {
		if ov.Module == module {
			return ov.Use
		}
	}
	return r.Use
}
