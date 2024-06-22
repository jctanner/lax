package laxcmd

import (
	"fmt"
	"os"

	"github.com/jctanner/lax/internal/galaxy_sync"
	"github.com/jctanner/lax/internal/repository"
	"github.com/jctanner/lax/internal/types"
	"github.com/jctanner/lax/internal/utils"

	"github.com/spf13/cobra"

	"github.com/jctanner/lax/internal/collections"
	"github.com/jctanner/lax/internal/roles"
)

func Execute() {

	kwargs := types.CmdKwargs{}

	defaultDestDir := utils.ExpandUser("~/.ansible")
	//dest = defaultDestDir
	defaultCacheDir := utils.ExpandUser("~/.ansible/lax_cache")
	//cachedir = defaultCacheDir

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
			if kwargs.DestDir == "" {
				kwargs.DestDir = defaultDestDir
			}
			if kwargs.CacheDir == "" {
				kwargs.CacheDir = defaultCacheDir
			}
			fmt.Printf("INSTALL1: cachedir:%s dest:%s\n", kwargs.CacheDir, kwargs.DestDir)
			collections.Install(&kwargs, args)
		},
	}

	var roleInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install",
		Run: func(cmd *cobra.Command, args []string) {
			if kwargs.DestDir == "" {
				kwargs.DestDir = defaultDestDir
			}
			if kwargs.CacheDir == "" {
				kwargs.CacheDir = defaultCacheDir
			}
			fmt.Printf("INSTALL1: cachedir:%s dest:%s\n", kwargs.CacheDir, kwargs.DestDir)
			roles.Install(&kwargs, args)
		},
	}

	var syncCmd = &cobra.Command{
		Use:   "galaxy-sync",
		Short: "Sync content from galaxy into a lax repo directory",
		Run: func(cmd *cobra.Command, args []string) {
			/*
			   if collections_only || (!collections_only && !roles_only) {
			       if !artifacts_only {
			           collections.SyncCollections(server, dest)
			           collections.SyncVersions(server, dest)
			       }
			       collections.SyncArtifacts(server, dest)
			   }
			   if roles_only || (!collections_only && !roles_only) {
			       roles.SyncRoles(server, dest)
			   }
			*/
			if kwargs.Server == "" {
				kwargs.Server = "https://galaxy.ansible.com"
			}
			err := galaxy_sync.GalaxySync(&kwargs)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
			}
		},
	}

	/*
		defaultDestDir := utils.ExpandUser("~/.ansible")
		dest = defaultDestDir
		//defaultDest, _ = utils.GetAbsPath(defaultDest)
		defaultCacheDir := utils.ExpandUser("~/.ansible/lax_cache")
		//defaultCacheDir, _ = utils.GetAbsPath(defaultCacheDir)
		cachedir = defaultCacheDir
	*/

	createRepoCmd.Flags().StringVar(&kwargs.DestDir, "dest", "", "where the files are")
	createRepoCmd.Flags().BoolVar(&kwargs.CollectionsOnly, "collections", false, "just process collections")
	createRepoCmd.Flags().BoolVar(&kwargs.RolesOnly, "roles", false, "just process roles")

	collectionInstallCmd.Flags().StringVar(&kwargs.Server, "server", "https://github.com", "server")
	collectionInstallCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	collectionInstallCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	collectionInstallCmd.Flags().StringVar(&kwargs.Version, "version", "", "version")
	collectionInstallCmd.Flags().StringVar(&kwargs.DestDir, "dest", defaultDestDir, "where to install")
	collectionInstallCmd.Flags().StringVar(&kwargs.CacheDir, "cachedir", defaultCacheDir, "where to store intermediate files")
	collectionInstallCmd.Flags().StringVarP(&kwargs.RequirementsFile, "requirements-file", "r", "", "requirements file")

	roleInstallCmd.Flags().StringVar(&kwargs.Server, "server", "https://github.com", "server")
	roleInstallCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	roleInstallCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	roleInstallCmd.Flags().StringVar(&kwargs.Version, "version", "", "version")
	roleInstallCmd.Flags().StringVar(&kwargs.CacheDir, "cachedir", defaultCacheDir, "where to store intermediate files")
	roleInstallCmd.Flags().StringVar(&kwargs.DestDir, "dest", defaultDestDir, "where to install")

	syncCmd.Flags().StringVar(&kwargs.Server, "server", "https://galaxy.ansible.com", "remote server")
	syncCmd.Flags().StringVar(&kwargs.DestDir, "dest", "", "where to store the data")
	syncCmd.Flags().BoolVar(&kwargs.CollectionsOnly, "collections", false, "just sync collections")
	syncCmd.Flags().BoolVar(&kwargs.RolesOnly, "roles", false, "just sync roles")
	syncCmd.Flags().BoolVar(&kwargs.ArtifactsOnly, "artifacts", false, "just sync the artifacts")
	syncCmd.Flags().StringVar(&kwargs.Namespace, "namespace", "", "namespace")
	syncCmd.Flags().StringVar(&kwargs.Name, "name", "", "name")
	syncCmd.Flags().IntVar(&kwargs.DownloadConcurrency, "concurrency", 1, "concurrency")
	syncCmd.Flags().BoolVar(&kwargs.LatestOnly, "latest", false, "get only the latest version")
	syncCmd.Flags().StringVarP(&kwargs.RequirementsFile, "requirements", "r", "", "requirements file")
	//syncCmd.MarkFlagRequired("server")
	syncCmd.MarkFlagRequired("dest")

	roleCmd.AddCommand(initCmd)
	roleCmd.AddCommand(roleInstallCmd)

	collectionCmd.AddCommand(initCmd)
	collectionCmd.AddCommand(collectionInstallCmd)

	//repoCmd.AddCommand(initCmd)
	//repoCmd.AddCommand(installCmd)

	rootCmd.AddCommand(createRepoCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(roleCmd)
	rootCmd.AddCommand(collectionCmd)
	//rootCmd.AddCommand(repoCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
