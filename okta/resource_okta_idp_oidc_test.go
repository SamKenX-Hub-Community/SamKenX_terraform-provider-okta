package okta

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccOktaIdpOidc_crud(t *testing.T) {
	ri := acctest.RandInt()
	mgr := newFixtureManager(idpOidc)
	config := mgr.GetFixtures("generic_oidc.tf", ri, t)
	updatedConfig := mgr.GetFixtures("generic_oidc_updated.tf", ri, t)
	resourceName := fmt.Sprintf("%s.test", idpOidc)

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ErrorCheck:        testAccErrorChecks(t),
		ProviderFactories: testAccProvidersFactories,
		CheckDestroy:      createCheckResourceDestroy(idpOidc, createDoesIdpExist()),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", buildResourceName(ri)),
					resource.TestCheckResourceAttr(resourceName, "authorization_url", "https://idp.example.com/authorize"),
					resource.TestCheckResourceAttr(resourceName, "authorization_binding", "HTTP-REDIRECT"),
					resource.TestCheckResourceAttr(resourceName, "token_url", "https://idp.example.com/token"),
					resource.TestCheckResourceAttr(resourceName, "token_binding", "HTTP-POST"),
					resource.TestCheckResourceAttr(resourceName, "user_info_url", "https://idp.example.com/userinfo"),
					resource.TestCheckResourceAttr(resourceName, "user_info_binding", "HTTP-REDIRECT"),
					resource.TestCheckResourceAttr(resourceName, "jwks_url", "https://idp.example.com/keys"),
					resource.TestCheckResourceAttr(resourceName, "jwks_binding", "HTTP-REDIRECT"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "efg456"),
					resource.TestCheckResourceAttr(resourceName, "client_secret", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
					resource.TestCheckResourceAttr(resourceName, "issuer_url", "https://id.example.com"),
					resource.TestCheckResourceAttr(resourceName, "username_template", "idpuser.email"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", buildResourceName(ri)),
					resource.TestCheckResourceAttr(resourceName, "authorization_url", "https://idp.example.com/authorize2"),
					resource.TestCheckResourceAttr(resourceName, "authorization_binding", "HTTP-REDIRECT"),
					resource.TestCheckResourceAttr(resourceName, "token_url", "https://idp.example.com/token2"),
					resource.TestCheckResourceAttr(resourceName, "token_binding", "HTTP-POST"),
					resource.TestCheckResourceAttr(resourceName, "user_info_url", "https://idp.example.com/userinfo2"),
					resource.TestCheckResourceAttr(resourceName, "user_info_binding", "HTTP-REDIRECT"),
					resource.TestCheckResourceAttr(resourceName, "jwks_url", "https://idp.example.com/keys2"),
					resource.TestCheckResourceAttr(resourceName, "jwks_binding", "HTTP-REDIRECT"),
					resource.TestCheckResourceAttr(resourceName, "client_id", "efg456"),
					resource.TestCheckResourceAttr(resourceName, "client_secret", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
					resource.TestCheckResourceAttr(resourceName, "issuer_url", "https://id.example.com"),
					resource.TestCheckResourceAttr(resourceName, "username_template", "idpuser.email"),
				),
			},
		},
	})
}
