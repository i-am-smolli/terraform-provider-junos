package providersdk_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// export TESTACC_INTERFACE=<inteface> for choose interface available else it's ge-0/0/3.
func TestAccDataSourceInterface_basic(t *testing.T) {
	if os.Getenv("TESTACC_DEPRECATED") != "" {
		testaccInterface := defaultInterfaceTestAcc
		if iface := os.Getenv("TESTACC_INTERFACE"); iface != "" {
			testaccInterface = iface
		}
		if os.Getenv("TESTACC_SWITCH") == "" {
			resource.Test(t, resource.TestCase{
				PreCheck:  func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: testAccDataSourceInterfaceConfigCreate(testaccInterface),
					},
					{
						Config: testAccDataSourceInterfaceConfigData(testaccInterface),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("data.junos_interface.testacc_datainterface",
								"id", testaccInterface+".100"),
							resource.TestCheckResourceAttr("data.junos_interface.testacc_datainterface",
								"name", testaccInterface+".100"),
							resource.TestCheckResourceAttr("data.junos_interface.testacc_datainterface",
								"inet_address.#", "1"),
							resource.TestCheckResourceAttr("data.junos_interface.testacc_datainterface",
								"inet_address.0.address", "192.0.2.1/25"),
							resource.TestCheckResourceAttr("data.junos_interface.testacc_datainterface2",
								"id", testaccInterface+".100"),
						),
					},
				},
				PreventPostDestroyRefresh: true,
			})
		}
	}
}

func testAccDataSourceInterfaceConfigCreate(interFace string) string {
	return fmt.Sprintf(`
resource "junos_interface" "testacc_datainterfaceP" {
  name         = "%s"
  description  = "testacc_datainterfaceP"
  vlan_tagging = true
}
resource "junos_interface" "testacc_datainterface" {
  name        = "${junos_interface.testacc_datainterfaceP.name}.100"
  description = "testacc_datainterface"
  inet_address {
    address = "192.0.2.1/25"
  }
}
`, interFace)
}

func testAccDataSourceInterfaceConfigData(interFace string) string {
	return fmt.Sprintf(`
resource "junos_interface" "testacc_datainterfaceP" {
  name         = "%s"
  description  = "testacc_datainterfaceP"
  vlan_tagging = true
}
resource "junos_interface" "testacc_datainterface" {
  name        = "${junos_interface.testacc_datainterfaceP.name}.100"
  description = "testacc_datainterface"
  inet_address {
    address = "192.0.2.1/25"
  }
}

data "junos_interface" "testacc_datainterface" {
  config_interface = "%s"
  match            = "192.0.2.1/"
}

data "junos_interface" "testacc_datainterface2" {
  match = "192.0.2.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)"
}
`, interFace, interFace)
}
