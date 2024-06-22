package galaxy_sync

import "github.com/sirupsen/logrus"

func syncCollections(server string, dest string, apiClient CachedGalaxyClient, namespace string, name string, latest_only bool) ([]CollectionVersionDetail, error) {
	// iterate roles ...
	collections, err := apiClient.GetCollections(namespace, name, latest_only)
	if err != nil {
		logrus.Errorf("Error fetching collections: %v", err)
	}
	return collections, nil
}
