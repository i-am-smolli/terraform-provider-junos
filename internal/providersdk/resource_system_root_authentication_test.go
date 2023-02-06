package providersdk_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJunosSystemRootAuthentication_basic(t *testing.T) {
	if os.Getenv("TESTACC_SWITCH") == "" {
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccJunosSystemRootAuthenticationCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_system_root_authentication.root_auth",
							"encrypted_password", "$6$XXXX"),
						resource.TestCheckResourceAttr("junos_system_root_authentication.root_auth",
							"ssh_public_keys.#", "1"),
					),
				},
				{
					ResourceName:      "junos_system_root_authentication.root_auth",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccJunosSystemRootAuthenticationUpdate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_system_root_authentication.root_auth",
							"ssh_public_keys.#", "0"),
					),
				},
				{
					Config: testAccJunosSystemRootAuthenticationUpdate2(),
				},
			},
		})
	}
}

func testAccJunosSystemRootAuthenticationCreate() string {
	return `
resource "junos_system_root_authentication" "root_auth" {
  encrypted_password = "$6$XXXX"
  ssh_public_keys = [
    "ssh-rsa XXXX",
  ]
}
`
}

func testAccJunosSystemRootAuthenticationUpdate() string {
	return `
resource "junos_system_root_authentication" "root_auth" {
  encrypted_password = "$6$XXX"
  no_public_keys     = true
}
`
}

func testAccJunosSystemRootAuthenticationUpdate2() string {
	return `
resource "junos_system_root_authentication" "root_auth" {
  plain_text_password = "testPassword1234"
}
`
}
