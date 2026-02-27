package auth

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ClaimRoles           = "urn:zitadel:iam:org:project:roles"
	ClaimResourceOwnerID = "urn:zitadel:iam:user:resourceowner:id"
	ClaimResourceOwner   = "urn:zitadel:iam:user:resourceowner:name"
)

// ZitadelClaims represents the JWT claims structure from Zitadel
type ZitadelClaims struct {
	jwt.RegisteredClaims
	Email             string                       `json:"email"`
	Roles             map[string]map[string]string `json:"urn:zitadel:iam:org:project:roles"`
	ResourceOwnerID   string                       `json:"urn:zitadel:iam:user:resourceowner:id"`
	ResourceOwnerName string                       `json:"urn:zitadel:iam:user:resourceowner:name"`
}

// UserRoles contains parsed role information for a user
type UserRoles struct {
	AllRoles      []string
	PlatformAdmin bool
	OrgRoles      map[string][]string
}

// AuthContext contains the authentication context for a request
type AuthContext struct {
	UserID            string
	Email             string
	ResourceOwnerID   string
	ResourceOwnerName string
	Roles             UserRoles
}

// ParseRoles extracts and categorizes roles from Zitadel role claims
func ParseRoles(roleClaims map[string]map[string]string) UserRoles {
	result := UserRoles{
		AllRoles: []string{},
		OrgRoles: make(map[string][]string),
	}

	for roleKey, orgMap := range roleClaims {
		for orgID, orgName := range orgMap {
			result.AllRoles = append(result.AllRoles, roleKey)

			if roleKey == "platform:admin" {
				result.PlatformAdmin = true
			}

			if result.OrgRoles[orgID] == nil {
				result.OrgRoles[orgID] = []string{}
			}
			result.OrgRoles[orgID] = append(result.OrgRoles[orgID], roleKey)

			if orgName != "" {
				if result.OrgRoles[orgName] == nil {
					result.OrgRoles[orgName] = []string{}
				}
				result.OrgRoles[orgName] = append(result.OrgRoles[orgName], roleKey)
			}
		}
	}

	return result
}

// HasRole checks if the user has a specific role
func (r *UserRoles) HasRole(role string) bool {
	for _, userRole := range r.AllRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasRoleWithPrefix checks if the user has any role starting with the given prefix
func (r *UserRoles) HasRoleWithPrefix(prefix string) bool {
	for _, userRole := range r.AllRoles {
		if strings.HasPrefix(userRole, prefix) {
			return true
		}
	}
	return false
}

// IsPlatformAdmin checks if the user is a platform admin
func (ac *AuthContext) IsPlatformAdmin() bool {
	return ac.Roles.PlatformAdmin
}
