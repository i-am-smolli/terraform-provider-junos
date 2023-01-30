package providersdk

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jeremmfr/terraform-provider-junos/internal/junos"
)

type policyPairPolicyOptions struct {
	zoneA      string
	zoneB      string
	policyAtoB string
	policyBtoA string
}

func resourceSecurityPolicyTunnelPairPolicy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityPolicyTunnelPairPolicyCreate,
		ReadWithoutTimeout:   resourceSecurityPolicyTunnelPairPolicyRead,
		DeleteWithoutTimeout: resourceSecurityPolicyTunnelPairPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSecurityPolicyTunnelPairPolicyImport,
		},
		Schema: map[string]*schema.Schema{
			"zone_a": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"zone_b": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"policy_a_to_b": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"policy_b_to_a": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
		},
	}
}

func resourceSecurityPolicyTunnelPairPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	clt := m.(*junos.Client)
	if clt.FakeCreateSetFile() {
		junSess := clt.NewSessionWithoutNetconf(ctx)
		if err := setSecurityPolicyTunnelPairPolicy(d, junSess); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(d.Get("zone_a").(string) + junos.IDSeparator + d.Get("policy_a_to_b").(string) +
			junos.IDSeparator + d.Get("zone_b").(string) + junos.IDSeparator + d.Get("policy_b_to_a").(string))

		return nil
	}
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer junSess.Close()
	if !junSess.CheckCompatibilitySecurity() {
		return diag.FromErr(fmt.Errorf("security policy tunnel pair policy not compatible with Junos device %s",
			junSess.SystemInformation.HardwareModel))
	}
	if err := junSess.ConfigLock(ctx); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	securityPolicyExists, err := checkSecurityPolicyExists(
		d.Get("zone_a").(string),
		d.Get("zone_b").(string),
		junSess,
	)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	if !securityPolicyExists {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(fmt.Errorf("security policy from %v to %v not exists",
			d.Get("zone_a").(string), d.Get("zone_b").(string)))...)
	}
	securityPolicyExists, err = checkSecurityPolicyExists(
		d.Get("zone_b").(string),
		d.Get("zone_a").(string),
		junSess,
	)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	if !securityPolicyExists {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(fmt.Errorf("security policy from %v to %v not exists",
			d.Get("zone_b").(string), d.Get("zone_a").(string)))...)
	}
	pairPolicyExists, err := checkSecurityPolicyPairExists(
		d.Get("zone_a").(string),
		d.Get("policy_a_to_b").(string),
		d.Get("zone_b").(string),
		d.Get("policy_b_to_a").(string),
		junSess,
	)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	if pairPolicyExists {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(fmt.Errorf("security policy pair policy %v(%v) / %v(%v) already exists",
			d.Get("zone_a").(string), d.Get("policy_a_to_b").(string),
			d.Get("zone_b").(string), d.Get("policy_b_to_a").(string)))...)
	}
	if err := setSecurityPolicyTunnelPairPolicy(d, junSess); err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := junSess.CommitConf("create resource junos_security_policy_tunnel_pair_policy")
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	pairPolicyExists, err = checkSecurityPolicyPairExists(
		d.Get("zone_a").(string),
		d.Get("policy_a_to_b").(string),
		d.Get("zone_b").(string),
		d.Get("policy_b_to_a").(string),
		junSess,
	)
	if err != nil {
		return append(diagWarns, diag.FromErr(err)...)
	}
	if pairPolicyExists {
		d.SetId(d.Get("zone_a").(string) + junos.IDSeparator + d.Get("policy_a_to_b").(string) +
			junos.IDSeparator + d.Get("zone_b").(string) + junos.IDSeparator + d.Get("policy_b_to_a").(string))
	} else {
		return append(diagWarns, diag.FromErr(fmt.Errorf("security policy pair policy not exists after commit "+
			"=> check your config"))...)
	}

	return append(diagWarns, resourceSecurityPolicyTunnelPairPolicyReadWJunSess(d, junSess)...)
}

func resourceSecurityPolicyTunnelPairPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	clt := m.(*junos.Client)
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer junSess.Close()

	return resourceSecurityPolicyTunnelPairPolicyReadWJunSess(d, junSess)
}

func resourceSecurityPolicyTunnelPairPolicyReadWJunSess(
	d *schema.ResourceData, junSess *junos.Session,
) diag.Diagnostics {
	junos.MutexLock()
	policyPairPolicyOptions, err := readSecurityPolicyTunnelPairPolicy(
		d.Get("zone_a").(string),
		d.Get("policy_a_to_b").(string),
		d.Get("zone_b").(string),
		d.Get("policy_b_to_a").(string),
		junSess,
	)
	junos.MutexUnlock()
	if err != nil {
		return diag.FromErr(err)
	}
	if policyPairPolicyOptions.policyAtoB == "" && policyPairPolicyOptions.policyBtoA == "" {
		d.SetId("")
	} else {
		fillSecurityPolicyPairData(d, policyPairPolicyOptions)
	}

	return nil
}

func resourceSecurityPolicyTunnelPairPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	clt := m.(*junos.Client)
	if clt.FakeDeleteAlso() {
		junSess := clt.NewSessionWithoutNetconf(ctx)
		if err := delSecurityPolicyTunnelPairPolicy(d, junSess); err != nil {
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
	if err := delSecurityPolicyTunnelPairPolicy(d, junSess); err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := junSess.CommitConf("delete resource junos_security_policy_tunnel_pair_policy")
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, junSess.ConfigClear())

		return append(diagWarns, diag.FromErr(err)...)
	}

	return diagWarns
}

func resourceSecurityPolicyTunnelPairPolicyImport(ctx context.Context, d *schema.ResourceData, m interface{},
) ([]*schema.ResourceData, error) {
	clt := m.(*junos.Client)
	junSess, err := clt.StartNewSession(ctx)
	if err != nil {
		return nil, err
	}
	defer junSess.Close()
	result := make([]*schema.ResourceData, 1)
	idList := strings.Split(d.Id(), junos.IDSeparator)
	if len(idList) < 4 {
		return nil, fmt.Errorf("missing element(s) in id with separator %v", junos.IDSeparator)
	}
	poliyPairPolicyExists, err := checkSecurityPolicyPairExists(
		idList[0],
		idList[1],
		idList[2],
		idList[3],
		junSess,
	)
	if err != nil {
		return nil, err
	}
	if !poliyPairPolicyExists {
		return nil, fmt.Errorf("don't find policy pair policy with id %v "+
			"(id must be <zone_a>"+junos.IDSeparator+"<policy_a_to_b>"+junos.IDSeparator+
			"<zone_b>"+junos.IDSeparator+"<policy_b_to_a>)", d.Id())
	}

	policyPairPolicyOptions, err := readSecurityPolicyTunnelPairPolicy(
		idList[0],
		idList[1],
		idList[2],
		idList[3],
		junSess,
	)
	if err != nil {
		return nil, err
	}
	fillSecurityPolicyPairData(d, policyPairPolicyOptions)

	result[0] = d

	return result, nil
}

func checkSecurityPolicyPairExists(
	zoneA, policyAtoB, zoneB, policyBtoA string, junSess *junos.Session,
) (
	bool, error,
) {
	showConfigPairAtoB, err := junSess.Command(junos.CmdShowConfig +
		"security policies from-zone " + zoneA + " to-zone " + zoneB + " policy " + policyAtoB +
		" then permit tunnel pair-policy" + junos.PipeDisplaySet)
	if err != nil {
		return false, err
	}
	showConfigPairBtoA, err := junSess.Command(junos.CmdShowConfig +
		"security policies from-zone " + zoneB + " to-zone " + zoneA + " policy " + policyBtoA +
		" then permit tunnel pair-policy" + junos.PipeDisplaySet)
	if err != nil {
		return false, err
	}
	if showConfigPairAtoB == junos.EmptyW && showConfigPairBtoA == junos.EmptyW {
		return false, nil
	}

	return true, nil
}

func setSecurityPolicyTunnelPairPolicy(d *schema.ResourceData, junSess *junos.Session) error {
	configSet := make([]string, 0, 2)

	configSet = append(configSet, "set security policies from-zone "+
		d.Get("zone_a").(string)+" to-zone "+d.Get("zone_b").(string)+
		" policy "+d.Get("policy_a_to_b").(string)+
		" then permit tunnel pair-policy "+d.Get("policy_b_to_a").(string))
	configSet = append(configSet, "set security policies from-zone "+
		d.Get("zone_b").(string)+" to-zone "+d.Get("zone_a").(string)+
		" policy "+d.Get("policy_b_to_a").(string)+
		" then permit tunnel pair-policy "+d.Get("policy_a_to_b").(string))

	return junSess.ConfigSet(configSet)
}

func readSecurityPolicyTunnelPairPolicy(
	zoneA, policyAtoB, zoneB, policyBtoA string, junSess *junos.Session,
) (
	confRead policyPairPolicyOptions, err error,
) {
	showConfig, err := junSess.Command(junos.CmdShowConfig +
		"security policies from-zone " + zoneA + " to-zone " + zoneB + " policy " + policyAtoB +
		" then permit tunnel pair-policy" + junos.PipeDisplaySet)
	if err != nil {
		return confRead, err
	}
	if showConfig != junos.EmptyW {
		confRead.zoneA = zoneA
		confRead.zoneB = zoneB
		for _, item := range strings.Split(showConfig, "\n") {
			if strings.Contains(item, junos.XMLStartTagConfigOut) {
				continue
			}
			if strings.Contains(item, junos.XMLEndTagConfigOut) {
				break
			}
			if strings.Contains(item, " tunnel pair-policy ") {
				confRead.policyBtoA = strings.TrimPrefix(item,
					"set security policies from-zone "+zoneA+" to-zone "+zoneB+
						" policy "+policyAtoB+" then permit tunnel pair-policy ")
			}
		}
	}
	showConfig, err = junSess.Command(junos.CmdShowConfig +
		"security policies from-zone " + zoneB + " to-zone " + zoneA + " policy " + policyBtoA +
		" then permit tunnel pair-policy" + junos.PipeDisplaySet)
	if err != nil {
		return confRead, err
	}
	if showConfig != junos.EmptyW {
		confRead.zoneA = zoneA
		confRead.zoneB = zoneB
		for _, item := range strings.Split(showConfig, "\n") {
			if strings.Contains(item, junos.XMLStartTagConfigOut) {
				continue
			}
			if strings.Contains(item, junos.XMLEndTagConfigOut) {
				break
			}
			if strings.Contains(item, " tunnel pair-policy ") {
				confRead.policyAtoB = strings.TrimPrefix(item,
					"set security policies from-zone "+zoneB+" to-zone "+zoneA+
						" policy "+policyBtoA+" then permit tunnel pair-policy ")
			}
		}
	}

	return confRead, nil
}

func delSecurityPolicyTunnelPairPolicy(d *schema.ResourceData, junSess *junos.Session) error {
	configSet := make([]string, 0, 2)
	configSet = append(configSet, "delete security policies"+
		" from-zone "+d.Get("zone_a").(string)+" to-zone "+d.Get("zone_b").(string)+
		" policy "+d.Get("policy_a_to_b").(string)+
		" then permit tunnel pair-policy "+d.Get("policy_b_to_a").(string))
	configSet = append(configSet, "delete security policies"+
		" from-zone "+d.Get("zone_b").(string)+" to-zone "+d.Get("zone_a").(string)+
		" policy "+d.Get("policy_b_to_a").(string)+
		" then permit tunnel pair-policy "+d.Get("policy_a_to_b").(string))

	return junSess.ConfigSet(configSet)
}

func fillSecurityPolicyPairData(d *schema.ResourceData, policyPairPolicyOptions policyPairPolicyOptions) {
	if tfErr := d.Set("zone_a", policyPairPolicyOptions.zoneA); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("zone_b", policyPairPolicyOptions.zoneB); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("policy_a_to_b", policyPairPolicyOptions.policyAtoB); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("policy_b_to_a", policyPairPolicyOptions.policyBtoA); tfErr != nil {
		panic(tfErr)
	}
}
