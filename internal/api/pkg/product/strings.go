package product

import (
	"dev.khulnasoft.com/admin-apis/pkg/licenseapi"
)

// LoginCmd returns the login command for the product
func LoginCmd() string {
	loginCmd := "loft login"

	switch Name() {
	case licenseapi.DevSpacePro:
		return "devspace login"
	case licenseapi.VClusterPro:
		return "vcluster login"
	case licenseapi.Loft:
	}

	return loginCmd
}

// StartCmd returns the start command for the product
func StartCmd() string {
	loginCmd := "loft start"

	switch Name() {
	case licenseapi.DevSpacePro:
		loginCmd = "devspace pro start"
	case licenseapi.VClusterPro:
		loginCmd = "vcluster platform start"
	case licenseapi.Loft:
	}

	return loginCmd
}

// Url returns the url command for the product
func Url() string {
	loginCmd := "loft-url"

	switch Name() {
	case licenseapi.DevSpacePro:
		loginCmd = "devspace-pro-url"
	case licenseapi.VClusterPro:
		loginCmd = "vcluster-pro-url"
	case licenseapi.Loft:
	}

	return loginCmd
}

// ResetPassword returns the reset password command for the product
func ResetPassword() string {
	resetPassword := "loft reset password"

	switch Name() {
	case licenseapi.DevSpacePro:
		return "devspace pro reset password"
	case licenseapi.VClusterPro:
		return "vcluster platform reset password"
	case licenseapi.Loft:
	}

	return resetPassword
}
