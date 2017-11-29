package hypercloud

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	hcc "bitbucket.org/mistarhee/hypercloud-go-client/hypercloud" //Replace with "official" repo
)

//Can't really test anything apart from just creating a basic instance with some ram, a name and a specific region/performance tier
func TestResourceHypercloudInstance_basic(t *testing.T) {
	t.Parallel()

	var name = fmt.Sprintf("terraform-instance-test-%s", acctest.RandString(10))
	var instance map[string]interface{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstance_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists("hypercloud_instance.PotatoStomper", &instance),
					testAccCheckInstanceName(&instance, name),
					testAccCheckInstanceRam(&instance, 4096),
					testAccCheckInstancePerformanceTier(&instance, "55f841d6-7e19-4de9-be47-93f650ff9f9b"),
					testAccCheckInstanceRegion(&instance, "9e9806d3-d542-4ef0-878a-588c49ffcf50"),
				),
			},
		},
	})
}

func testAccInstance_basic(instance string) string {
	return fmt.Sprintf(`
resource "hypercloud_instance" "PotatoStomper" {
    memory = 4096
    name = "%s"
    performance_tier = "55f841d6-7e19-4de9-be47-93f650ff9f9b"
    region = "9e9806d3-d542-4ef0-878a-588c49ffcf50"
}
`, instance)
}

func testAccCheckInstanceExists(n string, instance *map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		hc := hcc.ToHypercloud(testAccProvider.Meta())
		instanceInfo, err := hc.InstanceInfo(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Failed to get instance info: \n%v", err)
		}

		if instanceInfo.(map[string]interface{})["id"] != rs.Primary.ID {
			return fmt.Errorf("Instance not found")
		}
		*instance = instanceInfo.(map[string]interface{})

		return nil
	}
}

func testAccCheckInstanceName(instance *map[string]interface{}, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if (*instance)["name"].(string) != name {
			return fmt.Errorf("Instance name %s doesn't match generated name %s", (*instance)["name"].(string), name)
		}
		return nil
	}
}

func testAccCheckInstanceRam(instance *map[string]interface{}, ram int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if int((*instance)["memory"].(float64)) != ram {
			return fmt.Errorf("Instance memory %d doesn't match provided memory %d", int((*instance)["memory"].(float64)), ram)
		}
		return nil
	}
}

func testAccCheckInstancePerformanceTier(instance *map[string]interface{}, performance_tier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if (*instance)["performance_tier"].(map[string]interface{})["id"].(string) != performance_tier {
			return fmt.Errorf("Instance performance_tier %s doesn't match generated performance_tier %s", (*instance)["performance_tier"].(string), performance_tier)
		}
		return nil
	}
}

func testAccCheckInstanceRegion(instance *map[string]interface{}, region string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if (*instance)["region"].(map[string]interface{})["id"].(string) != region {
			return fmt.Errorf("Instance region %s doesn't match generated region %s", (*instance)["region"].(string), region)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	hc := hcc.ToHypercloud(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "hypercloud_instance" {
			continue
		}

		_, err := hc.InstanceInfo(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Instance %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
