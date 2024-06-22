package types

type CmdKwargs struct {
	Server              string
	DestDir             string
	CacheDir            string
	CollectionsOnly     bool
	RolesOnly           bool
	ArtifactsOnly       bool
	Namespace           string
	Name                string
	Version             string
	LatestOnly          bool
	RequirementsFile    string
	DownloadConcurrency int
}
