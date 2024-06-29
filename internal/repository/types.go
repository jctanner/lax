package repository

/***************************************************************
GLOBAL
***************************************************************/

type RepoMeta struct {
	Date                string       `json:"date"`
	CollectionManifests RepoMetaFile `json:"collection_manifests"`
	CollectionFiles     RepoMetaFile `json:"collection_files"`
	RoleManifests       RepoMetaFile `json:"role_manifests"`
	RoleFiles           RepoMetaFile `json:"role_files"`
}

type RepoMetaFile struct {
	Date     string `json:"date"`
	Filename string `json:"filename"`
}

/***************************************************************
COLLECTIONS
***************************************************************/

type CollectionManifest struct {
	CollectionInfo CollectionInfo `json:"collection_info"`
}

type CollectionInfo struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type CollectionFilesMeta struct {
	Files []CollectionFileInfo `json:"files"`
}

type CollectionFileInfo struct {
	Name           string `json:"name"`
	FType          string `json:"ftype"`
	CheckSumType   string `json:"chksum_type"`
	CheckSumSHA256 string `json:"chksum_sha256"`
}

type CollectionCachedFileInfo struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	FileName       string `json:"filename"`
	FileType       string `json:"filetype"`
	CheckSumSHA256 string `json:"chksum_sha256"`
}

/***************************************************************
ROLES
***************************************************************/

type RoleCachedFileInfo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Version   string `json:"version"`
	FileName  string `json:"filename"`
	FileType  string `json:"filetype"`
	//CheckSumType   string `json:"chksum_type"`
	//CheckSumSHA256 string `json:"chksum_sha256"`
}
