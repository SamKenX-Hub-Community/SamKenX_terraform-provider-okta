package okta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
)

func dataSourceApp() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"label", "label_prefix"},
			},
			"label": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "label_prefix"},
			},
			"label_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id", "label"},
			},
			"active_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Search only ACTIVE applications.",
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAppRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	filters, err := getAppFilters(d)
	if err != nil {
		return diag.Errorf("invalid app filters: %v", err)
	}
	var app *okta.Application
	if filters.ID != "" {
		respApp, resp, err := getOktaClientFromMetadata(m).Application.GetApplication(ctx, filters.ID, okta.NewApplication(), nil)
		if err := suppressErrorOn404(resp, err); err != nil {
			return diag.Errorf("failed get app by ID: %v", err)
		}
		if respApp == nil || respApp.(*okta.Application).Id == "" {
			return diag.Errorf("no application found with provided ID: %s", filters.ID)
		}
		app = respApp.(*okta.Application)
	} else {
		appList, err := listApps(ctx, m, filters, 1)
		if err != nil {
			return diag.Errorf("failed to list apps: %v", err)
		}
		if len(appList) < 1 {
			return diag.Errorf("no application found with the provided filter: %+v", filters)
		}
		if filters.Label != "" && appList[0].Label != filters.Label {
			return diag.Errorf("no application found with provided label: %s", filters.Label)
		} else {
			logger(m).Info("found multiple applications with the criteria supplied, using the first one, sorted by creation date")
			app = appList[0]
		}
	}
	d.SetId(app.Id)
	_ = d.Set("label", app.Label)
	_ = d.Set("name", app.Name)
	_ = d.Set("status", app.Status)
	return nil
}
