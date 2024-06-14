package git_collections

import (
	"errors"
	"fmt"
	"strings"

	"net/http"
	"os"
	"time"

	"cli/utils"
)

/*
func SplitSpec(r rune) bool {
	return r == ':' || r == '.'
}
*/

type InstallSpec struct {
	Server       string
	Namespace    string
	Name         string
	Version      string
	Dependencies []InstallSpec
}

func SplitSpec(input string) []string {

	// geerlingguy.mac
	// github.com:geerlingguy.mac
	// https://github.com:geerlingguy.mac
	// https://github.com/geerlingguy.mac
	// git@github.com:geerlingguy/mac

	colonIndex := strings.Index(input, ":")

	var result []string

	if colonIndex != -1 {
		// Check if the colon is part of a URL scheme
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			// Find the second colon if it exists
			secondColonIndex := strings.Index(input[colonIndex+1:], ":")
			if secondColonIndex != -1 {
				secondColonIndex += colonIndex + 1
				// Split at the second colon
				beforeColon := input[:secondColonIndex]
				afterColon := input[secondColonIndex+1:]

				// Split the part after the second colon on periods
				afterColonParts := strings.Split(afterColon, ".")

				// Combine the parts before the second colon and after the split
				result = append([]string{beforeColon}, afterColonParts...)
			} else {
				// If no second colon exists, treat the entire string as before the colon
				result = []string{input}
			}
		} else {
			// Split at the first colon
			beforeColon := input[:colonIndex]
			afterColon := input[colonIndex+1:]

			// Split the part after the first colon on periods
			afterColonParts := strings.Split(afterColon, ".")

			// Combine the parts before the colon and after the split
			result = append([]string{beforeColon}, afterColonParts...)
		}
	} else {
		// If no colon exists, split on all periods
		result = strings.Split(input, ".")
	}

	return result
}

func ensureDir(path string) error {
	// Check if the path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Path does not exist, create the directory
		err := os.Mkdir(path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check directory: %v", err)
	}

	// Path exists, check if it is a directory
	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory")
	}

	// Path exists and is a directory
	return nil
}

// exists performs an HTTP HEAD request to check if the URL exists
func UrlExists(url string) bool {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error performing request:", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func GetFirstValidURL(urls []string) (string, error) {
	for _, url := range urls {
		if UrlExists(url) {
			return url, nil
		}
	}
	return "", errors.New("no valid URL found")
}

func DoSpecInstall(dest string, server string, spec InstallSpec, recurse bool) error {

	ensureDir(dest)
	collectionsDir := dest + "/ansible_collections"
	ensureDir(collectionsDir)
	installPath := collectionsDir + "/" + spec.Namespace + "/" + spec.Name

	if !utils.FileExists(installPath) {

		// make a list of possible clone urls ...
		var cloneUrls []string
		cloneUrls = append(cloneUrls, fmt.Sprintf("%s/%s/%s", spec.Server, spec.Namespace, spec.Name))
		cloneUrls = append(cloneUrls, fmt.Sprintf("%s/%s/ansible-collection-%s", spec.Server, spec.Namespace, spec.Name))
		cloneUrls = append(cloneUrls, fmt.Sprintf("%s/ansible-collections/%s.%s", spec.Server, spec.Namespace, spec.Name))

		fmt.Printf("%s\n", cloneUrls)

		cloneUrl, err := GetFirstValidURL(cloneUrls)
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("using %s\n", cloneUrl)

		ensureDir(installPath)

		// clone it
		utils.CloneRepo(cloneUrl, installPath)

		// check for tags
		tags, err := utils.ListTags(installPath)
		if err == nil {
			fmt.Printf("tags:%s\n", tags)

			highestVersion, err := utils.GetHighestSemver(tags)
			if err == nil {
				fmt.Printf("switch to %s\n", highestVersion)
				err := utils.CheckoutTag(installPath, highestVersion)
				if err != nil {
					fmt.Printf("%s\n", err)
					return err
				}
			}
		}

	}

	if recurse {

		// what are it's dependencies?
		galaxyYamlFile := installPath + "/" + "galaxy.yml"
		deps, err := ReadDependencies(galaxyYamlFile)
		if err == nil {
			fmt.Printf("deps: %s\n", deps)

			for key, value := range deps {
				fmt.Printf("Key: %s, Value: %s\n", key, value)
				spec := SplitSpec(key)
				version := value
				ispec := InstallSpec{
					Server:    server,
					Namespace: spec[0],
					Name:      spec[1],
					Version:   version,
				}
				DoSpecInstall(dest, server, ispec, true)
			}
		}

	}

	return nil
}

func Install(dest string, cachedir string, server string, requirements_file string, namespace string, name string, version string, args []string) error {

	if server == "" {
		server = "https://github.com"
	}

	if requirements_file != "" {
		if utils.IsURL(requirements_file) {
			fp, err := utils.DownloadFile(requirements_file)
			if err != nil {
				fmt.Printf("ERROR %s\n", err)
			}
			fmt.Printf("tmp file %s\n", fp)
			req, err := ParseRequirementsYAMLFile(fp)
			if err != nil {
				fmt.Printf("ERROR %s\n", err)
			}
			//fmt.Printf("ds: %s\n", req)

			// Iterate through the collections and print the details
			collections := req["collections"].([]interface{})
			for _, collection := range collections {
				collectionMap := collection.(map[interface{}]interface{})
				name := collectionMap["name"].(string)
				source := collectionMap["source"].(string)
				version := collectionMap["version"].(string)

				fmt.Printf("Collection:\n  Name: %s\n  Source: %s\n  Version: %s\n", name, source, version)

				spec := SplitSpec(name)
				namespace = spec[0]
				name = spec[1]
				fmt.Printf("\tnamespace: %s\n", namespace)
				fmt.Printf("\tname: %s\n", name)

				ispec := InstallSpec{
					Server:    server,
					Namespace: namespace,
					Name:      name,
					Version:   version,
				}

				fmt.Printf("spec %s\n", ispec)
				DoSpecInstall(dest, server, ispec, false)

			}

		} else {
		}
	} else {
		fmt.Printf("collection install ... s:%s ns:%s n:%s v:%s args:%s\n", server, namespace, name, version, args)

		// parse namespace.name from args ...
		// geerlingguy.mac
		// github.com:geerlingguy.mac
		if len(args) > 0 {
			fqn := args[0]
			spec := SplitSpec(fqn)
			//fmt.Printf("spec split .. %s\n", spec)

			if len(spec) == 3 {
				server = spec[0]
				namespace = spec[1]
				name = spec[2]
			} else if len(spec) == 2 {
				namespace = spec[0]
				name = spec[1]
			}

		}

		ispec := InstallSpec{
			Server:    server,
			Namespace: namespace,
			Name:      name,
			Version:   version,
		}

		fmt.Printf("spec %s\n", ispec)
		DoSpecInstall(dest, server, ispec, true)

	}

	return nil
}
