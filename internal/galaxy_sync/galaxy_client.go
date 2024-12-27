package galaxy_sync

import (
	"encoding/json"
	"fmt"
	"io"
    "io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jctanner/lax/internal/utils"
	"github.com/sirupsen/logrus"
)

type CachedGalaxyClient struct {
	baseUrl     string
    authUrl     string
    token       string
    accessToken string
    apiPrefix   string
	cachePath   string
}

func NewCachedGalaxyClient(baseUrl string, authUrl string, token string, apiPrefix string, cachePath string) CachedGalaxyClient {

    // create the access token if there is an authUrl ...
	var accessToken string

	// If authUrl is provided, fetch the access token
	if authUrl != "" {
		// Prepare the form data
		formData := url.Values{}
		formData.Set("grant_type", "refresh_token")
		formData.Set("client_id", "cloud-services")
		formData.Set("refresh_token", token)
		fmt.Printf("Form Data: %s\n", formData.Encode())

		// Create the HTTP POST request
		resp, err := http.PostForm(authUrl, formData)
		if err != nil {
			fmt.Printf("Error making request to authUrl: %v\n", err)
			return CachedGalaxyClient{}
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return CachedGalaxyClient{}
		}

		// Parse the response JSON
		var respData map[string]interface{}
		err = json.Unmarshal(body, &respData)
		if err != nil {
			fmt.Printf("Error unmarshalling response JSON: %v\n", err)
			return CachedGalaxyClient{}
		}

		// Extract the access token
		if newtoken, ok := respData["access_token"].(string); ok {
            logrus.Infof("Access token: %s\n", newtoken)
			accessToken = newtoken
		} else {
			fmt.Printf("Access token not found in response: %v\n", respData)
		}
	}

	return CachedGalaxyClient{
		baseUrl: baseUrl,
        authUrl: authUrl,
        token: token,
        accessToken: accessToken,
        apiPrefix: apiPrefix,
        cachePath: cachePath,
	}
}

func (c *CachedGalaxyClient) GetUrl(url string) (resp *http.Response, err error) {
	logrus.Infof("GET URL %s", url)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logrus.Errorf("Error creating HTTP request: %v", err)
		return nil, err
	}

	// Add Authorization header if accessToken is non-empty
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
		logrus.Infof("Added Authorization header with token")
	}

	// Use the default HTTP client to send the request
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		logrus.Errorf("Error making HTTP request: %v", err)
		return nil, err
	}

	return resp, nil
}

// Role represents a single role in the response
type Role struct {
	ID            int               `json:"id"`
	Name          string            `json:"name"`
	GithubUser    string            `json:"github_user"`
	GithubRepo    string            `json:"github_repo"`
	GithubBranch  string            `json:"github_branch"`
	Commit        string            `json:"commit"`
	SummaryFields RoleSummaryFields `json:"summary_fields"`
}

type RoleSummaryFields struct {
	Namespace    RoleNamespace    `json:"namespace"`
	Dependencies RoleDependencies `json:"dependencies"`
	Versions     []RoleVersion    `json:"versions"`
}

type RoleNamespace struct {
	Name string `json:"name"`
}

// RolesResponse represents the API response structure
type RolesResponse struct {
	Results  []Role `json:"results"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Count    int    `json:"count"`
}

// VersionsResponse represents the API response structure for versions
type RoleVersionsResponse struct {
	Results []RoleVersion `json:"results"`
	Next    string        `json:"next"`
}

type RoleDependencies []string
type RoleVersion struct {
	//Id int `json:"id"`
	Name        string `json:"name"`
	ReleaseDate string `json:"release_date"`
}

// GetRoles fetches all the roles from the server's base URL
func (c *CachedGalaxyClient) GetRoles(namespace string, name string, latest_only bool) ([]Role, error) {

	var allRoles []Role
	url := fmt.Sprintf("%s/api/v1/roles/?order_by=-modified", c.baseUrl)
	if namespace != "" && name != "" {
		url = url + fmt.Sprintf("&namespace=%s&name=%s", namespace, name)
	} else if namespace != "" {
		url = url + fmt.Sprintf("&namespace=%s", namespace)
	} else if name != "" {
		url = url + fmt.Sprintf("&name=%s", name)
	}

	roleCount := 0
	rolesFetched := 0

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
				versions, err := c.GetRoleVersions(role.ID, latest_only)
				if err != nil {
					return nil, err
				}
				role.SummaryFields.Versions = versions
				allRoles = append(allRoles, role)
			} else if len(role.SummaryFields.Versions) > 0 && latest_only {
				newVs, _ := reduceRoleVersionsToHighest(role.SummaryFields.Versions)
				role.SummaryFields.Versions = newVs
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
func (c *CachedGalaxyClient) GetRoleVersions(roleID int, latest_only bool) ([]RoleVersion, error) {
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

	if latest_only {
		/*
			//fmt.Printf("%s\n", allVersions)
			vstrings := []string{}
			for _,v := range allVersions {
				vstrings = append(vstrings, v.Name)
			}
			//fmt.Printf("%s\n",vstrings)
			highest,_ := utils.GetHighestSemver(vstrings)

			trimmedVersions := []RoleVersion{}
			for _, v := range allVersions {
				if v.Name == highest {
					trimmedVersions = append(trimmedVersions, v)
				}
			}
			return trimmedVersions, nil
		*/
		trimmedVersions, _ := reduceRoleVersionsToHighest(allVersions)
		return trimmedVersions, nil

		//panic("")
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

func (c *CachedGalaxyClient) loadCollectionsFromCache(path string, response *CollectionResponse) error {
	fmt.Printf("read cached %s\n", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, response)
}

func (c *CachedGalaxyClient) loadCollectionVersionDetailFromCache(path string, response *CollectionVersionDetail) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, response)
}

func (c *CachedGalaxyClient) fetchCollectionVersionDetailsFromServer(url string, path string, response *CollectionVersionDetail) error {
	logrus.Infof("FETCH_COLLECTION_VERSION_DETAILS %s", url)
	//resp, err := http.Get(url)
	resp, err := c.GetUrl(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch collection version details: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, response); err != nil {
		return err
	}

	return os.WriteFile(path, body, 0644)
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
	logrus.Infof("FETCH_ROLES %s", url)
	//resp, err := http.Get(url)
	resp, err := c.GetUrl(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch: %s", resp.Status)
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

// fetchFromServer fetches the response from the server and saves it to the cache file
func (c *CachedGalaxyClient) fetchCollectionsFromServer(url, cacheFile string, response *CollectionResponse) error {
	logrus.Infof("FETCH_COLLECTIONS %s", url)
	//resp, err := http.Get(url)
	resp, err := c.GetUrl(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch collections: %s", resp.Status)
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
	logrus.Infof("FETCH_ROLE_VERSIONS %s", url)
	//resp, err := http.Get(url)
	resp, err := c.GetUrl(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch role versions: %s", resp.Status)
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

func Percentage(a, b int) int {
	if b == 0 {
		return 0
	}
	if a == 0 {
		return 0
	}
	return int((float64(b) / float64(a)) * 100)
}

type Collection struct {
	Href       string `json:"href"`
	Namespace  string `json:"namespace"`
	Name       string `json:"name"`
	Deprecated bool   `json:"deprecated"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type CollectionResponse struct {
	Meta struct {
		Count int `json:"count"`
	} `json:"meta"`
	Links struct {
		First    string `json:"first"`
		Previous string `json:"previous"`
		Next     string `json:"next"`
		Last     string `json:"last"`
	} `json:"links"`
	Data []CrossRepoCollectionIndex `json:"data"`
}

type CrossRepoCollectionIndex struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	CollectionVersion struct {
		PulpHref    string `json:"pulp_href"`
		PulpCreated string `json:"pulp_created"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Tags        []struct {
			Name string `json:"name"`
		} `json:"tags"`
		IsHighest bool `json:"is_highest"`
	} `json:"collection_version"`
}

type CollectionVersionDetail struct {
	Namespace struct {
		Name string `json:"name"`
	} `json:"namespace"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DownloadUrl string `json:"download_url"`
	MetaData    struct {
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	} `json:"metadata"`
	Artifact struct {
		FileName string `json:"filename"`
		Sha256   string `json:"sha256"`
		Size     int    `json:"size"`
	} `json:"artifact"`
}

// GetRoles fetches all the roles from the server's base URL
func (c *CachedGalaxyClient) GetCollections(namespace string, name string, latest_only bool) ([]CollectionVersionDetail, error) {

	// https://galaxy.ansible.com/api/v3/plugin/ansible/search/collection-versions/?is_deprecated=false&repository_label=!hide_from_search&is_highest=true&offset=0&limit=10&order_by=name
	// https://galaxy.ansible.com/api/v3/collectionsgeerlingguy/mac/

	//var allCollections []Collection
	var allCollectionVersionDetails []CollectionVersionDetail

	//url := fmt.Sprintf("%s/api/v3/plugin/ansible/search/collection-versions/", c.baseUrl)
	url := fmt.Sprintf("%s%s/v3/plugin/ansible/search/collection-versions/", c.baseUrl, c.apiPrefix)
	if namespace != "" && name != "" {
		url = url + fmt.Sprintf("?namespace=%s&name=%s", namespace, name)
	} else if namespace != "" {
		url = url + fmt.Sprintf("?namespace=%s", namespace)
	} else if name != "" {
		url = url + fmt.Sprintf("?name=%s", name)
	}

	if latest_only {
		if strings.Contains(url, "?") {
			url = url + "&is_highest=true"
		} else {
			url = url + "?is_highest=true"
		}
	}

	collectionCount := 0
	collectionsFetched := 0

	for url != "" {
		pct := Percentage(collectionCount, collectionsFetched)
		logrus.Infof("%d|%d %d%% %s\n", collectionCount, collectionsFetched, pct, url)

		cacheFile := c.getCacheFilePath(url)

		var collectionsResponse CollectionResponse
		if c.isCacheFileExist(cacheFile) {
			err := c.loadCollectionsFromCache(cacheFile, &collectionsResponse)
			if err != nil {
				return nil, err
			}
		} else {
			err := c.fetchCollectionsFromServer(url, cacheFile, &collectionsResponse)
			if err != nil {
				return nil, err
			}
		}

		for _, col := range collectionsResponse.Data {
			logrus.Debugf("process collection result: %v\n", col)
			// need to get the details page to find the download url ...
			// /api/v3/plugin/ansible/content/published/collections/index/geerlingguy/mac/versions/4.0.1/
            //logrus.Infof("%v\n", col)
			detailsUrl := fmt.Sprintf(
				"%s%s/v3/plugin/ansible/content/%s/collections/index/%s/%s/versions/%s/",
				c.baseUrl,
				c.apiPrefix,
                col.Repository.Name,
				col.CollectionVersion.Namespace,
				col.CollectionVersion.Name,
				col.CollectionVersion.Version,
			)
			logrus.Debugf("details url: %s\n", detailsUrl)

			detailsCacheFile := c.getCacheFilePath(detailsUrl)
			var details CollectionVersionDetail

			if c.isCacheFileExist(detailsCacheFile) {
				err := c.loadCollectionVersionDetailFromCache(detailsCacheFile, &details)
				if err != nil {
					//panic("")
					return nil, err
				}
			} else {
				err := c.fetchCollectionVersionDetailsFromServer(detailsUrl, detailsCacheFile, &details)
				if err != nil {
					//panic("")
					return nil, err
				}
			}

			//detail,_ := c.fetchCollectionVersionDetail(, detailsUrl, )
			logrus.Debugf("download-url: %s\n", details.DownloadUrl)

			allCollectionVersionDetails = append(allCollectionVersionDetails, details)
		}

		/*
			collectionCount = collectionsResponse.Count
			allCollections = append(allCollections, collectionsResponse.Data...)
			collectionsFetched = len(allCollections)
		*/
		if collectionsResponse.Links.Next != "" {
			url = c.baseUrl + collectionsResponse.Links.Next
		} else {
			url = ""
		}

	}

	return allCollectionVersionDetails, nil
}

func reduceRoleVersionsToHighest(allVersions []RoleVersion) ([]RoleVersion, error) {
	vstrings := []string{}
	for _, v := range allVersions {
		vstrings = append(vstrings, v.Name)
	}
	//fmt.Printf("%s\n",vstrings)
	highest, _ := utils.GetHighestSemver(vstrings)

	trimmedVersions := []RoleVersion{}
	for _, v := range allVersions {
		if v.Name == highest {
			trimmedVersions = append(trimmedVersions, v)
		}
	}
	return trimmedVersions, nil
}
