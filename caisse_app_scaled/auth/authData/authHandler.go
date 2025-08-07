package authData

import (
	"caisse-app-scaled/caisse_app_scaled/models"
	"slices"
)

func LoginAuth(employe string, role string, employeList []models.Employe) bool {
	available := false
	switch role {
	case "commis":
		available = IsCommis(employe, employeList)
	case "manager":
		available = IsManager(employe, employeList)
	default:
		available = IsUsernameValid(employe, employeList)
	}
	return available
}

func IsUsernameValid(nom string, employeList []models.Employe) bool {
	return slices.ContainsFunc(employeList, func(e models.Employe) bool {
		return e.Nom == nom
	})
}

func IsManager(nom string, employeList []models.Employe) bool {
	return slices.ContainsFunc(employeList, func(e models.Employe) bool {
		return e.Nom == nom && e.Role == "manager"
	})
}

func IsCommis(nom string, employeList []models.Employe) bool {
	return slices.ContainsFunc(employeList, func(e models.Employe) bool {
		return e.Nom == nom && e.Role != "manager"
	})
}

func MakeEmployees(nom string, role string) models.Employe {
	return models.Employe{
		Nom:  nom,
		Role: role,
	}
}
