package providersdk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceEventoptionsPolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceEventoptionsPolicyConfigCreate(),
			},
			{
				Config: testAccResourceEventoptionsPolicyConfigUpdate(),
			},
			{
				ResourceName:      "junos_eventoptions_policy.testacc_evtopts_policy",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceEventoptionsPolicyConfigUpdate2(),
			},
		},
	})
}

func testAccResourceEventoptionsPolicyConfigCreate() string {
	return `
resource "junos_eventoptions_destination" "testacc_evtopts_policy" {
  lifecycle {
    create_before_destroy = true
  }
  name = "testacc_evtopts_policy#1"
  archive_site {
    url = "https://example.com"
  }
}
resource "junos_system_login_user" "testacc_evtopts_policy" {
  lifecycle {
    create_before_destroy = true
  }
  name  = "testacc_evtopts_policy"
  class = "unauthorized"
}
resource "junos_eventoptions_policy" "testacc_evtopts_policy" {
  name   = "testacc_evtopts_policy#1"
  events = ["aaa_infra_fail", "acct_fork_err"]
  then {
    change_configuration {
      commands                         = ["test2 test1"]
      commit_options_check             = true
      commit_options_check_synchronize = true
    }
    event_script {
      filename = "filename_1"
      destination {
        name           = junos_eventoptions_destination.testacc_evtopts_policy.name
        retry_count    = 1
        retry_interval = 2
        transfer_delay = 3
      }
      output_filename = "filename_2"
      output_format   = "xml"
      user_name       = junos_system_login_user.testacc_evtopts_policy.name
    }
    execute_commands {
      commands = ["test4 test3"]
      destination {
        name           = junos_eventoptions_destination.testacc_evtopts_policy.name
        retry_count    = 1
        retry_interval = 2
        transfer_delay = 3
      }
      output_filename = "filename_3"
      output_format   = "xml"
      user_name       = junos_system_login_user.testacc_evtopts_policy.name
    }
    priority_override_facility = "external"
    priority_override_severity = "info"
    raise_trap                 = true
    upload {
      filename    = "filename_5"
      destination = junos_eventoptions_destination.testacc_evtopts_policy.name
    }
  }
  attributes_match {
    from    = "acct_fork_err.error-message"
    compare = "starts-with"
    to      = "aaa_infra_fail.test3"
  }
  within {
    time_interval = 7
    events        = ["aaa_infra_fail", "acct_fork_err"]
    not_events    = ["aaa_usage_err"]
    trigger_count = 8
    trigger_when  = "after"
  }
}
`
}

func testAccResourceEventoptionsPolicyConfigUpdate() string {
	return `
resource "junos_eventoptions_destination" "testacc_evtopts_policy" {
  lifecycle {
    create_before_destroy = true
  }
  name = "testacc_evtopts_policy#1"
  archive_site {
    url = "https://example.com"
  }
}
resource "junos_system_login_user" "testacc_evtopts_policy" {
  lifecycle {
    create_before_destroy = true
  }
  name  = "testacc_evtopts_policy"
  class = "unauthorized"
}
resource "junos_eventoptions_policy" "testacc_evtopts_policy" {
  name   = "testacc_evtopts_policy#1"
  events = ["aaa_infra_fail", "acct_fork_err"]
  then {
    change_configuration {
      commands                   = ["test1"]
      commit_options_force       = true
      commit_options_log         = "log_1"
      commit_options_synchronize = true
      retry_count                = 2
      retry_interval             = 1
      user_name                  = junos_system_login_user.testacc_evtopts_policy.name
    }
    event_script {
      filename = "filename_1"
      arguments {
        name  = "args2"
        value = "value2"
      }
      arguments {
        name  = "args1"
        value = "value1"
      }
      destination {
        name           = junos_eventoptions_destination.testacc_evtopts_policy.name
        retry_count    = 1
        retry_interval = 2
        transfer_delay = 3
      }
      output_filename = "filename_2"
      output_format   = "xml"
      user_name       = junos_system_login_user.testacc_evtopts_policy.name
    }
    execute_commands {
      commands = ["test2"]
      destination {
        name           = junos_eventoptions_destination.testacc_evtopts_policy.name
        retry_count    = 1
        retry_interval = 2
        transfer_delay = 3
      }
      output_filename = "filename_3"
      output_format   = "xml"
      user_name       = junos_system_login_user.testacc_evtopts_policy.name
    }
    priority_override_facility = "external"
    priority_override_severity = "info"
    raise_trap                 = true
    upload {
      filename    = "filename_5"
      destination = junos_eventoptions_destination.testacc_evtopts_policy.name
    }
    upload {
      filename       = "filename_4"
      destination    = junos_eventoptions_destination.testacc_evtopts_policy.name
      retry_count    = 4
      retry_interval = 5
      transfer_delay = 6
      user_name      = junos_system_login_user.testacc_evtopts_policy.name
    }
  }
  attributes_match {
    from    = "acct_fork_err.error-message"
    compare = "starts-with"
    to      = "aaa_infra_fail.test3"
  }
  attributes_match {
    from    = "aaa_infra_fail.error-message"
    compare = "equals"
    to      = "acct_fork_err.test2"
  }
  attributes_match {
    from    = "aaa_infra_fail.error-message"
    compare = "equals"
    to      = "acct_fork_err.test1"
  }
  within {
    time_interval = 7
    events        = ["aaa_infra_fail"]
    not_events    = ["aaa_usage_err"]
    trigger_count = 8
    trigger_when  = "after"
  }
  within {
    time_interval = 10
    events        = ["acct_fork_err"]
    trigger_count = 8
    trigger_when  = "after"
  }
}
`
}

func testAccResourceEventoptionsPolicyConfigUpdate2() string {
	return `
resource "junos_eventoptions_policy" "testacc_evtopts_policy" {
  name   = "testacc_evtopts_policy#1"
  events = ["aaa_infra_fail", "acct_fork_err"]
  then {
    ignore = true
  }
}
`
}
