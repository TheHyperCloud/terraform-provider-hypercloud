package hypercloud

import (
	"fmt"
	"strings"

	hcc "github.com/TheHyperCloud/hypercloud-go-client/hypercloud"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		//Credentials in format <access_key>:<secret_key> or just <access_token>
		Schema: map[string]*schema.Schema{
			"credentials": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"HC_CREDENTIALS"}, nil),
				Description: "The access token for the specified hypercloud account",
			},
			"base_url": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"HC_BASE_URL"}, nil),
				Description:  "The URL endpoint to access the hypercloud API",
				ValidateFunc: validateBaseURL,
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			//"hypercloud_disk_performance_tier" : dataSourceHypercloudDiskPerformanceTier(),
			//"hypercloud_instance_performance_tier" : dataSourceHypercloudInstancePerformanceTier(),
			"hypercloud_region": dataSourceHypercloudRegion(),
			//"hypercloud_template" : datasourceHypercloudTemplate(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"hypercloud_instance": resourceHypercloudInstance(),
			//"hypercloud_disk" : resourceHypercloudDisk(),
			//"hypercloud_network" : resourceHypercloudNetwork(),
			//"hypercloud_ip_address" : resourceHypercloudIPAddress(),
			//"hypercloud_public_key" : resourceHypercloudPublicKey(),
		},

		ConfigureFunc: initHyperCloud,
	}
}

func initHyperCloud(d *schema.ResourceData) (hc interface{}, err error) {
	auth := d.Get("credentials").(string)
	var mErr []error
	hc, mErr = hcc.NewHypercloud(d.Get("base_url").(string), auth)
	if mErr != nil {
		err = fmt.Errorf("%v", mErr)
	} else {
		err = nil
	}
	return
}

func validateBaseURL(v interface{}, k string) (warnings []string, errors []error) {
	url := v.(string)
	if len(url) == 0 {
		errors = append(errors, fmt.Errorf("Fatal: Base URL cannot be empty"))
		return
	}
	if !strings.HasPrefix(url, "https") {
		warnings = append(warnings, fmt.Sprintf("Warning: %s is not using SSL (potentially unsafe operation)", url))
	}
	return
}
