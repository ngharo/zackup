package cmd

import "github.com/spf13/cobra"

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [host [...]]",
	Short: "Creates backups and stores them in a local per-host ZFS dataset",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = tree.Hosts()
		}

		for _, host := range args {
			job := tree.Host(host)
			if job == nil {
				log.WithField("host", host).Warn("unknown host, ignoring")
				continue
			}
			queue.Enqueue(job)
		}
		queue.Wait()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
