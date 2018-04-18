package openldap

import (
	"fmt"

	"github.com/bzon/adop-ctl/pkg/gitlab"
)

// Group is an LDAP Group struct
type Group struct {
	cn            string
	uniqueMembers []string
}

// GetGroup gets specific group(s) and return an array of openldap.Group
func (openldap *Client) GetGroup(baseDN string, groupName ...string) ([]Group, error) {

	// Initialize Slice depending of number of groupnames
	var groups = make([]Group, len(groupName))

	// Loop through Groupnames
	for j := range groups {

		// Get Matched CNs
		cn, err := openldap.NewSearch("ou=groups,"+baseDN, "(&(objectClass=groupOfUniqueNames)(cn="+groupName[j]+"))", "cn")
		if err != nil {
			return nil, err
		}
		// Loop through Matched CNs
		for i, value := range cn {

			// Assign Group with exact match only
			if value == groupName[j] {
				uniqueMembers, err := openldap.NewSearch("cn="+cn[i]+",ou=groups,"+baseDN, "(&(objectClass=groupOfUniqueNames))", "uniqueMember")
				if err != nil {
					return nil, err
				}

				// Initialize only the group with the exact value
				groups[j] = Group{
					cn:            cn[i],
					uniqueMembers: uniqueMembers,
				}
			}

		}
	}

	return groups, nil
}

// GetGroupList gets List of groups under ou=groups
func (openldap *Client) GetGroupList(baseDN string) ([]string, error) {

	// Get Group List
	groups, err := openldap.NewSearch("ou=groups,"+baseDN, "(&(objectClass=groupOfUniqueNames))", "cn")
	if err != nil {
		return nil, fmt.Errorf("Failed to get groups. %s", err)
	}
	return groups, nil

}

// SyncGitlabGroup accepts list of groups
func (openldap *Client) SyncGitlabGroup(ldapGroup []Group, GitlabAPI gitlab.API) error {

	// loop through groups
	for j := 0; j < len(ldapGroup); j++ {
		gitlabGroup := gitlab.Group{
			Name: ldapGroup[0].cn,
			Path: "ldap_" + ldapGroup[0].cn,
		}

		// Delete group
		fmt.Printf("Deleting group path: \"%s\"\n", gitlabGroup.Path)
		GitlabAPI.DeleteGroupByPath(gitlabGroup.Path)

		// Create Group
		fmt.Printf("Creating group \"%s\" at path: \"%s\"\n", gitlabGroup.Name, gitlabGroup.Path)
		GitlabAPI.CreateGroup(gitlabGroup)

		// Loop through group
		for i := 0; i < len(ldapGroup[j].uniqueMembers); i++ {

			member := gitlab.User{
				Name:        ldapGroup[j].uniqueMembers[i],
				Username:    ldapGroup[j].uniqueMembers[i],
				AccessLevel: gitlab.OwnerLevel,
			}

			// add member to group
			GitlabAPI.AddMemberToGroup(member, gitlabGroup.Path)

		}
	}
	return nil
}
