package okta

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
)

const (
	statusActive   = "ACTIVE"
	statusInactive = "INACTIVE"

	userStatusPasswordExpired = "PASSWORD_EXPIRED"
	userStatusProvisioned     = "PROVISIONED"
	userStatusDeprovisioned   = "DEPROVISIONED"
	userStatusStaged          = "STAGED"
	userStatusSuspended       = "SUSPENDED"
	userStatusRecovery        = "RECOVERY"
	userStatusLockedOut       = "LOCKED_OUT"

	userScope = "USER"

	groupProfileEveryone = "Everyone"
)

var userProfileDataSchema = map[string]*schema.Schema{
	"admin_roles": {
		Type:     schema.TypeSet,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
	"roles": {
		Type:     schema.TypeSet,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
	"city": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"cost_center": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"country_code": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"custom_profile_attributes": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"department": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"display_name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"division": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"email": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"employee_number": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"first_name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"group_memberships": {
		Type:     schema.TypeSet,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	},
	"honorific_prefix": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"honorific_suffix": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"last_name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"locale": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"login": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"manager": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"manager_id": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"middle_name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"mobile_phone": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"nick_name": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"organization": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"postal_address": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"preferred_language": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"primary_phone": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"profile_url": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"second_email": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"state": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"status": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"street_address": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"timezone": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"title": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"user_type": {
		Type:     schema.TypeString,
		Computed: true,
	},
	"zip_code": {
		Type:     schema.TypeString,
		Computed: true,
	},
}

var validAdminRoles = []string{
	"API_ACCESS_MANAGEMENT_ADMIN",
	"APP_ADMIN",
	"CUSTOM",
	"GROUP_MEMBERSHIP_ADMIN",
	"HELP_DESK_ADMIN",
	"MOBILE_ADMIN",
	"ORG_ADMIN",
	"READ_ONLY_ADMIN",
	"REPORT_ADMIN",
	"SUPER_ADMIN",
	"USER_ADMIN",
}

func buildUserDataSourceSchema(target map[string]*schema.Schema) map[string]*schema.Schema {
	return buildSchema(userProfileDataSchema, target)
}

func assignAdminRolesToUser(ctx context.Context, userID string, roles []string, disableNotifications bool, client *okta.Client) error {
	for _, role := range roles {
		if role == "CUSTOM" {
			continue
		}
		_, _, err := client.User.AssignRoleToUser(ctx, userID, okta.AssignRoleRequest{Type: role},
			&query.Params{DisableNotifications: boolPtr(disableNotifications)})
		if err != nil {
			return fmt.Errorf("failed to assign role '%s' to user '%s': %w", role, userID, err)
		}
	}
	return nil
}

func assignGroupsToUser(ctx context.Context, userID string, groups []string, c *okta.Client) error {
	for _, group := range groups {
		_, err := c.Group.AddUserToGroup(ctx, group, userID)
		if err != nil {
			return fmt.Errorf("failed to assign group '%s' to user '%s': %w", group, userID, err)
		}
	}
	return nil
}

func populateUserProfile(d *schema.ResourceData) *okta.UserProfile {
	profile := okta.UserProfile{}

	if rawAttrs, ok := d.GetOk("custom_profile_attributes"); ok {
		var attrs map[string]interface{}
		str := rawAttrs.(string)

		// We validate the JSON, no need to check error
		_ = json.Unmarshal([]byte(str), &attrs)
		for k, v := range attrs {
			profile[k] = v
		}
	}

	profile["firstName"] = d.Get("first_name").(string)
	profile["lastName"] = d.Get("last_name").(string)
	profile["login"] = d.Get("login").(string)
	profile["email"] = d.Get("email").(string)

	getSetParams := []string{
		"city", "costCenter", "countryCode", "department", "displayName", "division",
		"employeeNumber", "honorificPrefix", "honorificSuffix", "locale", "manager", "managerId", "middleName",
		"mobilePhone", "nickName", "organization", "preferredLanguage", "primaryPhone", "profileUrl",
		"secondEmail", "state", "streetAddress", "timezone", "title", "userType", "zipCode",
	}

	for i := range getSetParams {
		if res, ok := d.GetOk(camelCaseToUnderscore(getSetParams[i])); ok {
			profile[getSetParams[i]] = res.(string)
		}
	}

	// need to set profile.postalAddress to nil explicitly if not set because of a bug with this field
	// have a support ticket open with okta about it
	if _, ok := d.GetOk("postal_address"); ok {
		profile["postalAddress"] = d.Get("postal_address").(string)
	} else {
		profile["postalAddress"] = nil
	}

	return &profile
}

func listUserRoles(ctx context.Context, c *okta.Client, userID string) (userOnlyRoles []*okta.Role, resp *okta.Response, err error) {
	roles, resp, err := c.User.ListAssignedRolesForUser(ctx, userID, nil)
	if err != nil {
		return
	}
	userOnlyRoles = append(userOnlyRoles, roles...)
	return
}

func getRoles(ctx context.Context, id string, c *okta.Client) ([]interface{}, error) {
	roleTypes := make([]interface{}, 0)
	roles, resp, err := listUserRoles(ctx, c, id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			// no-op
		} else {
			return nil, err
		}
	} else {
		for _, role := range roles {
			roleTypes = append(roleTypes, role.Type)
		}
	}
	return roleTypes, err
}

func setRoles(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	roleTypes, err := getRoles(ctx, d.Id(), getOktaClientFromMetadata(m))
	if err != nil {
		return fmt.Errorf("failed to get roles: %v", err)
	}
	// set the custom_profile_attributes values
	return setNonPrimitives(d, map[string]interface{}{
		"roles": schema.NewSet(schema.HashString, roleTypes),
	})
}

func listUserOnlyRoles(ctx context.Context, c *okta.Client, userID string) (userOnlyRoles []*okta.Role, resp *okta.Response, err error) {
	roles, resp, err := c.User.ListAssignedRolesForUser(ctx, userID, nil)
	if err != nil {
		return
	}
	for _, role := range roles {
		if role.AssignmentType == userScope && role.Type != "CUSTOM" {
			userOnlyRoles = append(userOnlyRoles, role)
		}
	}
	return
}

func getAdminRoles(ctx context.Context, id string, c *okta.Client) ([]interface{}, *okta.Response, error) {
	roleTypes := make([]interface{}, 0)
	roles, resp, err := listUserOnlyRoles(ctx, c, id)

	if err != nil {
		return roleTypes, resp, err
	} else {
		for _, role := range roles {
			roleTypes = append(roleTypes, role.Type)
		}
	}

	return roleTypes, resp, err
}

func setAdminRoles(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	roleTypes, resp, err := getAdminRoles(ctx, d.Id(), getOktaClientFromMetadata(m))
	if err := suppressErrorOn403("setting admin roles", m, resp, err); err != nil {
		return fmt.Errorf("failed to get admin roles: %v", err)
	}

	// set the custom_profile_attributes values
	return setNonPrimitives(d, map[string]interface{}{
		"admin_roles": schema.NewSet(schema.HashString, roleTypes),
	})
}

func getGroupsForUser(ctx context.Context, id string, c *okta.Client) ([]string, error) {
	groups, response, err := c.User.ListUserGroups(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to list user groups: %v", err)
	}

	groupIDs := make([]string, 0)

	for {
		for _, group := range groups {
			groupIDs = append(groupIDs, group.Id)
		}

		if !response.HasNextPage() {
			break
		}

		response, err = response.Next(ctx, &groups)

		if err != nil {
			return nil, fmt.Errorf("failed to list user groups: %v", err)
		}
	}

	return groupIDs, nil
}

// set all groups currently attached to the user
func setAllGroups(ctx context.Context, d *schema.ResourceData, c *okta.Client) error {
	groupIDs, err := getGroupsForUser(ctx, d.Id(), c)
	if err != nil {
		return err
	}
	gids := convertStringSliceToInterfaceSlice(groupIDs)
	return setNonPrimitives(d, map[string]interface{}{
		"group_memberships": schema.NewSet(schema.HashString, gids),
	})
}

// set groups attached to the user that can be changed
func setGroupUserMemberships(ctx context.Context, d *schema.ResourceData, c *okta.Client) error {
	groups, response, err := c.User.ListUserGroups(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("failed to list user groups: %v", err)
	}

	groupIDs := make([]interface{}, 0)

	for {
		// ignore saving build-in or app groups into state so we don't end up with perpetual diffs,
		// because it's impossible to remove user from build-in or app group via API
		for _, group := range groups {
			if group.Type != "BUILT_IN" && group.Type != "APP_GROUP" {
				groupIDs = append(groupIDs, group.Id)
			}
		}

		if !response.HasNextPage() {
			break
		}

		response, err = response.Next(ctx, &groups)

		if err != nil {
			return fmt.Errorf("failed to list user groups: %v", err)
		}
	}

	return setNonPrimitives(d, map[string]interface{}{
		"group_memberships": schema.NewSet(schema.HashString, groupIDs),
	})
}

func isCustomUserAttr(key string) bool {
	return !contains(profileKeys, key)
}

func flattenUser(u *okta.User, filteredCustomAttributes []string) map[string]interface{} {
	customAttributes := make(map[string]interface{})
	attrs := map[string]interface{}{}

	for k, v := range *u.Profile {
		if v != nil {
			attrKey := camelCaseToUnderscore(k)

			if isCustomUserAttr(attrKey) {

				// Exclude any custom attributes that should be filtered
				if contains(filteredCustomAttributes, attrKey) {
					continue
				}

				// Supporting any potential type
				ref := reflect.ValueOf(v)
				switch ref.Kind() {
				case reflect.String:
					customAttributes[k] = ref.String()
				case reflect.Float64:
					customAttributes[k] = ref.Float()
				case reflect.Int:
					customAttributes[k] = ref.Int()
				case reflect.Bool:
					customAttributes[k] = ref.Bool()
				case reflect.Slice:
					rawArr := v.([]interface{})
					customAttributes[k] = rawArr
				case reflect.Map:
					rawMap := v.(map[string]interface{})
					customAttributes[k] = rawMap
				}
			} else {
				attrs[attrKey] = v
			}
		}
	}

	attrs["status"] = mapStatus(u.Status)

	data, _ := json.Marshal(customAttributes)
	attrs["custom_profile_attributes"] = string(data)

	return attrs
}

// handle setting of user status based on what the current status is because okta
// only allows transitions to certain statuses from other statuses - consult okta User API docs for more info
// https://developer.okta.com/docs/api/resources/users#lifecycle-operations
func updateUserStatus(ctx context.Context, uid, desiredStatus string, c *okta.Client) error {
	user, _, err := c.User.GetUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	var statusErr error
	switch desiredStatus {
	case userStatusSuspended:
		_, statusErr = c.User.SuspendUser(ctx, uid)
	case userStatusDeprovisioned:
		_, statusErr = c.User.DeactivateUser(ctx, uid, nil)
	case statusActive:
		switch user.Status {
		case userStatusSuspended:
			_, statusErr = c.User.UnsuspendUser(ctx, uid)
		case userStatusPasswordExpired:
			// Ignore password expired status. This status is already activated.
			return nil
		case userStatusLockedOut:
			_, statusErr = c.User.UnlockUser(ctx, uid)
		default:
			_, _, statusErr = c.User.ActivateUser(ctx, uid, nil)
		}
	}
	if statusErr != nil {
		return statusErr
	}
	return waitForStatusTransition(ctx, uid, c)
}

// need to wait for user.TransitioningToStatus field to be empty before allowing Terraform to continue
// so the proper current status gets set in the state during the Read operation after a Status update
func waitForStatusTransition(ctx context.Context, u string, c *okta.Client) error {
	user, _, err := c.User.GetUser(ctx, u)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	for {
		if user.TransitioningToStatus == "" {
			return nil
		}

		log.Printf("[INFO] Transitioning to status = %v; waiting for 5 more seconds...", user.TransitioningToStatus)
		time.Sleep(5 * time.Second)
		user, _, err = c.User.GetUser(ctx, u)
		if err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}
	}
}
