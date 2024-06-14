package galaxy_sync

import "log"

func syncCollections(server string, dest string, apiClient CachedGalaxyClient, namespace string, name string, latest_only bool) ([]CollectionVersionDetail, error) {
	// iterate roles ...
	collections, err := apiClient.GetCollections(namespace, name, latest_only)
	if err != nil {
		log.Fatalf("Error fetching collections: %v", err)
	}
	return collections, nil
}
