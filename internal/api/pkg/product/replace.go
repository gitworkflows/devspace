package product

import (
	"fmt"
	"strings"

	"dev.khulnasoft.com/admin-apis/pkg/licenseapi"
)

// Replace replaces the product name in the given usage string
// based on the current product.Product().
//
// It replaces "loft" with the specific product name:
//   - "devspace pro" for product.DevSpacePro
//   - "vcluster platform" for product.VClusterPro
//   - No replacement for product.Loft
//
// This handles case insensitive replaces like "loft" -> "devspace pro".
//
// It also handles case sensitive replaces:
//   - "Loft" -> "DevSpace.Pro" for product.DevSpacePro
//   - "Loft" -> "vCluster Platform" for product.VClusterPro
//
// This allows customizing command usage text for different products.
//
// Parameters:
//   - content: The string to update
//
// Returns:
//   - The updated string with product name replaced if needed.
func Replace(content string) string {
	switch Name() {
	case licenseapi.DevSpacePro:
		content = strings.Replace(content, "loft.sh", "devspace.pro", -1)
		content = strings.Replace(content, "loft.host", "devspace.host", -1)

		content = strings.Replace(content, "loft", "devspace pro", -1)
		content = strings.Replace(content, "Loft", "DevSpace.Pro", -1)
	case licenseapi.VClusterPro:
		content = strings.Replace(content, "loft.sh", "vcluster.pro", -1)
		content = strings.Replace(content, "loft.host", "vcluster.host", -1)

		content = strings.Replace(content, "loft", "vcluster platform", -1)
		content = strings.Replace(content, "Loft", "vCluster Platform", -1)
	case licenseapi.Loft:
	}

	return content
}

// ReplaceWithHeader replaces the "loft" product name in the given
// usage string with the specific product name based on product.Product().
// It also adds a header with padding around the product name and usage.
//
// The product name replacements are:
//
//   - "devspace pro" for product.DevSpacePro
//   - "vcluster platform" for product.VClusterPro
//   - No replacement for product.Loft
//
// This handles case insensitive replaces like "loft" -> "devspace pro".
//
// It also handles case sensitive replaces:
//   - "Loft" -> "DevSpace.Pro" for product.DevSpacePro
//   - "Loft" -> "vCluster Platform" for product.VClusterPro
//
// Parameters:
//   - use: The usage string
//   - content: The content string to run product name replacement on
//
// Returns:
//   - The content string with product name replaced and header added
func ReplaceWithHeader(use, content string) string {
	maxChar := 56

	productName := licenseapi.Loft

	switch Name() {
	case licenseapi.DevSpacePro:
		productName = "devspace pro"
	case licenseapi.VClusterPro:
		productName = "vcluster platform"
	case licenseapi.Loft:
	}

	paddingSize := (maxChar - 2 - len(productName) - len(use)) / 2

	separator := strings.Repeat("#", paddingSize*2+len(productName)+len(use)+2+1)
	padding := strings.Repeat("#", paddingSize)

	return fmt.Sprintf(`%s
%s %s %s %s
%s
%s
`, separator, padding, productName, use, padding, separator, Replace(content))
}
