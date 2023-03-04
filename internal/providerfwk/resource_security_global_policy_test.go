package providerfwk_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJunosSecurityGlobalPolicy_basic(t *testing.T) {
	if os.Getenv("TESTACC_SRX") != "" {
		resource.Test(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccJunosSecurityGlobalPolicyConfigCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_security_global_policy.testacc_secglobpolicy",
							"policy.#", "1"),
						resource.TestCheckResourceAttr("junos_security_global_policy.testacc_secglobpolicy",
							"policy.0.name", "test"),
						resource.TestCheckResourceAttr("junos_security_global_policy.testacc_secglobpolicy",
							"policy.0.then", "permit"),
					),
				},
				{
					Config: testAccJunosSecurityGlobalPolicyConfigUpdate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_security_global_policy.testacc_secglobpolicy",
							"policy.#", "2"),
						resource.TestCheckResourceAttr("junos_security_global_policy.testacc_secglobpolicy",
							"policy.0.permit_application_services.idp", "true"),
						resource.TestCheckResourceAttr("junos_security_global_policy.testacc_secglobpolicy",
							"policy.1.then", "deny"),
					),
				},
				{
					ResourceName:      "junos_security_global_policy.testacc_secglobpolicy",
					ImportState:       true,
					ImportStateVerify: true,
				},
				{
					Config: testAccJunosSecurityGlobalPolicyConfigUpdate2(),
				},
			},
		})
	}
}

func testAccJunosSecurityGlobalPolicyConfigCreate() string {
	return `
resource "junos_security_zone" "testacc_secglobpolicy1" {
  name = "testacc_secglobpolicy1"
}
resource "junos_security_zone" "testacc_secglobpolicy2" {
  lifecycle {
    create_before_destroy = true
  }
  name = "testacc_secglobpolicy2"
}
resource "junos_security_address_book" "testacc_secglobpolicy" {
  network_address {
    name  = "blue"
    value = "192.0.2.1/32"
  }
  network_address {
    name  = "green"
    value = "192.0.2.2/32"
  }
}
resource "junos_services_user_identification_device_identity_profile" "profile" {
  lifecycle {
    create_before_destroy = true
  }
  name   = "testacc_secglobpolicy"
  domain = "testacc_secglobpolicy"
  attribute {
    name  = "device-identity"
    value = ["testacc_secglobpolicy"]
  }
}
resource "junos_security_global_policy" "testacc_secglobpolicy" {
  depends_on = [
    junos_security_address_book.testacc_secglobpolicy
  ]
  policy {
    name                               = "test"
    match_source_address               = ["blue"]
    match_destination_address          = ["green"]
    match_destination_address_excluded = true
    match_application                  = ["any"]
    match_dynamic_application          = ["any"]
    match_source_end_user_profile      = junos_services_user_identification_device_identity_profile.profile.name
    match_from_zone                    = [junos_security_zone.testacc_secglobpolicy1.name]
    match_to_zone                      = [junos_security_zone.testacc_secglobpolicy2.name]
  }
}
`
}

func testAccJunosSecurityGlobalPolicyConfigUpdate() string {
	return `
resource "junos_security_zone" "testacc_secglobpolicy1" {
  lifecycle {
    create_before_destroy = true
  }
  name = "testacc_secglobpolicy1"
}
resource "junos_security_address_book" "testacc_secglobpolicy" {
  lifecycle {
    create_before_destroy = true
  }
  network_address {
    name  = "blue"
    value = "192.0.2.1/32"
  }
  network_address {
    name  = "green"
    value = "192.0.2.2/32"
  }
}
resource "junos_services_advanced_anti_malware_policy" "testacc_secglobpolicy" {
  lifecycle {
    create_before_destroy = true
  }
  name                     = "testacc_secglobpolicy"
  verdict_threshold        = "recommended"
  default_notification_log = true
}
resource "junos_security_global_policy" "testacc_secglobpolicy" {
  depends_on = [
    junos_security_address_book.testacc_secglobpolicy
  ]
  policy {
    name                      = "test"
    match_source_address      = ["blue"]
    match_destination_address = ["any"]
    match_application         = ["any"]
    match_from_zone           = [junos_security_zone.testacc_secglobpolicy1.name]
    match_to_zone             = [junos_security_zone.testacc_secglobpolicy1.name]
    count                     = true
    log_init                  = true
    log_close                 = true
    permit_application_services {
      advanced_anti_malware_policy = junos_services_advanced_anti_malware_policy.testacc_secglobpolicy.name
      idp                          = true
      redirect_wx                  = true
      ssl_proxy {}
      uac_policy {}
    }
  }
  policy {
    name                          = "drop"
    match_source_address          = ["blue"]
    match_destination_address     = ["any"]
    match_application             = ["any"]
    match_from_zone               = ["any"]
    match_to_zone                 = ["any"]
    match_source_address_excluded = true
    then                          = "deny"
  }
}
`
}

func testAccJunosSecurityGlobalPolicyConfigUpdate2() string {
	return `
resource "junos_security_zone" "testacc_secglobpolicy1" {
  lifecycle {
    create_before_destroy = true
  }
  name = "testacc_secglobpolicy1"
}
resource "junos_security_idp_policy" "testacc_secglobpolicy" {
  lifecycle {
    create_before_destroy = true
  }
  name = "testacc_secglobpolicy"
}
resource "junos_security_address_book" "testacc_secglobpolicy" {
  lifecycle {
    create_before_destroy = true
  }
  network_address {
    name  = "blue"
    value = "192.0.2.1/32"
  }
  network_address {
    name  = "green"
    value = "192.0.2.2/32"
  }
}
resource "junos_security_global_policy" "testacc_secglobpolicy" {
  depends_on = [
    junos_security_address_book.testacc_secglobpolicy
  ]
  policy {
    name                      = "test"
    match_source_address      = ["blue"]
    match_destination_address = ["any"]
    match_application         = ["any"]
    match_from_zone           = [junos_security_zone.testacc_secglobpolicy1.name]
    match_to_zone             = [junos_security_zone.testacc_secglobpolicy1.name]
    count                     = true
    log_init                  = true
    log_close                 = true
    permit_application_services {
      idp_policy = junos_security_idp_policy.testacc_secglobpolicy.name
    }
  }
}
`
}
