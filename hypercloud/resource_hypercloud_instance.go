/*
   Ref: https://cloud.orionvm.com/developer/v1#instance
*/

package hypercloud

import (
	"fmt"
	"time"

	hcc "bitbucket.org/mistarhee/hypercloud-go-client/hypercloud"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceHypercloudInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceHypercloudInstanceCreate,
		Read:   resourceHypercloudInstanceRead,
		Update: resourceHypercloudInstanceUpdate,
		Delete: resourceHypercloudInstanceDelete,
		Exists: resourceHypercloudInstanceExists,

		SchemaVersion: 1, //For API v1

		Schema: map[string]*schema.Schema{
			"memory": &schema.Schema{
				Type:         schema.TypeInt, //Is float from json, but memory is gonna be of int anyway
				Required:     true,
				ValidateFunc: memoryValidation,
				Description:  "Instance RAM in megabytes. Must be a power of 2",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the instance",
			},
			"performance_tier": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the performance tier to assign to the instance",
			},
			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the region in which to create the instance",
			},
			"availability_group": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 2, //3 max in availiability group
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of the instances that should be grouped together with the instance for high availability",
			},
			"boot_device": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "disk",
				ValidateFunc: validation.StringInSlice([]string{"disk", "cdrom", "network"}, false),
				Description:  "Device from which the instance should boot. One of `disk`, `cdrom`, `network`",
			},
			"disks": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of disks to be attached to the instance, in device order",
			},
			"ip_addresses": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of IP addresses to be assigned to the instance",
			},
			"public_keys": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of public keys that can be used to access the instance",
			},
			"start_on_crash": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to restart the instance after a crash event",
			},
			"start_on_reboot": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to restart the instance after a reboot event",
			},
			"start_on_shutdown": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to restart the instance after a shutdown event",
			},
			"virtualization": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "hvm",
				ValidateFunc: validation.StringInSlice([]string{"disk", "cdrom", "network"}, false),
				Description:  "Virtualization mode. One of `hvm`, `pv`",
			},
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func memoryValidation(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(int)
	if !ok {
		es = append(es, fmt.Errorf("Expected type of %s to be int", k))
		return
	}
	if v < 512 {
		es = append(es, fmt.Errorf("%d ram is below the minimum of 512 MB", v))
	}
	if v == 0 {
		es = append(es, fmt.Errorf("0 is an invalid amount of RAM"))
		return
	}
	for ; v&1 == 0; v = v >> 1 {
	}
	if v != 1 {
		es = append(es, fmt.Errorf("Specified RAM amount %s is not a power of two. Aborting", k))
		return
	}
	return
}

func resourceHypercloudInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	hc := hcc.ToHypercloud(meta)
	requestData := make(map[string]interface{})

	/* Get the essentials (memory/name/performancetier/region) */
	requestData["memory"] = d.Get("memory").(int)
	requestData["name"] = d.Get("name").(string)
	requestData["performance_tier"] = d.Get("performance_tier").(string)
	requestData["region"] = d.Get("region").(string)

	/* Check for the other fields, if they exist, add them */
	ag, exists := d.GetOk("availability_group")
	if exists {
		requestData["availability_group"] = ag.([]string)
	}

	bd, exists := d.GetOk("boot_device")
	if exists {
		requestData["boot_device"] = bd.(string)
	}

	disks, exists := d.GetOk("disks")
	if exists {
		requestData["disks"] = disks.([]string)
	}

	ipAddr, exists := d.GetOk("ip_addresses")
	if exists {
		requestData["ip_addresses"] = ipAddr.([]string)
	}

	startOnCrash, exists := d.GetOk("start_on_crash")
	if exists {
		requestData["start_on_crash"] = startOnCrash.(bool)
	}

	startOnReboot, exists := d.GetOk("start_on_reboot")
	if exists {
		requestData["start_on_reboot"] = startOnReboot.(bool)
	}

	startOnShutdown, exists := d.GetOk("start_on_shutdown")
	if exists {
		requestData["start_on_shutdown"] = startOnShutdown.(bool)
	}

	virtualization, exists := d.GetOk("virtualization")
	if exists {
		requestData["virtualization"] = virtualization.(string)
	}

	createResponse, err := hc.InstanceAssemble(requestData)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	cr := createResponse.(map[string]interface{})
	d.SetId(cr["id"].(string))

	createTimeout, exists := d.GetOk("create_timeout")
	if !exists {
		createTimeout = 60 //1 min timeout might be too long (W H O K N O W S)
	}

	/* Wait until the resource is "ready" i.e. stopped state */
	waitErr := waitInstanceUp(hc, cr["id"].(string), createTimeout.(int))
	if waitErr != nil {
		d.SetId("")
		return waitErr
	}

	return resourceHypercloudInstanceRead(d, meta)
}

func resourceHypercloudInstanceRead(d *schema.ResourceData, meta interface{}) error {
	hc := hcc.ToHypercloud(meta)
	ret, err := hc.InstanceInfo(d.Id())
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	instance := ret.(map[string]interface{})

	/* Lets fill out the form shall we? */
	d.Set("memory", int(instance["memory"].(float64)))
	d.Set("name", instance["name"].(string))
	d.Set("performance_tier", instance["performance_tier"].(map[string]interface{})["id"].(string))
	d.Set("region", instance["region"].(map[string]interface{})["id"].(string))

	/* Alrighty-o time to fill out the rest. */

	//Availability group
	d.Set("availability_group", instance["availability_group"].([]interface{})) //Should actually be []string but IDK what json's doing

	//Boot device
	d.Set("boot_device", instance["boot_device"].(string))

	//Disks
	//This one is trickier. Have to pull out the list of disk IDs
	var mDisks []string
	for _, disk := range instance["disks"].([]interface{}) {
		md := disk.(map[string]interface{})
		mDisks = append(mDisks, md["id"].(string))
	}
	d.Set("disks", mDisks)

	//IP Addresses
	//Similar to disks, need to pull these guys out (wew)
	var mIps []string
	for _, na := range instance["network_adapters"].([]interface{}) {
		neta := na.(map[string]interface{})
		for _, ips := range neta["ip_addresses"].([]interface{}) {
			ip := ips.(map[string]interface{})
			mIps = append(mIps, ip["id"].(string))
		}
	}
	d.Set("ip_addresses", mIps)

	//Public Keys
	//Again, we need to pull out the IDs
	var mPKs []string
	for _, pks := range instance["public_keys"].([]interface{}) {
		pk := pks.(map[string]interface{})
		mPKs = append(mPKs, pk["id"].(string))
	}
	d.Set("public_keys", mPKs)

	//How unnecessarily verbose

	//Start on __x__
	d.Set("start_on_crash", instance["start_on_crash"].(bool))
	d.Set("start_on_reboot", instance["start_on_reboot"].(bool))
	d.Set("start_on_shutdown", instance["start_on_shutdown"].(bool))

	//Virtualization
	d.Set("virtualization", instance["virtualization"].(string))

	//Created At
	d.Set("created_at", instance["created_at"].(string))
	d.Set("updated_at", instance["updated_at"].(string))
	d.Set("instance_id", instance["id"].(string))

	d.SetId(instance["id"].(string))
	//We donezo 8^)
	return nil
}

func resourceHypercloudInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	hc := hcc.ToHypercloud(meta)

	d.Partial(true)
	//Lets start from the top of the round.
	//Memory - (I: 23)
	if d.HasChange("memory") {
		_, n := d.GetChange("memory")
		/* Ensure that the memory is correct */
		_, e := memoryValidation(n, fmt.Sprintf("%d", n))
		if e != nil && len(e) != 0 {
			return fmt.Errorf("%d is an invalid memory amount", d.Get("memory").(int))
		}
		update := make(map[string]interface{})
		update["memory"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			return fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("memory")
	}

	if d.HasChange("name") {
		_, n := d.GetChange("name")
		update := make(map[string]interface{})
		update["name"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("name")
	}

	if d.HasChange("performance_tier") {
		_, n := d.GetChange("performance_tier")
		update := make(map[string]interface{})
		update["name"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("performance_tier")
	}

	if d.HasChange("region") {
		_, n := d.GetChange("region")
		update := make(map[string]interface{})
		update["region"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("region")
	}

	if d.HasChange("availability_group") {
		_, n := d.GetChange("availability_group")
		update := make(map[string]interface{})
		update["availability_group"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("availability_group")
	}

	if d.HasChange("boot_device") {
		_, n := d.GetChange("boot_device")
		update := make(map[string]interface{})
		update["boot_device"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("boot_device")
	}

	if d.HasChange("disks") {
		_, n := d.GetChange("disks")
		update := make(map[string]interface{})
		update["disks"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("disks")
	}

	if d.HasChange("ip_addresses") {
		_, n := d.GetChange("ip_addresses")
		update := make(map[string]interface{})
		update["ip_addresses"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("ip_addresses")
	}

	if d.HasChange("public_keys") {
		_, n := d.GetChange("public_keys")
		update := make(map[string]interface{})
		update["public_keys"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("public_keys")
	}

	if d.HasChange("start_on_crash") {
		_, n := d.GetChange("start_on_crash")
		update := make(map[string]interface{})
		update["start_on_crash"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("start_on_crash")
	}

	if d.HasChange("start_on_reboot") {
		_, n := d.GetChange("start_on_reboot")
		update := make(map[string]interface{})
		update["start_on_reboot"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("start_on_reboot")
	}

	if d.HasChange("start_on_shutdown") {
		_, n := d.GetChange("start_on_shutdown")
		update := make(map[string]interface{})
		update["start_on_shutdown"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("start_on_shutdown")
	}

	if d.HasChange("virtualization") {
		_, n := d.GetChange("virtualization")
		update := make(map[string]interface{})
		update["virtualization"] = n
		_, err := hc.InstanceUpdate(d.Id(), update)
		if err != nil {
			fmt.Errorf("%v", err)
		}
		waitErr := waitInstanceUpdate(hc, d.Id(), 30) //Should be more like 10 secs
		if waitErr != nil {
			fmt.Errorf("%v", waitErr)
		}
		d.SetPartial("virtualization")
	}

	//We did it Reddit!
	d.Partial(false)

	return resourceHypercloudInstanceRead(d, meta)
}

func resourceHypercloudInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	hc := hcc.ToHypercloud(meta)
	_, err := hc.InstanceDelete(d.Id())
	if err != nil {
		fmt.Errorf("%v", err)
	}
	waitInstanceTerminate(hc, d.Id(), 30)
	d.SetId("")
	return nil
}

func resourceHypercloudInstanceExists(d *schema.ResourceData, meta interface{}) (exists bool, err error) {
	hc := hcc.ToHypercloud(meta)
	_, infErr := hc.InstanceInfo(d.Id())
	if infErr == nil {
		exists = true
		return
	}
	exists = false
	return
}

func waitInstanceUp(meta interface{}, id string, timeoutS int) error {
	hc := hcc.ToHypercloud(meta)
	start := time.Now() //Possibly need for logging
	end := start.Add(time.Duration(timeoutS) * time.Second)
	for end.After(time.Now()) {
		state, err := hc.InstanceInfo(id)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		if state.(map[string]interface{})["state"].(string) == "stopped" {
			return nil
		}
		if state.(map[string]interface{})["state"].(string) != "provisioning" || state.(map[string]interface{})["state"].(string) != "initial" || state.(map[string]interface{})["state"].(string) != "pending_verification" {
			return fmt.Errorf(state.(map[string]interface{})["state"].(string))
		}
	}
	return fmt.Errorf("Timeout (stuck provisioning?).")
}

func waitInstanceUpdate(meta interface{}, id string, timeoutS int) error {
	hc := hcc.ToHypercloud(meta)
	start := time.Now()
	end := start.Add(time.Duration(timeoutS) * time.Second)
	for end.After(time.Now()) {
		state, err := hc.InstanceInfo(id)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		if state.(map[string]interface{})["state"].(string) != "updating" { //TODO: Have non-normal states also return error/their state
			return nil
		}
	}

	return fmt.Errorf("Timed out. (resource stuck updating?)")
}

func waitInstanceTerminate(meta interface{}, id string, timeoutS int) error {
	hc := hcc.ToHypercloud(meta)
	start := time.Now()
	end := start.Add(time.Duration(timeoutS) * time.Second)
	for end.After(time.Now()) {
		state, err := hc.InstanceInfo(id)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		state = state.(map[string]interface{})
		state = state.(map[string]interface{})
		if state.(map[string]interface{})["state"].(string) == "terminated" {
			return nil
		}
	}

	return fmt.Errorf("Timed out on terminate.")
}
