package service

// Product identity is configuration, not permission to use a frozen profile.
type ManagementTraitsProductClass string

const (
	ManagementTraitsLegacy002 ManagementTraitsProductClass = "LEGACY002"
	ManagementTraitsNew005    ManagementTraitsProductClass = "NEW005"
	ManagementTraitsOther     ManagementTraitsProductClass = "OTHER"
)

func ClassifyManagementTraitsProduct(code string) ManagementTraitsProductClass {
	switch code {
	case "00201", "00202":
		return ManagementTraitsLegacy002
	case "00501", "00502":
		return ManagementTraitsNew005
	default:
		return ManagementTraitsOther
	}
}

func IsManagementTraitsProduct(code string) bool {
	return ClassifyManagementTraitsProduct(code) != ManagementTraitsOther
}
