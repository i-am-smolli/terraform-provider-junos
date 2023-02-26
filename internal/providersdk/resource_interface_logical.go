package providersdk

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	balt "github.com/jeremmfr/go-utils/basicalter"
	bchk "github.com/jeremmfr/go-utils/basiccheck"
	jdecode "github.com/jeremmfr/junosdecode"
	"github.com/jeremmfr/terraform-provider-junos/internal/junos"
)

type interfaceLogicalOptions struct {
	disable                  bool
	vlanID                   int
	description              string
	routingInstance          string
	securityZone             string
	securityInboundProtocols []string
	securityInboundServices  []string
	familyInet               []map[string]interface{}
	familyInet6              []map[string]interface{}
	tunnel                   []map[string]interface{}
}

func resourceInterfaceLogical() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInterfaceLogicalCreate,
		ReadWithoutTimeout:   resourceInterfaceLogicalRead,
		UpdateWithoutTimeout: resourceInterfaceLogicalUpdate,
		DeleteWithoutTimeout: resourceInterfaceLogicalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceInterfaceLogicalImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if strings.Count(value, ".") != 1 {
						errors = append(errors, fmt.Errorf(
							"%q in %q need to have 1 dot", value, k))
					}

					return
				},
			},
			"st0_also_on_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disable": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"family_inet": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:          schema.TypeList,
							Optional:      true,
							ConflictsWith: []string{"family_inet.0.dhcp"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cidr_ip": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validateIPMaskFunc(),
									},
									"preferred": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"primary": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"vrrp_group": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"identifier": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 255),
												},
												"virtual_address": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.IsIPAddress,
													},
												},
												"accept_data": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"advertise_interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 255),
												},
												"advertisements_threshold": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 15),
												},
												"authentication_key": {
													Type:      schema.TypeString,
													Optional:  true,
													Sensitive: true,
												},
												"authentication_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice([]string{"md5", "simple"}, false),
												},
												"no_accept_data": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"no_preempt": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"preempt": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"priority": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 255),
												},
												"track_interface": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"interface": {
																Type:     schema.TypeString,
																Required: true,
															},
															"priority_cost": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 254),
															},
														},
													},
												},
												"track_route": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"route": {
																Type:     schema.TypeString,
																Required: true,
															},
															"routing_instance": {
																Type:     schema.TypeString,
																Required: true,
															},
															"priority_cost": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 254),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"dhcp": {
							Type:          schema.TypeList,
							Optional:      true,
							ConflictsWith: []string{"family_inet.0.address"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"srx_old_option_name": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"client_identifier_ascii": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"family_inet.0.dhcp.0.client_identifier_hexadecimal"},
									},
									"client_identifier_hexadecimal": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"family_inet.0.dhcp.0.client_identifier_ascii"},
										ValidateFunc: validation.StringMatch(
											regexp.MustCompile(`^[0-9a-fA-F]+$`),
											"must be hexadecimal digits (0-9, a-f, A-F)"),
									},
									"client_identifier_prefix_hostname": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"client_identifier_prefix_routing_instance_name": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"client_identifier_use_interface_description": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"device", "logical"}, false),
									},
									"client_identifier_userid_ascii": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"family_inet.0.dhcp.0.client_identifier_userid_hexadecimal"},
									},
									"client_identifier_userid_hexadecimal": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"family_inet.0.dhcp.0.client_identifier_userid_ascii"},
										ValidateFunc: validation.StringMatch(
											regexp.MustCompile(`^[0-9a-fA-F]+$`),
											"must be hexadecimal digits (0-9, a-f, A-F)"),
									},
									"force_discover": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"lease_time": {
										Type:          schema.TypeInt,
										Optional:      true,
										ConflictsWith: []string{"family_inet.0.dhcp.0.lease_time_infinite"},
										ValidateFunc:  validation.IntBetween(60, 2147483647),
									},
									"lease_time_infinite": {
										Type:          schema.TypeBool,
										Optional:      true,
										ConflictsWith: []string{"family_inet.0.dhcp.0.lease_time"},
									},
									"metric": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      -1,
										ValidateFunc: validation.IntBetween(0, 255),
									},
									"no_dns_install": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"options_no_hostname": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"retransmission_attempt": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      -1,
										ValidateFunc: validation.IntBetween(0, 50000),
									},
									"retransmission_interval": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(4, 64),
									},
									"server_address": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsIPv4Address,
									},
									"update_server": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"vendor_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 60),
									},
								},
							},
						},
						"filter_input": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
						},
						"filter_output": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
						},
						"mtu": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 9500),
						},
						"rpf_check": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fail_filter": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"mode_loose": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"sampling_input": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"sampling_output": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"family_inet6": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:          schema.TypeList,
							Optional:      true,
							ConflictsWith: []string{"family_inet6.0.dhcpv6_client"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cidr_ip": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validateIPMaskFunc(),
									},
									"preferred": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"primary": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"vrrp_group": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"identifier": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 255),
												},
												"virtual_address": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.IsIPAddress,
													},
												},
												"virtual_link_local_address": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.IsIPAddress,
												},
												"accept_data": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"advertise_interval": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(100, 40000),
												},
												"advertisements_threshold": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 15),
												},
												"no_accept_data": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"no_preempt": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"preempt": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"priority": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(1, 255),
												},
												"track_interface": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"interface": {
																Type:     schema.TypeString,
																Required: true,
															},
															"priority_cost": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 254),
															},
														},
													},
												},
												"track_route": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"route": {
																Type:     schema.TypeString,
																Required: true,
															},
															"routing_instance": {
																Type:     schema.TypeString,
																Required: true,
															},
															"priority_cost": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 254),
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"dad_disable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"dhcpv6_client": {
							Type:          schema.TypeList,
							Optional:      true,
							ConflictsWith: []string{"family_inet6.0.address"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_identifier_duid_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"duid-ll", "duid-llt", "vendor"}, false),
									},
									"client_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"autoconfig", "stateful"}, false),
									},
									"client_ia_type_na": {
										Type:     schema.TypeBool,
										Optional: true,
										AtLeastOneOf: []string{
											"family_inet6.0.dhcpv6_client.0.client_ia_type_na",
											"family_inet6.0.dhcpv6_client.0.client_ia_type_pd",
										},
									},
									"client_ia_type_pd": {
										Type:     schema.TypeBool,
										Optional: true,
										AtLeastOneOf: []string{
											"family_inet6.0.dhcpv6_client.0.client_ia_type_na",
											"family_inet6.0.dhcpv6_client.0.client_ia_type_pd",
										},
									},
									"no_dns_install": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"prefix_delegating_preferred_prefix_length": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      -1,
										ValidateFunc: validation.IntBetween(0, 64),
									},
									"prefix_delegating_sub_prefix_length": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 127),
									},
									"rapid_commit": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"req_option": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"retransmission_attempt": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      -1,
										ValidateFunc: validation.IntBetween(0, 9),
									},
									"update_router_advertisement_interface": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"update_server": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"filter_input": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
						},
						"filter_output": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
						},
						"mtu": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 9500),
						},
						"rpf_check": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fail_filter": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"mode_loose": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"sampling_input": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"sampling_output": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"routing_instance": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"security_inbound_protocols": {
				Type:         schema.TypeSet,
				Optional:     true,
				RequiredWith: []string{"security_zone"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"security_inbound_services": {
				Type:         schema.TypeSet,
				Optional:     true,
				RequiredWith: []string{"security_zone"},
				Elem:         &schema.Schema{Type: schema.TypeString},
			},
			"security_zone": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"tunnel": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"source": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsIPAddress,
						},
						"allow_fragmentation": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"tunnel.0.do_not_fragment"},
						},
						"do_not_fragment": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"tunnel.0.allow_fragmentation"},
						},
						"flow_label": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntBetween(0, 1048575),
						},
						"no_path_mtu_discovery": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"tunnel.0.path_mtu_discovery"},
						},
						"path_mtu_discovery": {
							Type:          schema.TypeBool,
							Optional:      true,
							ConflictsWith: []string{"tunnel.0.no_path_mtu_discovery"},
						},
						"routing_instance_destination": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
						},
						"traffic_class": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntBetween(0, 255),
						},
						"ttl": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 255),
						},
					},
				},
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 4094),
			},
			"vlan_no_compute": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceInterfaceLogicalCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clt := m.(*junos.Client)
	if clt.FakeCreateSetFile() {
		junSess := clt.NewSessionWithoutNetconf(ctx)
		if err := delInterfaceNC(d, clt.GroupInterfaceDelete(), junSess); err != nil {
			return diag.FromErr(err)
		}
		if err := setInterfaceLogical(d, junSess); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(d.Get("name").(string))

		return nil
	}
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer junSess.Close()
	if err := junSess.ConfigLock(ctx); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	ncInt, emptyInt, _, err := checkInterfaceLogicalNCEmpty(d.Get("name").(string), clt.GroupInterfaceDelete(), junSess)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	if !ncInt && !emptyInt {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(fmt.Errorf("interface %s already configured", d.Get("name").(string)))...)
	}
	if ncInt {
		if err := delInterfaceNC(d, clt.GroupInterfaceDelete(), junSess); err != nil {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns, diag.FromErr(err)...)
		}
	}
	if d.Get("security_zone").(string) != "" {
		if !junSess.CheckCompatibilitySecurity() {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns, diag.FromErr(fmt.Errorf("security zone not compatible with Junos device %s",
				junSess.SystemInformation.HardwareModel))...)
		}
		zonesExists, err := checkSecurityZonesExists(d.Get("security_zone").(string), junSess)
		if err != nil {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns, diag.FromErr(err)...)
		}
		if !zonesExists {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns,
				diag.FromErr(fmt.Errorf("security zone %v doesn't exist", d.Get("security_zone").(string)))...)
		}
	}
	if d.Get("routing_instance").(string) != "" {
		instanceExists, err := checkRoutingInstanceExists(d.Get("routing_instance").(string), junSess)
		if err != nil {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns, diag.FromErr(err)...)
		}
		if !instanceExists {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns,
				diag.FromErr(fmt.Errorf("routing instance %v doesn't exist", d.Get("routing_instance").(string)))...)
		}
	}
	if err := setInterfaceLogical(d, junSess); err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := junSess.CommitConf("create resource junos_interface_logical")
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	ncInt, emptyInt, setInt, err := checkInterfaceLogicalNCEmpty(
		d.Get("name").(string),
		clt.GroupInterfaceDelete(),
		junSess,
	)
	if err != nil {
		return append(diagWarns, diag.FromErr(err)...)
	}
	if ncInt {
		return append(diagWarns, diag.FromErr(fmt.Errorf("interface %v always disable (NC) after commit "+
			"=> check your config", d.Get("name").(string)))...)
	}
	if emptyInt && !setInt {
		intExists, err := checkInterfaceExists(d.Get("name").(string), junSess)
		if err != nil {
			return append(diagWarns, diag.FromErr(err)...)
		}
		if !intExists {
			return append(diagWarns, diag.FromErr(fmt.Errorf("interface %v not exists and "+
				"config can't found after commit => check your config", d.Get("name").(string)))...)
		}
	}
	d.SetId(d.Get("name").(string))

	return append(diagWarns, resourceInterfaceLogicalReadWJunSess(d, clt, junSess)...)
}

func resourceInterfaceLogicalRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clt := m.(*junos.Client)
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer junSess.Close()

	return resourceInterfaceLogicalReadWJunSess(d, clt, junSess)
}

func resourceInterfaceLogicalReadWJunSess(d *schema.ResourceData, clt *junos.Client, junSess *junos.Session,
) diag.Diagnostics {
	junos.MutexLock()
	ncInt, emptyInt, setInt, err := checkInterfaceLogicalNCEmpty(
		d.Get("name").(string),
		clt.GroupInterfaceDelete(),
		junSess,
	)
	if err != nil {
		junos.MutexUnlock()

		return diag.FromErr(err)
	}
	if ncInt {
		d.SetId("")
		junos.MutexUnlock()

		return nil
	}
	if emptyInt && !setInt {
		intExists, err := checkInterfaceExists(d.Get("name").(string), junSess)
		if err != nil {
			junos.MutexUnlock()

			return diag.FromErr(err)
		}
		if !intExists {
			d.SetId("")
			junos.MutexUnlock()

			return nil
		}
	}
	interfaceLogicalOpt, err := readInterfaceLogical(d.Get("name").(string), junSess)
	junos.MutexUnlock()
	if err != nil {
		return diag.FromErr(err)
	}
	fillInterfaceLogicalData(d, interfaceLogicalOpt)

	return nil
}

func resourceInterfaceLogicalUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	clt := m.(*junos.Client)
	if clt.FakeUpdateAlso() {
		junSess := clt.NewSessionWithoutNetconf(ctx)
		if err := delInterfaceLogicalOpts(d, junSess); err != nil {
			return diag.FromErr(err)
		}
		if d.HasChange("security_zone") {
			if oSecurityZone, _ := d.GetChange("security_zone"); oSecurityZone.(string) != "" {
				if err := delZoneInterfaceLogical(oSecurityZone.(string), d, junSess); err != nil {
					return diag.FromErr(err)
				}
			}
		} else if v := d.Get("security_zone").(string); v != "" {
			if err := delZoneInterfaceLogical(v, d, junSess); err != nil {
				return diag.FromErr(err)
			}
		}
		if d.HasChange("routing_instance") {
			if oRoutingInstance, _ := d.GetChange("routing_instance"); oRoutingInstance.(string) != "" {
				if err := delRoutingInstanceInterfaceLogical(oRoutingInstance.(string), d, junSess); err != nil {
					return diag.FromErr(err)
				}
			}
		}
		if err := setInterfaceLogical(d, junSess); err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)

		return nil
	}
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer junSess.Close()
	if err := junSess.ConfigLock(ctx); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	if err := delInterfaceLogicalOpts(d, junSess); err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	if d.HasChange("security_zone") {
		oSecurityZone, nSecurityZone := d.GetChange("security_zone")
		if nSecurityZone.(string) != "" {
			if !junSess.CheckCompatibilitySecurity() {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns, diag.FromErr(fmt.Errorf("security zone not compatible with Junos device %s",
					junSess.SystemInformation.HardwareModel))...)
			}
			zonesExists, err := checkSecurityZonesExists(nSecurityZone.(string), junSess)
			if err != nil {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns, diag.FromErr(err)...)
			}
			if !zonesExists {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns, diag.FromErr(fmt.Errorf("security zone %v doesn't exist", nSecurityZone.(string)))...)
			}
		}
		if oSecurityZone.(string) != "" {
			err = delZoneInterfaceLogical(oSecurityZone.(string), d, junSess)
			if err != nil {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns, diag.FromErr(err)...)
			}
		}
	} else if v := d.Get("security_zone").(string); v != "" {
		if err := delZoneInterfaceLogical(v, d, junSess); err != nil {
			appendDiagWarns(&diagWarns, junSess.ConfigClear())

			return append(diagWarns, diag.FromErr(err)...)
		}
	}
	if d.HasChange("routing_instance") {
		oRoutingInstance, nRoutingInstance := d.GetChange("routing_instance")
		if nRoutingInstance.(string) != "" {
			instanceExists, err := checkRoutingInstanceExists(nRoutingInstance.(string), junSess)
			if err != nil {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns, diag.FromErr(err)...)
			}
			if !instanceExists {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns,
					diag.FromErr(fmt.Errorf("routing instance %v doesn't exist", nRoutingInstance.(string)))...)
			}
		}
		if oRoutingInstance.(string) != "" {
			err = delRoutingInstanceInterfaceLogical(oRoutingInstance.(string), d, junSess)
			if err != nil {
				appendDiagWarns(&diagWarns, junSess.ConfigClear())

				return append(diagWarns, diag.FromErr(err)...)
			}
		}
	}
	if err := setInterfaceLogical(d, junSess); err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := junSess.CommitConf("update resource junos_interface_logical")
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	d.Partial(false)

	return append(diagWarns, resourceInterfaceLogicalReadWJunSess(d, clt, junSess)...)
}

func resourceInterfaceLogicalDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clt := m.(*junos.Client)
	if clt.FakeDeleteAlso() {
		junSess := clt.NewSessionWithoutNetconf(ctx)
		if err := delInterfaceLogical(d, junSess); err != nil {
			return diag.FromErr(err)
		}

		return nil
	}
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer junSess.Close()
	if err := junSess.ConfigLock(ctx); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	if err := delInterfaceLogical(d, junSess); err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := junSess.CommitConf("delete resource junos_interface_logical")
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}

	return diagWarns
}

func resourceInterfaceLogicalImport(ctx context.Context, d *schema.ResourceData, m interface{},
) ([]*schema.ResourceData, error) {
	if strings.Count(d.Id(), ".") != 1 {
		return nil, fmt.Errorf("name of interface %s need to have 1 dot", d.Id())
	}
	clt := m.(*junos.Client)
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return nil, err
	}
	defer junSess.Close()
	result := make([]*schema.ResourceData, 1)
	ncInt, emptyInt, setInt, err := checkInterfaceLogicalNCEmpty(d.Id(), clt.GroupInterfaceDelete(), junSess)
	if err != nil {
		return nil, err
	}
	if ncInt {
		return nil, fmt.Errorf("interface '%v' is disabled (NC), import is not possible", d.Id())
	}
	if emptyInt && !setInt {
		intExists, err := checkInterfaceExists(d.Id(), junSess)
		if err != nil {
			return nil, err
		}
		if !intExists {
			return nil, fmt.Errorf("don't find interface with id '%v' (id must be <name>)", d.Id())
		}
	}
	interfaceLogicalOpt, err := readInterfaceLogical(d.Id(), junSess)
	if err != nil {
		return nil, err
	}
	if tfErr := d.Set("name", d.Id()); tfErr != nil {
		panic(tfErr)
	}
	if interfaceLogicalOpt.vlanID == 0 {
		intCut := strings.Split(d.Id(), ".")
		if !bchk.InSlice(intCut[0], []string{junos.St0Word, "irb", "vlan"}) &&
			intCut[1] != "0" {
			if tfErr := d.Set("vlan_no_compute", true); tfErr != nil {
				panic(tfErr)
			}
		}
	}

	fillInterfaceLogicalData(d, interfaceLogicalOpt)

	result[0] = d

	return result, nil
}

func checkInterfaceLogicalNCEmpty(interFace, groupInterfaceDelete string, junSess *junos.Session,
) (ncInt, emtyInt, justSet bool, _err error) {
	showConfig, err := junSess.Command(junos.CmdShowConfig + "interfaces " + interFace + junos.PipeDisplaySetRelative)
	if err != nil {
		return false, false, false, err
	}
	showConfigLines := make([]string, 0)
	// remove unused lines
	for _, item := range strings.Split(showConfig, "\n") {
		// exclude ethernet-switching (parameters in junos_interface_physical)
		if strings.Contains(item, "ethernet-switching") {
			continue
		}
		if strings.Contains(item, junos.XMLStartTagConfigOut) {
			continue
		}
		if strings.Contains(item, junos.XMLEndTagConfigOut) {
			break
		}
		if item == "" {
			continue
		}
		showConfigLines = append(showConfigLines, item)
	}
	if len(showConfigLines) == 0 {
		return false, true, true, nil
	}
	showConfig = strings.Join(showConfigLines, "\n")
	if groupInterfaceDelete != "" {
		if showConfig == "set apply-groups "+groupInterfaceDelete {
			return true, false, false, nil
		}
	}
	if showConfig == "set description NC\nset disable" ||
		showConfig == "set disable\nset description NC" {
		return true, false, false, nil
	}
	switch {
	case showConfig == junos.SetLS:
		return false, true, true, nil
	case showConfig == junos.EmptyW:
		return false, true, false, nil
	default:
		return false, false, false, nil
	}
}

func setInterfaceLogical(d *schema.ResourceData, junSess *junos.Session) error {
	intCut := strings.Split(d.Get("name").(string), ".")
	if len(intCut) != 2 {
		return fmt.Errorf("the name %s doesn't contain one dot", d.Get("name").(string))
	}
	configSet := make([]string, 0)
	setPrefix := "set interfaces " + d.Get("name").(string) + " "
	configSet = append(configSet, setPrefix)
	if d.Get("description").(string) != "" {
		configSet = append(configSet, setPrefix+"description \""+d.Get("description").(string)+"\"")
	}
	if d.Get("disable").(bool) {
		if d.Get("description").(string) == "NC" {
			return fmt.Errorf("disable=true and description=NC is not allowed " +
				"because the provider might consider the resource deleted")
		}
		configSet = append(configSet, setPrefix+"disable")
	}
	for _, v := range d.Get("family_inet").([]interface{}) {
		configSet = append(configSet, setPrefix+"family inet")
		if v != nil {
			familyInet := v.(map[string]interface{})
			configSetFamilyInet, err := setFamilyAddress(familyInet, setPrefix, junos.InetW)
			if err != nil {
				return err
			}
			configSet = append(configSet, configSetFamilyInet...)
			for _, dhcp := range familyInet["dhcp"].([]interface{}) {
				configSet = append(configSet, setFamilyInetDhcp(dhcp.(map[string]interface{}), setPrefix)...)
			}
			if familyInet["filter_input"].(string) != "" {
				configSet = append(configSet, setPrefix+"family inet filter input "+
					familyInet["filter_input"].(string))
			}
			if familyInet["filter_output"].(string) != "" {
				configSet = append(configSet, setPrefix+"family inet filter output "+
					familyInet["filter_output"].(string))
			}
			if familyInet["mtu"].(int) > 0 {
				configSet = append(configSet, setPrefix+"family inet mtu "+
					strconv.Itoa(familyInet["mtu"].(int)))
			}
			for _, v2 := range familyInet["rpf_check"].([]interface{}) {
				configSet = append(configSet, setPrefix+"family inet rpf-check")
				if v2 != nil {
					rpfCheck := v2.(map[string]interface{})
					if rpfCheck["fail_filter"].(string) != "" {
						configSet = append(configSet, setPrefix+"family inet rpf-check fail-filter "+
							"\""+rpfCheck["fail_filter"].(string)+"\"")
					}
					if rpfCheck["mode_loose"].(bool) {
						configSet = append(configSet, setPrefix+"family inet rpf-check mode loose ")
					}
				}
			}
			if familyInet["sampling_input"].(bool) {
				configSet = append(configSet, setPrefix+"family inet sampling input")
			}
			if familyInet["sampling_output"].(bool) {
				configSet = append(configSet, setPrefix+"family inet sampling output")
			}
		}
	}
	for _, v := range d.Get("family_inet6").([]interface{}) {
		configSet = append(configSet, setPrefix+"family inet6")
		if v != nil {
			familyInet6 := v.(map[string]interface{})
			configSetFamilyInet6, err := setFamilyAddress(familyInet6, setPrefix, junos.Inet6W)
			if err != nil {
				return err
			}
			configSet = append(configSet, configSetFamilyInet6...)
			for _, dhcp := range familyInet6["dhcpv6_client"].([]interface{}) {
				configSet = append(configSet, setFamilyInet6Dhcpv6Client(dhcp.(map[string]interface{}), setPrefix)...)
			}
			if familyInet6["dad_disable"].(bool) {
				configSet = append(configSet, setPrefix+"family inet6 dad-disable")
			}
			if familyInet6["filter_input"].(string) != "" {
				configSet = append(configSet, setPrefix+"family inet6 filter input "+
					familyInet6["filter_input"].(string))
			}
			if familyInet6["filter_output"].(string) != "" {
				configSet = append(configSet, setPrefix+"family inet6 filter output "+
					familyInet6["filter_output"].(string))
			}
			if familyInet6["mtu"].(int) > 0 {
				configSet = append(configSet, setPrefix+"family inet6 mtu "+
					strconv.Itoa(familyInet6["mtu"].(int)))
			}
			for _, v2 := range familyInet6["rpf_check"].([]interface{}) {
				configSet = append(configSet, setPrefix+"family inet6 rpf-check")
				if v2 != nil {
					rpfCheck := v2.(map[string]interface{})
					if rpfCheck["fail_filter"].(string) != "" {
						configSet = append(configSet, setPrefix+"family inet6 rpf-check fail-filter "+
							"\""+rpfCheck["fail_filter"].(string)+"\"")
					}
					if rpfCheck["mode_loose"].(bool) {
						configSet = append(configSet, setPrefix+"family inet6 rpf-check mode loose ")
					}
				}
			}
			if familyInet6["sampling_input"].(bool) {
				configSet = append(configSet, setPrefix+"family inet6 sampling input")
			}
			if familyInet6["sampling_output"].(bool) {
				configSet = append(configSet, setPrefix+"family inet6 sampling output")
			}
		}
	}
	if instance := d.Get("routing_instance").(string); instance != "" {
		configSet = append(configSet, junos.SetRoutingInstances+instance+" interface "+d.Get("name").(string))
	}
	if zone := d.Get("security_zone").(string); zone != "" {
		configSet = append(configSet, "set security zones security-zone "+zone+
			" interfaces "+d.Get("name").(string))
		for _, v := range sortSetOfString(d.Get("security_inbound_protocols").(*schema.Set).List()) {
			configSet = append(configSet, "set security zones security-zone "+zone+
				" interfaces "+d.Get("name").(string)+" host-inbound-traffic protocols "+v)
		}
		for _, v := range sortSetOfString(d.Get("security_inbound_services").(*schema.Set).List()) {
			configSet = append(configSet, "set security zones security-zone "+zone+
				" interfaces "+d.Get("name").(string)+" host-inbound-traffic system-services "+v)
		}
	}
	for _, tunnelElem := range d.Get("tunnel").([]interface{}) {
		tunnel := tunnelElem.(map[string]interface{})
		configSet = append(configSet, setPrefix+"tunnel destination "+tunnel["destination"].(string))
		configSet = append(configSet, setPrefix+"tunnel source "+tunnel["source"].(string))
		if tunnel["allow_fragmentation"].(bool) {
			configSet = append(configSet, setPrefix+"tunnel allow-fragmentation")
		}
		if tunnel["do_not_fragment"].(bool) {
			configSet = append(configSet, setPrefix+"tunnel do-not-fragment")
		}
		if v := tunnel["flow_label"].(int); v != -1 {
			configSet = append(configSet, setPrefix+"tunnel flow-label "+strconv.Itoa(v))
		}
		if tunnel["no_path_mtu_discovery"].(bool) {
			configSet = append(configSet, setPrefix+"tunnel no-path-mtu-discovery")
		}
		if tunnel["path_mtu_discovery"].(bool) {
			configSet = append(configSet, setPrefix+"tunnel path-mtu-discovery")
		}
		if v := tunnel["routing_instance_destination"].(string); v != "" {
			configSet = append(configSet, setPrefix+"tunnel routing-instance destination "+v)
		}
		if v := tunnel["traffic_class"].(int); v != -1 {
			configSet = append(configSet, setPrefix+"tunnel traffic-class "+strconv.Itoa(v))
		}
		if v := tunnel["ttl"].(int); v != 0 {
			configSet = append(configSet, setPrefix+"tunnel ttl "+strconv.Itoa(v))
		}
	}
	if d.Get("vlan_id").(int) != 0 {
		configSet = append(configSet, setPrefix+"vlan-id "+strconv.Itoa(d.Get("vlan_id").(int)))
	} else if !bchk.InSlice(intCut[0], []string{junos.St0Word, "irb", "vlan"}) &&
		intCut[1] != "0" && !d.Get("vlan_no_compute").(bool) {
		configSet = append(configSet, setPrefix+"vlan-id "+intCut[1])
	}

	return junSess.ConfigSet(configSet)
}

func readInterfaceLogical(interFace string, junSess *junos.Session,
) (confRead interfaceLogicalOptions, err error) {
	showConfig, err := junSess.Command(junos.CmdShowConfig + "interfaces " + interFace + junos.PipeDisplaySetRelative)
	if err != nil {
		return confRead, err
	}

	if showConfig != junos.EmptyW {
		for _, item := range strings.Split(showConfig, "\n") {
			// exclude ethernet-switching (parameters in junos_interface_physical)
			if strings.Contains(item, "ethernet-switching") {
				continue
			}
			if strings.Contains(item, junos.XMLStartTagConfigOut) {
				continue
			}
			if strings.Contains(item, junos.XMLEndTagConfigOut) {
				break
			}
			itemTrim := strings.TrimPrefix(item, junos.SetLS)
			switch {
			case balt.CutPrefixInString(&itemTrim, "description "):
				confRead.description = strings.Trim(itemTrim, "\"")
			case itemTrim == "disable":
				confRead.disable = true
			case balt.CutPrefixInString(&itemTrim, "family inet6"):
				if len(confRead.familyInet6) == 0 {
					confRead.familyInet6 = append(confRead.familyInet6, map[string]interface{}{
						"address":         make([]map[string]interface{}, 0),
						"dad_disable":     false,
						"dhcpv6_client":   make([]map[string]interface{}, 0),
						"filter_input":    "",
						"filter_output":   "",
						"mtu":             0,
						"rpf_check":       make([]map[string]interface{}, 0),
						"sampling_input":  false,
						"sampling_output": false,
					})
				}
				switch {
				case balt.CutPrefixInString(&itemTrim, " address "):
					confRead.familyInet6[0]["address"], err = readFamilyInetAddress(
						itemTrim, confRead.familyInet6[0]["address"].([]map[string]interface{}), junos.Inet6W)
					if err != nil {
						return confRead, err
					}
				case balt.CutPrefixInString(&itemTrim, " dhcpv6-client "):
					if len(confRead.familyInet6[0]["dhcpv6_client"].([]map[string]interface{})) == 0 {
						confRead.familyInet6[0]["dhcpv6_client"] = append(
							confRead.familyInet6[0]["dhcpv6_client"].([]map[string]interface{}), map[string]interface{}{
								"client_identifier_duid_type":               "",
								"client_type":                               "",
								"client_ia_type_na":                         false,
								"client_ia_type_pd":                         false,
								"no_dns_install":                            false,
								"prefix_delegating_preferred_prefix_length": -1,
								"prefix_delegating_sub_prefix_length":       0,
								"rapid_commit":                              false,
								"req_option":                                make([]string, 0),
								"retransmission_attempt":                    -1,
								"update_router_advertisement_interface":     make([]string, 0),
								"update_server":                             false,
							})
					}
					if err := readFamilyInet6Dhcpv6Client(
						itemTrim, confRead.familyInet6[0]["dhcpv6_client"].([]map[string]interface{})[0]); err != nil {
						return confRead, err
					}
				case itemTrim == " dad-disable":
					confRead.familyInet6[0]["dad_disable"] = true
				case balt.CutPrefixInString(&itemTrim, " filter input "):
					confRead.familyInet6[0]["filter_input"] = itemTrim
				case balt.CutPrefixInString(&itemTrim, " filter output "):
					confRead.familyInet6[0]["filter_output"] = itemTrim
				case balt.CutPrefixInString(&itemTrim, " mtu "):
					confRead.familyInet6[0]["mtu"], err = strconv.Atoi(itemTrim)
					if err != nil {
						return confRead, fmt.Errorf(failedConvAtoiError, itemTrim, err)
					}
				case balt.CutPrefixInString(&itemTrim, " rpf-check"):
					if len(confRead.familyInet6[0]["rpf_check"].([]map[string]interface{})) == 0 {
						confRead.familyInet6[0]["rpf_check"] = append(
							confRead.familyInet6[0]["rpf_check"].([]map[string]interface{}), map[string]interface{}{
								"fail_filter": "",
								"mode_loose":  false,
							})
					}
					switch {
					case balt.CutPrefixInString(&itemTrim, " fail-filter "):
						confRead.familyInet6[0]["rpf_check"].([]map[string]interface{})[0]["fail_filter"] = strings.Trim(
							itemTrim, "\"")
					case itemTrim == " mode loose":
						confRead.familyInet6[0]["rpf_check"].([]map[string]interface{})[0]["mode_loose"] = true
					}
				case itemTrim == " sampling input":
					confRead.familyInet6[0]["sampling_input"] = true
				case itemTrim == " sampling output":
					confRead.familyInet6[0]["sampling_output"] = true
				}
			case balt.CutPrefixInString(&itemTrim, "family inet"):
				if len(confRead.familyInet) == 0 {
					confRead.familyInet = append(confRead.familyInet, map[string]interface{}{
						"address":         make([]map[string]interface{}, 0),
						"dhcp":            make([]map[string]interface{}, 0),
						"mtu":             0,
						"filter_input":    "",
						"filter_output":   "",
						"rpf_check":       make([]map[string]interface{}, 0),
						"sampling_input":  false,
						"sampling_output": false,
					})
				}
				switch {
				case balt.CutPrefixInString(&itemTrim, " address "):
					confRead.familyInet[0]["address"], err = readFamilyInetAddress(
						itemTrim, confRead.familyInet[0]["address"].([]map[string]interface{}), junos.InetW)
					if err != nil {
						return confRead, err
					}
				case strings.HasPrefix(itemTrim, " dhcp"):
					if len(confRead.familyInet[0]["dhcp"].([]map[string]interface{})) == 0 {
						confRead.familyInet[0]["dhcp"] = append(
							confRead.familyInet[0]["dhcp"].([]map[string]interface{}), map[string]interface{}{
								"srx_old_option_name":                            strings.HasPrefix(itemTrim, " dhcp-client"),
								"client_identifier_ascii":                        "",
								"client_identifier_hexadecimal":                  "",
								"client_identifier_prefix_hostname":              false,
								"client_identifier_prefix_routing_instance_name": false,
								"client_identifier_use_interface_description":    "",
								"client_identifier_userid_ascii":                 "",
								"client_identifier_userid_hexadecimal":           "",
								"force_discover":                                 false,
								"lease_time":                                     0,
								"lease_time_infinite":                            false,
								"metric":                                         -1,
								"no_dns_install":                                 false,
								"options_no_hostname":                            false,
								"retransmission_attempt":                         -1,
								"retransmission_interval":                        0,
								"server_address":                                 "",
								"update_server":                                  false,
								"vendor_id":                                      "",
							})
					}
					if balt.CutPrefixInString(&itemTrim, " dhcp ") || balt.CutPrefixInString(&itemTrim, " dhcp-client ") {
						if err := readFamilyInetDhcp(
							itemTrim, confRead.familyInet[0]["dhcp"].([]map[string]interface{})[0]); err != nil {
							return confRead, err
						}
					}
				case balt.CutPrefixInString(&itemTrim, " filter input "):
					confRead.familyInet[0]["filter_input"] = itemTrim
				case balt.CutPrefixInString(&itemTrim, " filter output "):
					confRead.familyInet[0]["filter_output"] = itemTrim
				case balt.CutPrefixInString(&itemTrim, " mtu "):
					confRead.familyInet[0]["mtu"], err = strconv.Atoi(itemTrim)
					if err != nil {
						return confRead, fmt.Errorf(failedConvAtoiError, itemTrim, err)
					}
				case balt.CutPrefixInString(&itemTrim, " rpf-check"):
					if len(confRead.familyInet[0]["rpf_check"].([]map[string]interface{})) == 0 {
						confRead.familyInet[0]["rpf_check"] = append(
							confRead.familyInet[0]["rpf_check"].([]map[string]interface{}), map[string]interface{}{
								"fail_filter": "",
								"mode_loose":  false,
							})
					}
					switch {
					case balt.CutPrefixInString(&itemTrim, " fail-filter "):
						confRead.familyInet[0]["rpf_check"].([]map[string]interface{})[0]["fail_filter"] = strings.Trim(
							itemTrim, "\"")
					case itemTrim == " mode loose":
						confRead.familyInet[0]["rpf_check"].([]map[string]interface{})[0]["mode_loose"] = true
					}
				case itemTrim == " sampling input":
					confRead.familyInet[0]["sampling_input"] = true
				case itemTrim == " sampling output":
					confRead.familyInet[0]["sampling_output"] = true
				}
			case balt.CutPrefixInString(&itemTrim, "tunnel "):
				if len(confRead.tunnel) == 0 {
					confRead.tunnel = append(confRead.tunnel, map[string]interface{}{
						"destination":                  "",
						"source":                       "",
						"allow_fragmentation":          false,
						"do_not_fragment":              false,
						"flow_label":                   -1,
						"no_path_mtu_discovery":        false,
						"path_mtu_discovery":           false,
						"routing_instance_destination": "",
						"traffic_class":                -1,
						"ttl":                          0,
					})
				}
				switch {
				case balt.CutPrefixInString(&itemTrim, "destination "):
					confRead.tunnel[0]["destination"] = itemTrim
				case balt.CutPrefixInString(&itemTrim, "source "):
					confRead.tunnel[0]["source"] = itemTrim
				case itemTrim == "allow-fragmentation":
					confRead.tunnel[0]["allow_fragmentation"] = true
				case itemTrim == "do-not-fragment":
					confRead.tunnel[0]["do_not_fragment"] = true
				case balt.CutPrefixInString(&itemTrim, "flow-label "):
					confRead.tunnel[0]["flow_label"], err = strconv.Atoi(itemTrim)
					if err != nil {
						return confRead, fmt.Errorf(failedConvAtoiError, itemTrim, err)
					}
				case itemTrim == "no-path-mtu-discovery":
					confRead.tunnel[0]["no_path_mtu_discovery"] = true
				case itemTrim == "path-mtu-discovery":
					confRead.tunnel[0]["path_mtu_discovery"] = true
				case balt.CutPrefixInString(&itemTrim, "routing-instance destination "):
					confRead.tunnel[0]["routing_instance_destination"] = itemTrim
				case balt.CutPrefixInString(&itemTrim, "traffic-class "):
					confRead.tunnel[0]["traffic_class"], err = strconv.Atoi(itemTrim)
					if err != nil {
						return confRead, fmt.Errorf(failedConvAtoiError, itemTrim, err)
					}
				case balt.CutPrefixInString(&itemTrim, "ttl "):
					confRead.tunnel[0]["ttl"], err = strconv.Atoi(itemTrim)
					if err != nil {
						return confRead, fmt.Errorf(failedConvAtoiError, itemTrim, err)
					}
				}
			case balt.CutPrefixInString(&itemTrim, "vlan-id "):
				confRead.vlanID, err = strconv.Atoi(itemTrim)
				if err != nil {
					return confRead, fmt.Errorf(failedConvAtoiError, itemTrim, err)
				}
			default:
				continue
			}
		}
	}
	showConfigRoutingInstances, err := junSess.Command(junos.CmdShowConfig +
		"routing-instances" + junos.PipeDisplaySetRelative)
	if err != nil {
		return confRead, err
	}
	regexpInt := regexp.MustCompile(`set \S+ interface ` + interFace + `$`)
	for _, item := range strings.Split(showConfigRoutingInstances, "\n") {
		intMatch := regexpInt.MatchString(item)
		if intMatch {
			confRead.routingInstance = strings.TrimPrefix(strings.TrimSuffix(item, " interface "+interFace), junos.SetLS)

			break
		}
	}
	if junSess.CheckCompatibilitySecurity() {
		showConfigSecurityZones, err := junSess.Command(junos.CmdShowConfig + "security zones" + junos.PipeDisplaySetRelative)
		if err != nil {
			return confRead, err
		}
		regexpInts := regexp.MustCompile(`set security-zone \S+ interfaces ` + interFace + `( host-inbound-traffic .*)?$`)
		for _, item := range strings.Split(showConfigSecurityZones, "\n") {
			intMatch := regexpInts.MatchString(item)
			if intMatch {
				itemTrimFields := strings.Split(strings.TrimPrefix(item, "set security-zone "), " ")
				confRead.securityZone = itemTrimFields[0]
				if err := confRead.readInterfaceLogicalSecurityInboundTraffic(interFace, junSess); err != nil {
					return confRead, err
				}

				break
			}
		}
	}

	return confRead, nil
}

func (confRead *interfaceLogicalOptions) readInterfaceLogicalSecurityInboundTraffic(
	interFace string, junSess *junos.Session,
) error {
	showConfig, err := junSess.Command(junos.CmdShowConfig +
		"security zones security-zone " + confRead.securityZone + " interfaces " + interFace + junos.PipeDisplaySetRelative)
	if err != nil {
		return err
	}

	if showConfig != junos.EmptyW {
		for _, item := range strings.Split(showConfig, "\n") {
			if strings.Contains(item, junos.XMLStartTagConfigOut) {
				continue
			}
			if strings.Contains(item, junos.XMLEndTagConfigOut) {
				break
			}
			itemTrim := strings.TrimPrefix(item, junos.SetLS)
			switch {
			case balt.CutPrefixInString(&itemTrim, "host-inbound-traffic protocols "):
				confRead.securityInboundProtocols = append(confRead.securityInboundProtocols, itemTrim)
			case balt.CutPrefixInString(&itemTrim, "host-inbound-traffic system-services "):
				confRead.securityInboundServices = append(confRead.securityInboundServices, itemTrim)
			}
		}
	}

	return nil
}

func delInterfaceLogical(d *schema.ResourceData, junSess *junos.Session) error {
	if err := junSess.ConfigSet([]string{"delete interfaces " + d.Get("name").(string)}); err != nil {
		return err
	}
	if strings.HasPrefix(d.Get("name").(string), "st0.") && !d.Get("st0_also_on_destroy").(bool) {
		// interface totally delete by
		// - junos_interface_st0_unit resource
		// else there is an interface st0.x empty
		err := junSess.ConfigSet([]string{"set interfaces " + d.Get("name").(string)})
		if err != nil {
			return err
		}
	}
	if d.Get("routing_instance").(string) != "" {
		if err := delRoutingInstanceInterfaceLogical(d.Get("routing_instance").(string), d, junSess); err != nil {
			return err
		}
	}
	if d.Get("security_zone").(string) != "" {
		if !junSess.HasNetconf() || junSess.CheckCompatibilitySecurity() {
			if err := delZoneInterfaceLogical(d.Get("security_zone").(string), d, junSess); err != nil {
				return err
			}
		}
	}

	return nil
}

func delInterfaceLogicalOpts(d *schema.ResourceData, junSess *junos.Session) error {
	configSet := make([]string, 0, 1)
	delPrefix := "delete interfaces " + d.Get("name").(string) + " "
	configSet = append(configSet,
		delPrefix+"description",
		delPrefix+"disable",
		delPrefix+"family inet",
		delPrefix+"family inet6",
		delPrefix+"tunnel",
	)

	return junSess.ConfigSet(configSet)
}

func delZoneInterfaceLogical(zone string, d *schema.ResourceData, junSess *junos.Session) error {
	configSet := make([]string, 0, 1)
	configSet = append(configSet, "delete security zones security-zone "+zone+" interfaces "+d.Get("name").(string))

	return junSess.ConfigSet(configSet)
}

func delRoutingInstanceInterfaceLogical(
	instance string, d *schema.ResourceData, junSess *junos.Session,
) error {
	configSet := make([]string, 0, 1)
	configSet = append(configSet, junos.DelRoutingInstances+instance+" interface "+d.Get("name").(string))

	return junSess.ConfigSet(configSet)
}

func fillInterfaceLogicalData(d *schema.ResourceData, interfaceLogicalOpt interfaceLogicalOptions) {
	if tfErr := d.Set("description", interfaceLogicalOpt.description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("disable", interfaceLogicalOpt.disable); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("family_inet", interfaceLogicalOpt.familyInet); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("family_inet6", interfaceLogicalOpt.familyInet6); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("routing_instance", interfaceLogicalOpt.routingInstance); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("security_inbound_protocols", interfaceLogicalOpt.securityInboundProtocols); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("security_inbound_services", interfaceLogicalOpt.securityInboundServices); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("tunnel", interfaceLogicalOpt.tunnel); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("security_zone", interfaceLogicalOpt.securityZone); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("vlan_id", interfaceLogicalOpt.vlanID); tfErr != nil {
		panic(tfErr)
	}
}

func readFamilyInetAddress(itemTrim string, inetAddress []map[string]interface{}, family string,
) ([]map[string]interface{}, error) {
	itemTrimFields := strings.Split(itemTrim, " ")
	balt.CutPrefixInString(&itemTrim, itemTrimFields[0]+" ")

	mAddr := genFamilyInetAddress(itemTrimFields[0])
	inetAddress = copyAndRemoveItemMapList("cidr_ip", mAddr, inetAddress)

	switch {
	case itemTrim == "primary":
		mAddr["primary"] = true
	case itemTrim == "preferred":
		mAddr["preferred"] = true
	case balt.CutPrefixInString(&itemTrim, "vrrp-group "), balt.CutPrefixInString(&itemTrim, "vrrp-inet6-group "):
		if len(itemTrimFields) < 3 { // <address> (vrrp-group|vrrp-inet6-group) <vrrpID>
			return inetAddress, fmt.Errorf(junos.CantReadValuesNotEnoughFields, "vrrp-group|vrrp-inet6-group", itemTrim)
		}
		vrrpGroup := genVRRPGroup(family)
		vrrpID, err := strconv.Atoi(itemTrimFields[2])
		if err != nil {
			return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
		vrrpGroup["identifier"] = vrrpID
		mAddr["vrrp_group"] = copyAndRemoveItemMapList("identifier", vrrpGroup,
			mAddr["vrrp_group"].([]map[string]interface{}))
		balt.CutPrefixInString(&itemTrim, itemTrimFields[2]+" ")
		switch {
		case balt.CutPrefixInString(&itemTrim, "virtual-address "):
			vrrpGroup["virtual_address"] = append(vrrpGroup["virtual_address"].([]string), itemTrim)
		case balt.CutPrefixInString(&itemTrim, "virtual-inet6-address "):
			vrrpGroup["virtual_address"] = append(vrrpGroup["virtual_address"].([]string), itemTrim)
		case balt.CutPrefixInString(&itemTrim, "virtual-link-local-address "):
			vrrpGroup["virtual_link_local_address"] = itemTrim
		case itemTrim == "accept-data":
			vrrpGroup["accept_data"] = true
		case balt.CutPrefixInString(&itemTrim, "advertise-interval "):
			vrrpGroup["advertise_interval"], err = strconv.Atoi(itemTrim)
			if err != nil {
				return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
			}
		case balt.CutPrefixInString(&itemTrim, "inet6-advertise-interval "):
			vrrpGroup["advertise_interval"], err = strconv.Atoi(itemTrim)
			if err != nil {
				return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
			}
		case balt.CutPrefixInString(&itemTrim, "advertisements-threshold "):
			vrrpGroup["advertisements_threshold"], err = strconv.Atoi(itemTrim)
			if err != nil {
				return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
			}
		case balt.CutPrefixInString(&itemTrim, "authentication-key "):
			vrrpGroup["authentication_key"], err = jdecode.Decode(strings.Trim(itemTrim, "\""))
			if err != nil {
				return inetAddress, fmt.Errorf("decoding authentication-key: %w", err)
			}
		case balt.CutPrefixInString(&itemTrim, "authentication-type "):
			vrrpGroup["authentication_type"] = itemTrim
		case itemTrim == "no-accept-data":
			vrrpGroup["no_accept_data"] = true
		case itemTrim == "no-preempt":
			vrrpGroup["no_preempt"] = true
		case itemTrim == "preempt":
			vrrpGroup["preempt"] = true
		case balt.CutPrefixInString(&itemTrim, "priority "):
			vrrpGroup["priority"], err = strconv.Atoi(itemTrim)
			if err != nil {
				return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
			}
		case balt.CutPrefixInString(&itemTrim, "track interface "):
			itemTrackFields := strings.Split(itemTrim, " ")
			if len(itemTrackFields) < 3 { // <interface> priority-cost <priority_cost>
				return inetAddress, fmt.Errorf(junos.CantReadValuesNotEnoughFields, "track interface", itemTrim)
			}
			cost, err := strconv.Atoi(itemTrackFields[2])
			if err != nil {
				return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
			}
			trackInt := map[string]interface{}{
				"interface":     itemTrackFields[0],
				"priority_cost": cost,
			}
			vrrpGroup["track_interface"] = append(vrrpGroup["track_interface"].([]map[string]interface{}), trackInt)
		case balt.CutPrefixInString(&itemTrim, "track route "):
			itemTrackFields := strings.Split(itemTrim, " ")
			if len(itemTrackFields) < 5 { // <route> routing-instance <routing_instance> priority-cost <priority_cost>
				return inetAddress, fmt.Errorf(junos.CantReadValuesNotEnoughFields, "track route", itemTrim)
			}
			cost, err := strconv.Atoi(itemTrackFields[4])
			if err != nil {
				return inetAddress, fmt.Errorf(failedConvAtoiError, itemTrim, err)
			}
			trackRoute := map[string]interface{}{
				"route":            itemTrackFields[0],
				"routing_instance": itemTrackFields[2],
				"priority_cost":    cost,
			}
			vrrpGroup["track_route"] = append(vrrpGroup["track_route"].([]map[string]interface{}), trackRoute)
		}
		mAddr["vrrp_group"] = append(mAddr["vrrp_group"].([]map[string]interface{}), vrrpGroup)
	}
	inetAddress = append(inetAddress, mAddr)

	return inetAddress, nil
}

func readFamilyInetDhcp(itemTrim string, dhcp map[string]interface{}) (err error) {
	switch {
	case balt.CutPrefixInString(&itemTrim, "client-identifier ascii "):
		dhcp["client_identifier_ascii"] = strings.Trim(itemTrim, "\"")
	case balt.CutPrefixInString(&itemTrim, "client-identifier hexadecimal "):
		dhcp["client_identifier_hexadecimal"] = itemTrim
	case itemTrim == "client-identifier prefix host-name":
		dhcp["client_identifier_prefix_hostname"] = true
	case itemTrim == "client-identifier prefix routing-instance-name":
		dhcp["client_identifier_prefix_routing_instance_name"] = true
	case balt.CutPrefixInString(&itemTrim, "client-identifier use-interface-description "):
		dhcp["client_identifier_use_interface_description"] = itemTrim
	case balt.CutPrefixInString(&itemTrim, "client-identifier user-id ascii "):
		dhcp["client_identifier_userid_ascii"] = strings.Trim(itemTrim, "\"")
	case balt.CutPrefixInString(&itemTrim, "client-identifier user-id hexadecimal "):
		dhcp["client_identifier_userid_hexadecimal"] = itemTrim
	case itemTrim == "force-discover":
		dhcp["force_discover"] = true
	case itemTrim == "lease-time infinite":
		dhcp["lease_time_infinite"] = true
	case balt.CutPrefixInString(&itemTrim, "lease-time "):
		dhcp["lease_time"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case balt.CutPrefixInString(&itemTrim, "metric "):
		dhcp["metric"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case itemTrim == "no-dns-install":
		dhcp["no_dns_install"] = true
	case itemTrim == "options no-hostname":
		dhcp["options_no_hostname"] = true
	case balt.CutPrefixInString(&itemTrim, "retransmission-attempt "):
		dhcp["retransmission_attempt"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case balt.CutPrefixInString(&itemTrim, "retransmission-interval "):
		dhcp["retransmission_interval"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case balt.CutPrefixInString(&itemTrim, "server-address "):
		dhcp["server_address"] = itemTrim
	case itemTrim == "update-server":
		dhcp["update_server"] = true
	case balt.CutPrefixInString(&itemTrim, "vendor-id "):
		dhcp["vendor_id"] = strings.Trim(itemTrim, "\"")
	}

	return nil
}

func readFamilyInet6Dhcpv6Client(itemTrim string, dhcp map[string]interface{}) (err error) {
	switch {
	case balt.CutPrefixInString(&itemTrim, "client-identifier duid-type "):
		dhcp["client_identifier_duid_type"] = itemTrim
	case balt.CutPrefixInString(&itemTrim, "client-type "):
		dhcp["client_type"] = itemTrim
	case itemTrim == "client-ia-type ia-na":
		dhcp["client_ia_type_na"] = true
	case itemTrim == "client-ia-type ia-pd":
		dhcp["client_ia_type_pd"] = true
	case itemTrim == "no-dns-install":
		dhcp["no_dns_install"] = true
	case balt.CutPrefixInString(&itemTrim, "prefix-delegating preferred-prefix-length "):
		dhcp["prefix_delegating_preferred_prefix_length"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case balt.CutPrefixInString(&itemTrim, "prefix-delegating sub-prefix-length "):
		dhcp["prefix_delegating_sub_prefix_length"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case itemTrim == "rapid-commit":
		dhcp["rapid_commit"] = true
	case balt.CutPrefixInString(&itemTrim, "req-option "):
		dhcp["req_option"] = append(dhcp["req_option"].([]string), itemTrim)
	case balt.CutPrefixInString(&itemTrim, "retransmission-attempt "):
		dhcp["retransmission_attempt"], err = strconv.Atoi(itemTrim)
		if err != nil {
			return fmt.Errorf(failedConvAtoiError, itemTrim, err)
		}
	case balt.CutPrefixInString(&itemTrim, "update-router-advertisement interface "):
		dhcp["update_router_advertisement_interface"] = append(
			dhcp["update_router_advertisement_interface"].([]string),
			itemTrim,
		)
	case itemTrim == "update-server":
		dhcp["update_server"] = true
	}

	return nil
}

func setFamilyAddress(inetAddress map[string]interface{}, setPrefix, family string) ([]string, error) {
	configSet := make([]string, 0)
	if family != junos.InetW && family != junos.Inet6W {
		panic(fmt.Sprintf("setFamilyAddress() unknown family %v", family))
	}
	addressCIDRIPList := make([]string, 0)
	for _, address := range inetAddress["address"].([]interface{}) {
		addressMap := address.(map[string]interface{})
		if bchk.InSlice(addressMap["cidr_ip"].(string), addressCIDRIPList) {
			if family == junos.InetW {
				return configSet, fmt.Errorf("multiple blocks family_inet with the same cidr_ip %s",
					addressMap["cidr_ip"].(string))
			}
			if family == junos.Inet6W {
				return configSet, fmt.Errorf("multiple blocks family_inet6 with the same cidr_ip %s",
					addressMap["cidr_ip"].(string))
			}
		}
		addressCIDRIPList = append(addressCIDRIPList, addressMap["cidr_ip"].(string))
		setPrefixAddress := setPrefix + "family " + family + " address " + addressMap["cidr_ip"].(string)
		configSet = append(configSet, setPrefixAddress)
		if addressMap["preferred"].(bool) {
			configSet = append(configSet, setPrefixAddress+" preferred")
		}
		if addressMap["primary"].(bool) {
			configSet = append(configSet, setPrefixAddress+" primary")
		}
		vrrpGroupIDList := make([]int, 0)
		for _, vrrpGroup := range addressMap["vrrp_group"].([]interface{}) {
			if strings.Contains(setPrefix, "set interfaces st0 unit") {
				return configSet, fmt.Errorf("vrrp not available on st0")
			}
			vrrpGroupMap := vrrpGroup.(map[string]interface{})
			if vrrpGroupMap["no_preempt"].(bool) && vrrpGroupMap["preempt"].(bool) {
				return configSet, fmt.Errorf("ConflictsWith no_preempt and preempt")
			}
			if vrrpGroupMap["no_accept_data"].(bool) && vrrpGroupMap["accept_data"].(bool) {
				return configSet, fmt.Errorf("ConflictsWith no_accept_data and accept_data")
			}
			if bchk.InSlice(vrrpGroupMap["identifier"].(int), vrrpGroupIDList) {
				return configSet, fmt.Errorf("multiple blocks vrrp_group with the same identifier %d",
					vrrpGroupMap["identifier"].(int))
			}
			vrrpGroupIDList = append(vrrpGroupIDList, vrrpGroupMap["identifier"].(int))
			var setNameAddVrrp string
			switch family {
			case junos.InetW:
				setNameAddVrrp = setPrefixAddress + " vrrp-group " + strconv.Itoa(vrrpGroupMap["identifier"].(int))
				for _, ip := range vrrpGroupMap["virtual_address"].([]interface{}) {
					configSet = append(configSet, setNameAddVrrp+" virtual-address "+ip.(string))
				}
				if vrrpGroupMap["advertise_interval"].(int) != 0 {
					configSet = append(configSet, setNameAddVrrp+" advertise-interval "+
						strconv.Itoa(vrrpGroupMap["advertise_interval"].(int)))
				}
				if vrrpGroupMap["authentication_key"].(string) != "" {
					configSet = append(configSet, setNameAddVrrp+" authentication-key \""+
						vrrpGroupMap["authentication_key"].(string)+"\"")
				}
				if vrrpGroupMap["authentication_type"].(string) != "" {
					configSet = append(configSet, setNameAddVrrp+" authentication-type "+
						vrrpGroupMap["authentication_type"].(string))
				}
			case junos.Inet6W:
				setNameAddVrrp = setPrefixAddress + " vrrp-inet6-group " + strconv.Itoa(vrrpGroupMap["identifier"].(int))
				for _, ip := range vrrpGroupMap["virtual_address"].([]interface{}) {
					configSet = append(configSet, setNameAddVrrp+" virtual-inet6-address "+ip.(string))
				}
				configSet = append(configSet, setNameAddVrrp+" virtual-link-local-address "+
					vrrpGroupMap["virtual_link_local_address"].(string))
				if vrrpGroupMap["advertise_interval"].(int) != 0 {
					configSet = append(configSet, setNameAddVrrp+" inet6-advertise-interval "+
						strconv.Itoa(vrrpGroupMap["advertise_interval"].(int)))
				}
			}
			if vrrpGroupMap["accept_data"].(bool) {
				configSet = append(configSet, setNameAddVrrp+" accept-data")
			}
			if vrrpGroupMap["advertisements_threshold"].(int) != 0 {
				configSet = append(configSet, setNameAddVrrp+" advertisements-threshold "+
					strconv.Itoa(vrrpGroupMap["advertisements_threshold"].(int)))
			}
			if vrrpGroupMap["no_accept_data"].(bool) {
				configSet = append(configSet, setNameAddVrrp+" no-accept-data")
			}
			if vrrpGroupMap["no_preempt"].(bool) {
				configSet = append(configSet, setNameAddVrrp+" no-preempt")
			}
			if vrrpGroupMap["preempt"].(bool) {
				configSet = append(configSet, setNameAddVrrp+" preempt")
			}
			if vrrpGroupMap["priority"].(int) != 0 {
				configSet = append(configSet, setNameAddVrrp+" priority "+strconv.Itoa(vrrpGroupMap["priority"].(int)))
			}
			trackInterfaceList := make([]string, 0)
			for _, trackInterface := range vrrpGroupMap["track_interface"].([]interface{}) {
				trackInterfaceMap := trackInterface.(map[string]interface{})
				if bchk.InSlice(trackInterfaceMap["interface"].(string), trackInterfaceList) {
					return configSet, fmt.Errorf("multiple blocks track_interface with the same interface %s",
						trackInterfaceMap["interface"].(string))
				}
				trackInterfaceList = append(trackInterfaceList, trackInterfaceMap["interface"].(string))
				configSet = append(configSet, setNameAddVrrp+" track interface "+trackInterfaceMap["interface"].(string)+
					" priority-cost "+strconv.Itoa(trackInterfaceMap["priority_cost"].(int)))
			}
			trackRouteList := make([]string, 0)
			for _, trackRoute := range vrrpGroupMap["track_route"].([]interface{}) {
				trackRouteMap := trackRoute.(map[string]interface{})
				if bchk.InSlice(trackRouteMap["route"].(string), trackRouteList) {
					return configSet, fmt.Errorf("multiple blocks track_route with the same interface %s",
						trackRouteMap["route"].(string))
				}
				trackRouteList = append(trackRouteList, trackRouteMap["route"].(string))
				configSet = append(configSet, setNameAddVrrp+" track route "+trackRouteMap["route"].(string)+
					" routing-instance "+trackRouteMap["routing_instance"].(string)+
					" priority-cost "+strconv.Itoa(trackRouteMap["priority_cost"].(int)))
			}
		}
	}

	return configSet, nil
}

func setFamilyInetDhcp(dhcp map[string]interface{}, setPrefixInt string) []string {
	configSet := make([]string, 0)
	setPrefix := setPrefixInt + "family inet dhcp "
	if dhcp["srx_old_option_name"].(bool) {
		setPrefix = setPrefixInt + "family inet dhcp-client "
	}

	configSet = append(configSet, setPrefix)
	if v := dhcp["client_identifier_ascii"].(string); v != "" {
		configSet = append(configSet, setPrefix+"client-identifier ascii \""+v+"\"")
	}
	if v := dhcp["client_identifier_hexadecimal"].(string); v != "" {
		configSet = append(configSet, setPrefix+"client-identifier hexadecimal "+v)
	}
	if dhcp["client_identifier_prefix_hostname"].(bool) {
		configSet = append(configSet, setPrefix+"client-identifier prefix host-name")
	}
	if dhcp["client_identifier_prefix_routing_instance_name"].(bool) {
		configSet = append(configSet, setPrefix+"client-identifier prefix routing-instance-name")
	}
	if v := dhcp["client_identifier_use_interface_description"].(string); v != "" {
		configSet = append(configSet, setPrefix+"client-identifier use-interface-description "+v)
	}
	if v := dhcp["client_identifier_userid_ascii"].(string); v != "" {
		configSet = append(configSet, setPrefix+"client-identifier user-id ascii \""+v+"\"")
	}
	if v := dhcp["client_identifier_userid_hexadecimal"].(string); v != "" {
		configSet = append(configSet, setPrefix+"client-identifier user-id hexadecimal "+v)
	}
	if dhcp["force_discover"].(bool) {
		configSet = append(configSet, setPrefix+"force-discover")
	}
	if v := dhcp["lease_time"].(int); v != 0 {
		configSet = append(configSet, setPrefix+"lease-time "+strconv.Itoa(v))
	}
	if dhcp["lease_time_infinite"].(bool) {
		configSet = append(configSet, setPrefix+"lease-time infinite")
	}
	if v := dhcp["metric"].(int); v != -1 {
		configSet = append(configSet, setPrefix+"metric "+strconv.Itoa(v))
	}
	if dhcp["no_dns_install"].(bool) {
		configSet = append(configSet, setPrefix+"no-dns-install")
	}
	if dhcp["options_no_hostname"].(bool) {
		configSet = append(configSet, setPrefix+"options no-hostname")
	}
	if v := dhcp["retransmission_attempt"].(int); v != -1 {
		configSet = append(configSet, setPrefix+"retransmission-attempt "+strconv.Itoa(v))
	}
	if v := dhcp["retransmission_interval"].(int); v != 0 {
		configSet = append(configSet, setPrefix+"retransmission-interval "+strconv.Itoa(v))
	}
	if v := dhcp["server_address"].(string); v != "" {
		configSet = append(configSet, setPrefix+"server-address "+v)
	}
	if dhcp["update_server"].(bool) {
		configSet = append(configSet, setPrefix+"update-server")
	}
	if v := dhcp["vendor_id"].(string); v != "" {
		configSet = append(configSet, setPrefix+"vendor-id \""+v+"\"")
	}

	return configSet
}

func setFamilyInet6Dhcpv6Client(dhcp map[string]interface{}, setPrefixInt string) []string {
	configSet := make([]string, 0)
	setPrefix := setPrefixInt + "family inet6 dhcpv6-client "

	configSet = append(configSet, setPrefix+"client-identifier duid-type "+dhcp["client_identifier_duid_type"].(string))
	configSet = append(configSet, setPrefix+"client-type "+dhcp["client_type"].(string))
	if dhcp["client_ia_type_na"].(bool) {
		configSet = append(configSet, setPrefix+"client-ia-type ia-na")
	}
	if dhcp["client_ia_type_pd"].(bool) {
		configSet = append(configSet, setPrefix+"client-ia-type ia-pd")
	}
	if dhcp["no_dns_install"].(bool) {
		configSet = append(configSet, setPrefix+"no-dns-install")
	}
	if v := dhcp["prefix_delegating_preferred_prefix_length"].(int); v != -1 {
		configSet = append(configSet, setPrefix+"prefix-delegating preferred-prefix-length "+strconv.Itoa(v))
	}
	if v := dhcp["prefix_delegating_sub_prefix_length"].(int); v != 0 {
		configSet = append(configSet, setPrefix+"prefix-delegating sub-prefix-length "+strconv.Itoa(v))
	}
	if dhcp["rapid_commit"].(bool) {
		configSet = append(configSet, setPrefix+"rapid-commit")
	}
	for _, v := range sortSetOfString(dhcp["req_option"].(*schema.Set).List()) {
		configSet = append(configSet, setPrefix+"req-option "+v)
	}
	if v := dhcp["retransmission_attempt"].(int); v != -1 {
		configSet = append(configSet, setPrefix+"retransmission-attempt "+strconv.Itoa(v))
	}
	for _, v := range sortSetOfString(dhcp["update_router_advertisement_interface"].(*schema.Set).List()) {
		configSet = append(configSet, setPrefix+"update-router-advertisement interface "+v)
	}
	if dhcp["update_server"].(bool) {
		configSet = append(configSet, setPrefix+"update-server")
	}

	return configSet
}

func genFamilyInetAddress(address string) map[string]interface{} {
	return map[string]interface{}{
		"cidr_ip":    address,
		"primary":    false,
		"preferred":  false,
		"vrrp_group": make([]map[string]interface{}, 0),
	}
}

func genVRRPGroup(family string) map[string]interface{} {
	vrrpGroup := map[string]interface{}{
		"identifier":               0,
		"virtual_address":          make([]string, 0),
		"accept_data":              false,
		"advertise_interval":       0,
		"advertisements_threshold": 0,
		"no_accept_data":           false,
		"no_preempt":               false,
		"preempt":                  false,
		"priority":                 0,
		"track_interface":          make([]map[string]interface{}, 0),
		"track_route":              make([]map[string]interface{}, 0),
	}
	if family == junos.InetW {
		vrrpGroup["authentication_key"] = ""
		vrrpGroup["authentication_type"] = ""
	}
	if family == junos.Inet6W {
		vrrpGroup["virtual_link_local_address"] = ""
	}

	return vrrpGroup
}
