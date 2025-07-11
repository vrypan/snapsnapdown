package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/vrypan/snapdown/downloader"
	"github.com/vrypan/snapdown/ui"
)

var shards []int

var extractCmd = &cobra.Command{
	Use:     "extract <input dir> <output dir>",
	Aliases: []string{"x"},
	Short:   "Extract downloaded snapshot",
	Long: `
If you downloaded the snapshot in ./snapshot you will probably
want to run:
		snapdown extract ./snapshot .rocks
to extract the files in .rocks. Then you can start your node.

WARNING! Files in <destination dir> will be overwritten!
	`,
	Run: extractRun,
}

func init() {
	rootCmd.AddCommand(extractCmd)
	extractCmd.Flags().IntSliceVar(&shards, "shards", []int{0, 1, 2}, "List of shard indices (e.g. --shard=0,1,2)")
	extractCmd.Flags().Bool("no-tty", false, "Plain text output")
}

func extractRun(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		fmt.Println("Provide input and output dirs")
		os.Exit(1)
	}

	if !downloader.HasTarInPath() {
		fmt.Println("Error: 'tar' not found in PATH.")
		os.Exit(1)
	}

	srcDir := args[0]
	dstDir := args[1]

	progressCh := make(chan downloader.XUpdMsg, 1000)
	notty, _ := cmd.Flags().GetBool("no-tty")

	fmt.Printf("\nExtracting Snapshot [%s] -> [%s]\n\n", srcDir, dstDir)

	go func() {
		for _, shard := range shards {
			downloader.ExtractWithNativeTar(srcDir, dstDir, shard, progressCh)
		}
		progressCh <- downloader.XUpdMsg{Quit: true}
	}()

	const maxShard = 2

	if notty {
		runNoTtyExtraction(maxShard, progressCh)
	} else {
		runTtyExtraction(maxShard, progressCh)
	}
}

func runNoTtyExtraction(numShards int, progressCh chan downloader.XUpdMsg) {
	model := ui.NewNoTtyExtract(numShards, progressCh)
	model.Run()
	if len(model.Errors) > 0 {
		os.Exit(1)
	}
}

func runTtyExtraction(numShards int, progressCh chan downloader.XUpdMsg) {
	model := ui.NewTtyExtract(numShards, progressCh)
	prog := tea.NewProgram(model)

	finalModel, err := prog.Run()
	if err != nil {
		fmt.Println("error:", err)
	}

	extractModel := finalModel.(ui.TtyExtract)
	if len(extractModel.Errors) > 0 {
		for _, e := range extractModel.Errors {
			fmt.Println(e)
		}
		os.Exit(1)
	}
}
