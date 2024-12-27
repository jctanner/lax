package laxcmd

import (
	"fmt"
	"os"

	"github.com/jctanner/lax/internal/galaxy_sync"
	"github.com/jctanner/lax/internal/repository"
	"github.com/jctanner/lax/internal/types"
	"github.com/jctanner/lax/internal/utils"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"

	"github.com/jctanner/lax/internal/collections"
	"github.com/jctanner/lax/internal/roles"
)

func SetLogLevel(kwargs *types.CmdKwargs) {
	if kwargs.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func Execute() {

	kwargs := types.CmdKwargs{}

	// pre-set verbose to false ...
	kwargs.Verbose = false

	// pre-fill the dir options ...
	defaultDestDir := utils.ExpandUser("~/.ansible")
	defaultCacheDir := utils.ExpandUser("~/.ansible/lax_cache")
	kwargs.DestDir = defaultDestDir
	kwargs.CacheDir = defaultCacheDir

	var rootCmd = &cobra.Command{Use: "cli"}

	var roleCmd = &cobra.Command{
		Use:   "role",
		Short: "Manage roles",
	}

	var collectionCmd = &cobra.Command{
		Use:   "collection",
		Short: "Manage collections",
	}

	/*
	   var repoCmd = &cobra.Command{
	       Use:   "repo",
	       Short: "Manage repositories",
	   }
	*/

	var createRepoCmd = &cobra.Command{
		Use:   "createrepo",
		Short: "Create repository metadata from a directory of artifacts",
		Run: func(cmd *cobra.Command, args []string) {
			SetLogLevel(&kwargs)
			repository.CreateRepo(&kwargs)
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Create a new role or collection",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Initialized")
		},
	}

	var collectionInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install",
		Run: func(cmd *cobra.Command, args []string) {
			SetLogLevel(&kwargs)
			if kwargs.DestDir == "" {
				kwargs.DestDir = defaultDestDir
			}
			if kwargs.CacheDir == "" {
				kwargs.CacheDir = defaultCacheDir
			}
			logrus.Debugf("INSTALL1: cachedir:%s dest:%s\n", kwargs.CacheDir, kwargs.DestDir)
			collections.Install(&kwargs, args)
		},
	}

	var roleInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install",
		Run: func(cmd *cobra.Command, args []string) {
			SetLogLevel(&kwargs)
			if kwargs.DestDir == "" {
				kwargs.DestDir = defaultDestDir
			}
			if kwargs.CacheDir == "" {
				kwargs.CacheDir = defaultCacheDir
			}
			logrus.Debugf("INSTALL1: cachedir:%s dest:%s\n", kwargs.CacheDir, kwargs.DestDir)
			roles.Install(&kwargs, args)
		},
	}

	var syncCmd = &cobra.Command{
		Use:   "galaxy-sync",
		Short: "Sync content from galaxy into a lax repo directory",
		Run: func(cmd *cobra.Command, args []string) {
			SetLogLevel(&kwargs)
			if kwargs.Server == "" || kwargs.Server == "https://console.redhat.com" {
				kwargs.Server = "https://galaxy.ansible.com"
			}
			kwargs.ApiPrefix = "/api"
			kwargs.AuthUrl = ""

			fmt.Println("###########################################")
			fmt.Printf("%v\n", kwargs)
			fmt.Println("###########################################")

			err := galaxy_sync.GalaxySync(&kwargs)
			if err != nil {
				logrus.Errorf("ERROR: %s\n", err)
			}
		},
	}

	var crcSyncCmd = &cobra.Command{
		Use:   "crc-sync",
		Short: "Sync content from console.redhat.com into a lax repo directory",
		Run: func(cmd *cobra.Command, args []string) {
			SetLogLevel(&kwargs)
			if kwargs.Server == "" || kwargs.Server == "https://galaxy.ansible.com" {
				kwargs.Server = "https://console.redhat.com"
			}
			kwargs.ApiPrefix = "/api/automation-hub"
			if kwargs.AuthUrl == "" {
				kwargs.AuthUrl = "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token"
			}
			err := galaxy_sync.GalaxySync(&kwargs)
			if err != nil {
				logrus.Errorf("ERROR: %s\n", err)
			}
		},
	}

	createRepoCmd.Flags().StringVar(&kwargs.DestDir, "dest", "", "where the files are")
	createRepoCmd.Flags().BoolVar(&kwargs.CollectionsOnly, "collections", false, "just process collections")
	createRepoCmd.Flags().BoolVar(&kwargs.RolesOnly, "roles", false, "just process roles")
	createRepoCmd.Flags().BoolVar(&kwargs.Verbose, "verbose", false, "use debug output")

	collectionInstallCmd.Flags().StringVar(&kwargs.Server, "server", "https://github.com", "server")
	collectionInstallCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	collectionInstallCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	collectionInstallCmd.Flags().StringVar(&kwargs.Version, "version", "", "version")
	collectionInstallCmd.Flags().StringVar(&kwargs.DestDir, "dest", defaultDestDir, "where to install")
	collectionInstallCmd.Flags().StringVar(&kwargs.CacheDir, "cachedir", defaultCacheDir, "where to store intermediate files")
	collectionInstallCmd.Flags().StringVarP(&kwargs.RequirementsFile, "requirements-file", "r", "", "requirements file")
	collectionInstallCmd.Flags().BoolVar(&kwargs.Verbose, "verbose", false, "use debug output")

	roleInstallCmd.Flags().StringVar(&kwargs.Server, "server", "https://github.com", "server")
	roleInstallCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	roleInstallCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	roleInstallCmd.Flags().StringVar(&kwargs.Version, "version", "", "version")
	roleInstallCmd.Flags().StringVar(&kwargs.CacheDir, "cachedir", defaultCacheDir, "where to store intermediate files")
	roleInstallCmd.Flags().StringVar(&kwargs.DestDir, "dest", defaultDestDir, "where to install")
	roleInstallCmd.Flags().BoolVar(&kwargs.Verbose, "verbose", false, "use debug output")

	syncCmd.Flags().StringVar(&kwargs.Server, "server", "https://galaxy.ansible.com", "remote server")
	syncCmd.Flags().StringVar(&kwargs.DestDir, "dest", "", "where to store the data")
	syncCmd.Flags().BoolVar(&kwargs.CollectionsOnly, "collections", false, "just sync collections")
	syncCmd.Flags().BoolVar(&kwargs.RolesOnly, "roles", false, "just sync roles")
	syncCmd.Flags().BoolVar(&kwargs.ArtifactsOnly, "artifacts", false, "just sync the artifacts")
	syncCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	syncCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	syncCmd.Flags().StringVar(&kwargs.Version, "version", "", "version")
	syncCmd.Flags().IntVar(&kwargs.DownloadConcurrency, "concurrency", 1, "concurrency")
	syncCmd.Flags().BoolVar(&kwargs.LatestOnly, "latest", false, "get only the latest version")
	syncCmd.Flags().StringVarP(&kwargs.RequirementsFile, "requirements", "r", "", "requirements file")
	syncCmd.Flags().BoolVar(&kwargs.Verbose, "verbose", false, "use debug output")
	syncCmd.MarkFlagRequired("dest")

	crcSyncCmd.Flags().StringVar(&kwargs.Server, "server", "", "remote server")
	crcSyncCmd.Flags().StringVar(&kwargs.AuthUrl, "auth_url", "https://sso.redhat.com/auth/realms/redhat-external/protocol/openid-connect/token", "auth url")
	crcSyncCmd.Flags().StringVar(&kwargs.Token, "token", "", "authentication token")
	crcSyncCmd.Flags().StringVar(&kwargs.DestDir, "dest", "", "where to store the data")
	crcSyncCmd.Flags().BoolVar(&kwargs.CollectionsOnly, "collections", false, "just sync collections")
	crcSyncCmd.Flags().BoolVar(&kwargs.RolesOnly, "roles", false, "just sync roles")
	crcSyncCmd.Flags().BoolVar(&kwargs.ArtifactsOnly, "artifacts", false, "just sync the artifacts")
	crcSyncCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	crcSyncCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	crcSyncCmd.Flags().StringVar(&kwargs.Version, "version", "", "version")
	crcSyncCmd.Flags().IntVar(&kwargs.DownloadConcurrency, "concurrency", 1, "concurrency")
	crcSyncCmd.Flags().BoolVar(&kwargs.LatestOnly, "latest", false, "get only the latest version")
	crcSyncCmd.Flags().StringVarP(&kwargs.RequirementsFile, "requirements", "r", "", "requirements file")
	crcSyncCmd.Flags().BoolVar(&kwargs.Verbose, "verbose", false, "use debug output")
	crcSyncCmd.MarkFlagRequired("dest")

	roleCmd.AddCommand(initCmd)
	roleCmd.AddCommand(roleInstallCmd)

	collectionCmd.AddCommand(initCmd)
	collectionCmd.AddCommand(collectionInstallCmd)

	//repoCmd.AddCommand(initCmd)
	//repoCmd.AddCommand(installCmd)

	rootCmd.AddCommand(createRepoCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(crcSyncCmd)
	rootCmd.AddCommand(roleCmd)
	rootCmd.AddCommand(collectionCmd)
	//rootCmd.AddCommand(repoCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
