package junos

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type screenWhiteListOptions struct {
	name    string
	address []string
}

func resourceSecurityScreenWhiteList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityScreenWhiteListCreate,
		ReadWithoutTimeout:   resourceSecurityScreenWhiteListRead,
		UpdateWithoutTimeout: resourceSecurityScreenWhiteListUpdate,
		DeleteWithoutTimeout: resourceSecurityScreenWhiteListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceSecurityScreenWhiteListImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 32, formatDefault),
			},
			"address": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validateCIDRNetworkFunc(),
				},
			},
		},
	}
}

func resourceSecurityScreenWhiteListCreate(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	sess := m.(*Session)
	if sess.junosFakeCreateSetFile != "" {
		if err := setSecurityScreenWhiteList(d, sess, nil); err != nil {
			return diag.FromErr(err)
		}
		d.SetId(d.Get("name").(string))

		return nil
	}
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(junSess)
	if !checkCompatibilitySecurity(junSess) {
		return diag.FromErr(fmt.Errorf("security screen white-list not compatible with Junos device %s",
			junSess.SystemInformation.HardwareModel))
	}
	if err := sess.configLock(ctx, junSess); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	securityScreenWhiteListExists, err := checkSecurityScreenWhiteListExists(d.Get("name").(string), sess, junSess)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	if securityScreenWhiteListExists {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns,
			diag.FromErr(fmt.Errorf("security screen white-list %v already exists", d.Get("name").(string)))...)
	}

	if err := setSecurityScreenWhiteList(d, sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("create resource junos_security_screen_whitelist", junSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	securityScreenWhiteListExists, err = checkSecurityScreenWhiteListExists(d.Get("name").(string), sess, junSess)
	if err != nil {
		return append(diagWarns, diag.FromErr(err)...)
	}
	if securityScreenWhiteListExists {
		d.SetId(d.Get("name").(string))
	} else {
		return append(diagWarns, diag.FromErr(fmt.Errorf("security screen white-list %v not exists after commit "+
			"=> check your config", d.Get("name").(string)))...)
	}

	return append(diagWarns, resourceSecurityScreenWhiteListReadWJunSess(d, sess, junSess)...)
}

func resourceSecurityScreenWhiteListRead(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	sess := m.(*Session)
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(junSess)

	return resourceSecurityScreenWhiteListReadWJunSess(d, sess, junSess)
}

func resourceSecurityScreenWhiteListReadWJunSess(d *schema.ResourceData, sess *Session, junSess *junosSession,
) diag.Diagnostics {
	mutex.Lock()
	whiteListOptions, err := readSecurityScreenWhiteList(d.Get("name").(string), sess, junSess)
	mutex.Unlock()
	if err != nil {
		return diag.FromErr(err)
	}
	if whiteListOptions.name == "" {
		d.SetId("")
	} else {
		fillSecurityScreenWhiteListData(d, whiteListOptions)
	}

	return nil
}

func resourceSecurityScreenWhiteListUpdate(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	d.Partial(true)
	sess := m.(*Session)
	if sess.junosFakeUpdateAlso {
		if err := delSecurityScreenWhiteList(d.Get("name").(string), sess, nil); err != nil {
			return diag.FromErr(err)
		}
		if err := setSecurityScreenWhiteList(d, sess, nil); err != nil {
			return diag.FromErr(err)
		}
		d.Partial(false)

		return nil
	}
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(junSess)
	if err := sess.configLock(ctx, junSess); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	if err := delSecurityScreenWhiteList(d.Get("name").(string), sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	if err := setSecurityScreenWhiteList(d, sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("update resource junos_security_screen_whitelist", junSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	d.Partial(false)

	return append(diagWarns, resourceSecurityScreenWhiteListReadWJunSess(d, sess, junSess)...)
}

func resourceSecurityScreenWhiteListDelete(ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	sess := m.(*Session)
	if sess.junosFakeDeleteAlso {
		if err := delSecurityScreenWhiteList(d.Get("name").(string), sess, nil); err != nil {
			return diag.FromErr(err)
		}

		return nil
	}
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(junSess)
	if err := sess.configLock(ctx, junSess); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	if err := delSecurityScreenWhiteList(d.Get("name").(string), sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("delete resource junos_security_screen_whitelist", junSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}

	return diagWarns
}

func resourceSecurityScreenWhiteListImport(ctx context.Context, d *schema.ResourceData, m interface{},
) ([]*schema.ResourceData, error) {
	sess := m.(*Session)
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.closeSession(junSess)
	result := make([]*schema.ResourceData, 1)
	securityScreenWhiteListExists, err := checkSecurityScreenWhiteListExists(d.Id(), sess, junSess)
	if err != nil {
		return nil, err
	}
	if !securityScreenWhiteListExists {
		return nil, fmt.Errorf("don't find screen white-list with id '%v' (id must be <name>)", d.Id())
	}
	whiteListOptions, err := readSecurityScreenWhiteList(d.Id(), sess, junSess)
	if err != nil {
		return nil, err
	}
	fillSecurityScreenWhiteListData(d, whiteListOptions)

	result[0] = d

	return result, nil
}

func checkSecurityScreenWhiteListExists(name string, sess *Session, junSess *junosSession) (bool, error) {
	showConfig, err := sess.command(cmdShowConfig+"security screen white-list "+name+pipeDisplaySet, junSess)
	if err != nil {
		return false, err
	}
	if showConfig == emptyW {
		return false, nil
	}

	return true, nil
}

func setSecurityScreenWhiteList(d *schema.ResourceData, sess *Session, junSess *junosSession) error {
	configSet := make([]string, 0)

	setPrefix := "set security screen white-list " + d.Get("name").(string) + " "

	for _, v := range sortSetOfString(d.Get("address").(*schema.Set).List()) {
		configSet = append(configSet, setPrefix+"address "+v)
	}

	return sess.configSet(configSet, junSess)
}

func readSecurityScreenWhiteList(name string, sess *Session, junSess *junosSession) (screenWhiteListOptions, error) {
	var confRead screenWhiteListOptions

	showConfig, err := sess.command(cmdShowConfig+
		"security screen white-list "+name+pipeDisplaySetRelative, junSess)
	if err != nil {
		return confRead, err
	}
	if showConfig != emptyW {
		confRead.name = name
		for _, item := range strings.Split(showConfig, "\n") {
			if strings.Contains(item, xmlStartTagConfigOut) {
				continue
			}
			if strings.Contains(item, xmlEndTagConfigOut) {
				break
			}
			itemTrim := strings.TrimPrefix(item, setLS)
			if strings.HasPrefix(itemTrim, "address ") {
				confRead.address = append(confRead.address, strings.TrimPrefix(itemTrim, "address "))
			}
		}
	}

	return confRead, nil
}

func delSecurityScreenWhiteList(name string, sess *Session, junSess *junosSession) error {
	configSet := make([]string, 0, 1)
	configSet = append(configSet, "delete security screen white-list "+name)

	return sess.configSet(configSet, junSess)
}

func fillSecurityScreenWhiteListData(d *schema.ResourceData, whiteListOptions screenWhiteListOptions) {
	if tfErr := d.Set("name", whiteListOptions.name); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("address", whiteListOptions.address); tfErr != nil {
		panic(tfErr)
	}
}
