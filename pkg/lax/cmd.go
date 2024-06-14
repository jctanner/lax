package laxcmd

import (
	"fmt"
	"os"

	"github.com/jctanner/lax/internal/galaxy_sync"
	"github.com/jctanner/lax/internal/repository"
	"github.com/jctanner/lax/internal/utils"

	"github.com/spf13/cobra"

	"github.com/jctanner/lax/internal/collections"
	"github.com/jctanner/lax/internal/roles"
)

var server string
var cachedir string
var dest string
var collections_only bool
var artifacts_only bool
var roles_only bool
var namespace string
var name string
var version string
var requirements_file string

var download_concurrency int
var latest_only bool

func Execute() {

	defaultDestDir := utils.ExpandUser("~/.ansible")
	dest = defaultDestDir
	defaultCacheDir := utils.ExpandUser("~/.ansible/lax_cache")
	cachedir = defaultCacheDir

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
			repository.CreateRepo(dest, roles_only, collections_only)
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
			if dest == "" {
				dest = defaultDestDir
			}
			if cachedir == "" {
				cachedir = defaultCacheDir
			}
			fmt.Printf("INSTALL1: cachedir:%s dest:%s\n", cachedir, dest)
			collections.Install(dest, cachedir, server, requirements_file, namespace, name, version, args)
		},
	}

	var roleInstallCmd = &cobra.Command{
		Use:   "install",
		Short: "Install",
		Run: func(cmd *cobra.Command, args []string) {
			if dest == "" {
				dest = defaultDestDir
			}
			if cachedir == "" {
				cachedir = defaultCacheDir
			}
			fmt.Printf("INSTALL1: cachedir:%s dest:%s\n", cachedir, dest)
			roles.Install(dest, cachedir, server, requirements_file, namespace, name, version, args)
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
			if server == "" {
				server = "https://galaxy.ansible.com"
			}
			err := galaxy_sync.GalaxySync(server, dest, download_concurrency, collections_only, roles_only, latest_only, namespace, name, requirements_file)
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

	createRepoCmd.Flags().StringVar(&dest, "dest", "", "where the files are")
	createRepoCmd.Flags().BoolVar(&collections_only, "collections", false, "just process collections")
	createRepoCmd.Flags().BoolVar(&roles_only, "roles", false, "just process roles")

	collectionInstallCmd.Flags().StringVar(&server, "server", "https://github.com", "server")
	collectionInstallCmd.Flags().StringVar(&namespace, "namespace", "", "namespace")
	collectionInstallCmd.Flags().StringVar(&name, "name", "", "name")
	collectionInstallCmd.Flags().StringVar(&version, "version", "", "version")
	collectionInstallCmd.Flags().StringVar(&dest, "dest", defaultDestDir, "where to install")
	collectionInstallCmd.Flags().StringVar(&cachedir, "cachedir", defaultCacheDir, "where to store intermediate files")
	collectionInstallCmd.Flags().StringVarP(&requirements_file, "requirements-file", "r", "", "requirements file")

	roleInstallCmd.Flags().StringVar(&server, "server", "https://github.com", "server")
	roleInstallCmd.Flags().StringVar(&namespace, "namespace", "", "namespace")
	roleInstallCmd.Flags().StringVar(&name, "name", "", "name")
	roleInstallCmd.Flags().StringVar(&version, "version", "", "version")
	roleInstallCmd.Flags().StringVar(&cachedir, "cachedir", defaultCacheDir, "where to store intermediate files")
	roleInstallCmd.Flags().StringVar(&dest, "dest", defaultDestDir, "where to install")

	syncCmd.Flags().StringVar(&server, "server", "https://galaxy.ansible.com", "remote server")
	syncCmd.Flags().StringVar(&dest, "dest", "", "where to store the data")
	syncCmd.Flags().BoolVar(&collections_only, "collections", false, "just sync collections")
	syncCmd.Flags().BoolVar(&roles_only, "roles", false, "just sync roles")
	syncCmd.Flags().BoolVar(&artifacts_only, "artifacts", false, "just sync the artifacts")
	syncCmd.Flags().StringVar(&namespace, "namespace", "", "namespace")
	syncCmd.Flags().StringVar(&name, "name", "", "name")
	syncCmd.Flags().IntVar(&download_concurrency, "concurrency", 1, "concurrency")
	syncCmd.Flags().BoolVar(&latest_only, "latest", false, "get only the latest version")
	syncCmd.Flags().StringVarP(&requirements_file, "requirements", "r", "", "requirements file")
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
