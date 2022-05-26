package junos

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type communityOptions struct {
	invertMatch bool
	name        string
	members     []string
}

func resourcePolicyoptionsCommunity() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePolicyoptionsCommunityCreate,
		ReadWithoutTimeout:   resourcePolicyoptionsCommunityRead,
		UpdateWithoutTimeout: resourcePolicyoptionsCommunityUpdate,
		DeleteWithoutTimeout: resourcePolicyoptionsCommunityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePolicyoptionsCommunityImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: validateNameObjectJunos([]string{}, 64, formatDefault),
			},
			"members": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"invert_match": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourcePolicyoptionsCommunityCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	if sess.junosFakeCreateSetFile != "" {
		if err := setPolicyoptionsCommunity(d, sess, nil); err != nil {
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
	if err := sess.configLock(ctx, junSess); err != nil {
		return diag.FromErr(err)
	}
	var diagWarns diag.Diagnostics
	policyoptsCommunityExists, err := checkPolicyoptionsCommunityExists(d.Get("name").(string), sess, junSess)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	if policyoptsCommunityExists {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns,
			diag.FromErr(fmt.Errorf("policy-options community %v already exists", d.Get("name").(string)))...)
	}

	if err := setPolicyoptionsCommunity(d, sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("create resource junos_policyoptions_community", junSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	policyoptsCommunityExists, err = checkPolicyoptionsCommunityExists(d.Get("name").(string), sess, junSess)
	if err != nil {
		return append(diagWarns, diag.FromErr(err)...)
	}
	if policyoptsCommunityExists {
		d.SetId(d.Get("name").(string))
	} else {
		return append(diagWarns, diag.FromErr(fmt.Errorf("policy-options community %v not exists after commit "+
			"=> check your config", d.Get("name").(string)))...)
	}

	return append(diagWarns, resourcePolicyoptionsCommunityReadWJunSess(d, sess, junSess)...)
}

func resourcePolicyoptionsCommunityRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer sess.closeSession(junSess)

	return resourcePolicyoptionsCommunityReadWJunSess(d, sess, junSess)
}

func resourcePolicyoptionsCommunityReadWJunSess(d *schema.ResourceData, sess *Session, junSess *junosSession,
) diag.Diagnostics {
	mutex.Lock()
	communityOptions, err := readPolicyoptionsCommunity(d.Get("name").(string), sess, junSess)
	mutex.Unlock()
	if err != nil {
		return diag.FromErr(err)
	}
	if communityOptions.name == "" {
		d.SetId("")
	} else {
		fillPolicyoptionsCommunityData(d, communityOptions)
	}

	return nil
}

func resourcePolicyoptionsCommunityUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.Partial(true)
	sess := m.(*Session)
	if sess.junosFakeUpdateAlso {
		if err := delPolicyoptionsCommunity(d.Get("name").(string), sess, nil); err != nil {
			return diag.FromErr(err)
		}
		if err := setPolicyoptionsCommunity(d, sess, nil); err != nil {
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
	if err := delPolicyoptionsCommunity(d.Get("name").(string), sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	if err := setPolicyoptionsCommunity(d, sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("update resource junos_policyoptions_community", junSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	d.Partial(false)

	return append(diagWarns, resourcePolicyoptionsCommunityReadWJunSess(d, sess, junSess)...)
}

func resourcePolicyoptionsCommunityDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sess := m.(*Session)
	if sess.junosFakeDeleteAlso {
		if err := delPolicyoptionsCommunity(d.Get("name").(string), sess, nil); err != nil {
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
	if err := delPolicyoptionsCommunity(d.Get("name").(string), sess, junSess); err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}
	warns, err := sess.commitConf("delete resource junos_policyoptions_community", junSess)
	appendDiagWarns(&diagWarns, warns)
	if err != nil {
		appendDiagWarns(&diagWarns, sess.configClear(junSess))

		return append(diagWarns, diag.FromErr(err)...)
	}

	return diagWarns
}

func resourcePolicyoptionsCommunityImport(ctx context.Context, d *schema.ResourceData, m interface{},
) ([]*schema.ResourceData, error) {
	sess := m.(*Session)
	junSess, err := sess.startNewSession(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.closeSession(junSess)
	result := make([]*schema.ResourceData, 1)

	policyoptsCommunityExists, err := checkPolicyoptionsCommunityExists(d.Id(), sess, junSess)
	if err != nil {
		return nil, err
	}
	if !policyoptsCommunityExists {
		return nil, fmt.Errorf("don't find policy-options community with id '%v' (id must be <name>)", d.Id())
	}
	communityOptions, err := readPolicyoptionsCommunity(d.Id(), sess, junSess)
	if err != nil {
		return nil, err
	}
	fillPolicyoptionsCommunityData(d, communityOptions)

	result[0] = d

	return result, nil
}

func checkPolicyoptionsCommunityExists(name string, sess *Session, junSess *junosSession) (bool, error) {
	showConfig, err := sess.command(cmdShowConfig+"policy-options community "+name+pipeDisplaySet, junSess)
	if err != nil {
		return false, err
	}
	if showConfig == emptyW {
		return false, nil
	}

	return true, nil
}

func setPolicyoptionsCommunity(d *schema.ResourceData, sess *Session, junSess *junosSession) error {
	configSet := make([]string, 0)

	setPrefix := "set policy-options community " + d.Get("name").(string) + " "
	for _, v := range d.Get("members").([]interface{}) {
		configSet = append(configSet, setPrefix+"members "+v.(string))
	}
	if d.Get("invert_match").(bool) {
		configSet = append(configSet, setPrefix+"invert-match")
	}

	return sess.configSet(configSet, junSess)
}

func readPolicyoptionsCommunity(name string, sess *Session, junSess *junosSession) (communityOptions, error) {
	var confRead communityOptions

	showConfig, err := sess.command(cmdShowConfig+
		"policy-options community "+name+pipeDisplaySetRelative, junSess)
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
			switch {
			case strings.HasPrefix(itemTrim, "members "):
				confRead.members = append(confRead.members, strings.TrimPrefix(itemTrim, "members "))
			case itemTrim == "invert-match":
				confRead.invertMatch = true
			}
		}
	}

	return confRead, nil
}

func delPolicyoptionsCommunity(community string, sess *Session, junSess *junosSession) error {
	configSet := make([]string, 0, 1)
	configSet = append(configSet, "delete policy-options community "+community)

	return sess.configSet(configSet, junSess)
}

func fillPolicyoptionsCommunityData(d *schema.ResourceData, communityOptions communityOptions) {
	if tfErr := d.Set("name", communityOptions.name); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("members", communityOptions.members); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("invert_match", communityOptions.invertMatch); tfErr != nil {
		panic(tfErr)
	}
}
