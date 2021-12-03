package vault

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-vault/util"
	"github.com/hashicorp/vault/api"
)

var (
	awsAuthBackendConfigIdentityBackendFromPathRegex = regexp.MustCompile("^auth/(.+)/config/identity$")
)

func awsAuthBackendConfigIdentityResource() *schema.Resource {
	return &schema.Resource{
		Create: awsAuthBackendConfigIdentityWrite,
		Update: awsAuthBackendConfigIdentityWrite,
		Read:   awsAuthBackendConfigIdentityRead,
		Delete: awsAuthBackendConfigIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"iam_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "role_id",
				Description:  "How to generate the identity alias when using the iam auth method.",
				ValidateFunc: validation.StringInSlice([]string{"role_id", "unique_id", "full_arn"}, false),
			},
			"iam_metadata": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The metadata to include on the token returned by the login endpoint.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ec2_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Configures how to generate the identity alias when using the ec2 auth method.",
				Default:      "role_id",
				ValidateFunc: validation.StringInSlice([]string{"role_id", "instance_id", "image_id"}, false),
			},
			"ec2_metadata": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The metadata to include on the token returned by the login endpoint.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"backend": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Unique name of the auth backend to configure.",
				ForceNew:    true,
				Default:     "aws",
				// standardise on no beginning or trailing slashes
				StateFunc: func(v interface{}) string {
					return strings.Trim(v.(string), "/")
				},
			},
		},
	}
}

func awsAuthBackendConfigIdentityWrite(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	var iamMetadata, ec2Metadata []string
	backend := d.Get("backend").(string)
	iamAlias := d.Get("iam_alias").(string)
	ec2Alias := d.Get("ec2_alias").(string)

	if iamMetadataConfig, ok := d.GetOk("iam_metadata"); ok {
		iamMetadata = util.TerraformSetToStringArray(iamMetadataConfig)
	}

	if ec2MetadataConfig, ok := d.GetOk("ec2_metadata"); ok {
		ec2Metadata = util.TerraformSetToStringArray(ec2MetadataConfig)
	}

	path := awsAuthBackendConfigIdentityPath(backend)
	data := map[string]interface{}{
		"iam_alias":    iamAlias,
		"iam_metadata": iamMetadata,
		"ec2_alias":    ec2Alias,
		"ec2_metadata": ec2Metadata,
	}

	log.Printf("[DEBUG] Writing AWS identity config to %q", path)
	_, err := client.Logical().Write(path, data)

	if err != nil {
		return fmt.Errorf("error configuring AWS auth identity config %q: %s", path, err)
	}
	d.SetId(path)

	log.Printf("[DEBUG] Wrote AWS identity config to %q", path)

	return awsAuthBackendConfigIdentityRead(d, meta)
}

func awsAuthBackendConfigIdentityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)

	path := d.Id()

	backend, err := awsAuthBackendConfigIdentityBackendFromPath(path)
	if err != nil {
		return fmt.Errorf("invalid path %q for AWS auth identity config:  %s", path, err)
	}

	log.Printf("[DEBUG] Reading identity config %q from AWS auth backend", path)
	resp, err := client.Logical().Read(path)
	if err != nil {
		return fmt.Errorf("error reading AWS auth backend identity config %q: %s", path, err)
	}
	log.Printf("[DEBUG] Read identity config %q from AWS auth backend", path)
	if resp == nil {
		log.Printf("[WARN] AWS auth backend identity config %q not found, removing it from state", path)
		d.SetId("")
		return nil
	}

	d.Set("iam_alias", resp.Data["iam_alias"])
	d.Set("iam_metadata", resp.Data["iam_metadata"])
	d.Set("ec2_alias", resp.Data["ec2_alias"])
	d.Set("ec2_metadata", resp.Data["ec2_metadata"])
	d.Set("backend", backend)

	return nil
}

func awsAuthBackendConfigIdentityDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting AWS identity config from state file")
	return nil
}

func awsAuthBackendConfigIdentityExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*api.Client)

	path := d.Id()

	log.Printf("[DEBUG] Checking if identity config %q exists in AWS auth backend", path)
	resp, err := client.Logical().Read(path)
	if err != nil {
		return true, fmt.Errorf("error checking for existence of AWS auth backend identity config %q: %s", path, err)
	}
	log.Printf("[DEBUG] Checked if identity config %q exists in AWS auth backend", path)
	return resp != nil, nil
}

func awsAuthBackendConfigIdentityPath(backend string) string {
	return "auth/" + strings.Trim(backend, "/") + "/config/identity"
}

func awsAuthBackendConfigIdentityBackendFromPath(path string) (string, error) {
	if !awsAuthBackendConfigIdentityBackendFromPathRegex.MatchString(path) {
		return "", fmt.Errorf("no backend found")
	}
	res := awsAuthBackendConfigIdentityBackendFromPathRegex.FindStringSubmatch(path)
	if len(res) != 2 {
		return "", fmt.Errorf("unexpected number of matches (%d) for backend", len(res))
	}
	return res[1], nil
}