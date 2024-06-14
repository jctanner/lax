package collections

type CollectionVersionDetailNamespace struct {
	Name string `json:"name"`
	//MetadataSha256 string `json:"metadata_sha256"`
}

type CollectionVersionDetailArtifact struct {
	FileName string `json:"filename"`
}

type CollectionVersionDetail struct {
	Namespace   CollectionVersionDetailNamespace `json:"namespace"`
	Name        string                           `json:"name"`
	Version     string                           `json:"version"`
	DownloadUrl string                           `json:"download_url"`
	Artifact    CollectionVersionDetailArtifact  `json:"artifact"`
}
