package galaxy_sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type CachedGalaxyClient struct {
	baseUrl string
	cachePath string
}


// Role represents a single role in the response
type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	GithubUser string `json:"github_user"`
	GithubRepo string `json:"github_repo"`
	GithubBranch string `json:"github_branch"`
	Commit string `json:"commit"`
	SummaryFields RoleSummaryFields `json:"summary_fields"`
}

type RoleSummaryFields struct {
	Namespace RoleNamespace `json:"namespace"`
	Dependencies RoleDependencies  `json:"dependencies"`
	Versions []RoleVersion `json:"versions"`
}

type RoleNamespace struct {
	Name      string `json:"name"`
}

// RolesResponse represents the API response structure
type RolesResponse struct {
	Results []Role `json:"results"`
	Next string `json:"next"`
	Previous string `json:"previous"`
	Count int `json:"count"`
}

// VersionsResponse represents the API response structure for versions
type RoleVersionsResponse struct {
	Results []RoleVersion `json:"results"`
	Next    string `json:"next"`
}

type RoleDependencies []string
type RoleVersion struct {
	//Id int `json:"id"`
	Name string `json:"name"`
	ReleaseDate string `json:"release_date"`
}

// GetRoles fetches all the roles from the server's base URL
func (c *CachedGalaxyClient) GetRoles(namespace string, name string) ([]Role, error) {

	var allRoles []Role
	url := fmt.Sprintf("%s/api/v1/roles/?order_by=-modified", c.baseUrl)
	if namespace != "" && name != "" {
		url = url + fmt.Sprintf("&namespace=%s&name=%s", namespace, name)
	} else if namespace != "" {
		url = url + fmt.Sprintf("&namespace=%s", namespace)
	} else if name != "" {
		url = url + fmt.Sprintf("&name=%s", name)
	}

	roleCount := 0;
	rolesFetched := 0;

	for url != "" {
		pct := Percentage(roleCount, rolesFetched)
		fmt.Printf("%d|%d %d%% %s\n", roleCount, rolesFetched, pct, url)

		cacheFile := c.getCacheFilePath(url)

		var rolesResponse RolesResponse
		if c.isCacheFileExist(cacheFile) {
			err := c.loadFromCache(cacheFile, &rolesResponse)
			if err != nil {
				return nil, err
			}
		} else {
			err := c.fetchFromServer(url, cacheFile, &rolesResponse)
			if err != nil {
				return nil, err
			}
		}

		for _, role := range rolesResponse.Results {
			// Fetch all versions for each role if the summary has more than 10 ...
			if len(role.SummaryFields.Versions) >= 10 {
				versions, err := c.GetRoleVersions(role.ID)
				if err != nil {
					return nil, err
				}
				role.SummaryFields.Versions = versions
				allRoles = append(allRoles, role)
			}
		}

		roleCount = rolesResponse.Count
		allRoles = append(allRoles, rolesResponse.Results...)
		rolesFetched = len(allRoles)
		url = rolesResponse.Next

	}

	return allRoles, nil
}


// GetVersions fetches all versions for a given role ID, handling pagination and caching
func (c *CachedGalaxyClient) GetRoleVersions(roleID int) ([]RoleVersion, error) {
	var allVersions []RoleVersion
	url := fmt.Sprintf("%s/api/v1/roles/%d/versions/", c.baseUrl, roleID)

	for url != "" {
		fmt.Printf("\t%s\n", url)
		cacheFile := c.getCacheFilePath(url)

		var versionsResponse RoleVersionsResponse

		if c.isCacheFileExist(cacheFile) {
			err := c.loadVersionsFromCache(cacheFile, &versionsResponse)
			if err != nil {
				return nil, err
			}
		} else {
			err := c.fetchVersionsFromServer(url, cacheFile, &versionsResponse)
			if err != nil {
				return nil, err
			}
		}

		allVersions = append(allVersions, versionsResponse.Results...)
		url = versionsResponse.Next
	}

	return allVersions, nil
}

// UnmarshalJSON custom unmarshaler to handle different types for dependencies field
func (d *RoleDependencies) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a slice of objects with "role" field
	var objectDependencies []struct {
		Role string `json:"role"`
	}
	if err := json.Unmarshal(data, &objectDependencies); err == nil {
		for _, dep := range objectDependencies {
			*d = append(*d, dep.Role)
		}
		return nil
	}

	// Try to unmarshal as a slice of strings
	var stringDependencies []string
	if err := json.Unmarshal(data, &stringDependencies); err == nil {
		*d = RoleDependencies(stringDependencies)
		return nil
	}

	return fmt.Errorf("dependencies field is of an unsupported type")
}

// getCacheFilePath generates the cache file path for a given URL
func (c *CachedGalaxyClient) getCacheFilePath(url string) string {
	return filepath.Join(c.cachePath, "roles", fmt.Sprintf("%x.json", hash(url)))
}

// isCacheFileExist checks if the cache file exists
func (c *CachedGalaxyClient) isCacheFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// loadFromCache loads the response from the cache file
func (c *CachedGalaxyClient) loadFromCache(path string, response *RolesResponse) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, response)
}

func (c *CachedGalaxyClient) loadVersionsFromCache(path string, response *RoleVersionsResponse) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, response)
}

// fetchFromServer fetches the response from the server and saves it to the cache file
func (c *CachedGalaxyClient) fetchFromServer(url, cacheFile string, response *RolesResponse) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch roles: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	return os.WriteFile(cacheFile, body, 0644)
}


func (c *CachedGalaxyClient) fetchVersionsFromServer(url, cacheFile string, response *RoleVersionsResponse) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch roles: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	return os.WriteFile(cacheFile, body, 0644)
}

// hash generates a hash value for a given string (URL)
func hash(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}


func Percentage(a, b int) (int) {
	if b == 0 {
		return 0
	}
	if a == 0 {
		return 0
	}
	return int((float64(b) / float64(a)) * 100)
}

