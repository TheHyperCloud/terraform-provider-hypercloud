package hypercloud

import (
	"fmt"

	//Hypercloud-client. Replace with actual repo on git
	hcc "bitbucket.org/mistarhee/hypercloud-go-client/hypercloud"
	"github.com/hashicorp/terraform/helper/schema"
)

//Name, code, id

func dataSourceHypercloudRegion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceHypercloudRegionRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"code": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceHypercloudRegionRead(d *schema.ResourceData, meta interface{}) error {
	hc := hcc.ToHypercloud(meta)
	id := d.Get("id").(string)

	region, err := hc.RegionInfo(id)
	if err != nil {
		return fmt.Errorf("Unable to get regions: \n%v", err)
	}
	d.Set("name", region.(map[string]interface{})["name"].(string))
	d.Set("code", region.(map[string]interface{})["code"].(string))
	return nil
}
