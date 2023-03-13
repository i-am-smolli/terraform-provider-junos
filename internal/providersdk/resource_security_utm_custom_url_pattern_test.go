package providersdk_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJunosSecurityUtmCustomURLPattern_basic(t *testing.T) {
	if os.Getenv("TESTACC_SRX") != "" {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccJunosSecurityUtmCustomURLPatternConfigCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_security_utm_custom_url_pattern.testacc_UrlPattern",
							"value.#", "1"),
						resource.TestCheckResourceAttr("junos_security_utm_custom_url_pattern.testacc_UrlPattern",
							"value.0", "*.google.com"),
					),
				},
				{
					Config: testAccJunosSecurityUtmCustomURLPatternConfigUpdate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_security_utm_custom_url_pattern.testacc_UrlPattern",
							"value.#", "2"),
						resource.TestCheckResourceAttr("junos_security_utm_custom_url_pattern.testacc_UrlPattern",
							"value.0", "*.google.com"),
						resource.TestCheckResourceAttr("junos_security_utm_custom_url_pattern.testacc_UrlPattern",
							"value.1", "*.google.fr"),
					),
				},
				{
					ResourceName:      "junos_security_utm_custom_url_pattern.testacc_UrlPattern",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	}
}

func testAccJunosSecurityUtmCustomURLPatternConfigCreate() string {
	return `
resource "junos_security_utm_custom_url_pattern" "testacc_UrlPattern" {
  name  = "testacc_UrlPattern"
  value = ["*.google.com"]
}
`
}

func testAccJunosSecurityUtmCustomURLPatternConfigUpdate() string {
	return `
resource "junos_security_utm_custom_url_pattern" "testacc_UrlPattern" {
  name  = "testacc_UrlPattern"
  value = ["*.google.com", "*.google.fr"]
}
`
}
