package lax

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "lax/internal/repository"
    //"lax/internal/roles"
    "lax/internal/collections"
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

func Execute() {
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
            repository.CreateRepo(dest)
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
            collections.Install(dest, cachedir, server, requirements_file, namespace, name, version, args)
        },
    }

    var roleInstallCmd = &cobra.Command{
        Use:   "install",
        Short: "Install",
        Run: func(cmd *cobra.Command, args []string) {
            //collections.Install(server, namespace, name, version, args)
        },
    }

    var syncCmd = &cobra.Command{
        Use:   "sync",
        Short: "Sync",
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
        },
    }

    createRepoCmd.Flags().StringVar(&dest, "dest", "", "where the files are")

    collectionInstallCmd.Flags().StringVar(&server, "server", "https://github.com", "server")
    collectionInstallCmd.Flags().StringVar(&namespace, "namespace", "", "namespace")
    collectionInstallCmd.Flags().StringVar(&name, "name", "", "name")
    collectionInstallCmd.Flags().StringVar(&version, "version", "", "version")
    collectionInstallCmd.Flags().StringVar(&dest, "dest", "", "where to install")
    collectionInstallCmd.Flags().StringVar(&cachedir, "cachedir", "~/.cache/galaxy", "where to store intermediate files")
    collectionInstallCmd.Flags().StringVarP(&requirements_file, "requirements-file", "r", "", "requirements file")

    roleInstallCmd.Flags().StringVar(&server, "server", "https://github.com", "server")
    roleInstallCmd.Flags().StringVar(&namespace, "namespace", "", "namespace")
    roleInstallCmd.Flags().StringVar(&name, "name", "", "name")
    roleInstallCmd.Flags().StringVar(&version, "version", "", "version")
    roleInstallCmd.Flags().StringVar(&cachedir, "cachedir", "~/.cache/galaxy", "where to store intermediate files")

    syncCmd.Flags().StringVar(&server, "server", "", "remote server")
    syncCmd.Flags().StringVar(&dest, "dest", "", "where to store the data")
    syncCmd.Flags().BoolVar(&collections_only, "collections", false, "just sync collections")
    syncCmd.Flags().BoolVar(&roles_only, "roles", false, "just sync roles")
    syncCmd.Flags().BoolVar(&artifacts_only, "artifacts", false, "just sync the artifacts")
    syncCmd.MarkFlagRequired("server")
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

