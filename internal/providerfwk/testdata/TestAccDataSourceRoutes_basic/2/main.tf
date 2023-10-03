resource "junos_routing_instance" "testacc_data_routes" {
  name = "testacc_data_routes"
}
resource "junos_interface_physical" "testacc_data_routes" {
  name         = var.interface
  vlan_tagging = true
}
resource "junos_interface_logical" "testacc_interface_logical" {
  name             = "${junos_interface_physical.testacc_data_routes.name}.100"
  routing_instance = junos_routing_instance.testacc_data_routes.name
  family_inet {
    address {
      cidr_ip = "192.0.2.1/25"
    }
  }
}

data "junos_routes" "all" {}
data "junos_routes" "default" {
  table_name = "inet.0"
}
data "junos_routes" "testacc" {
  table_name = "${junos_routing_instance.testacc_data_routes.name}.inet.0"
}
