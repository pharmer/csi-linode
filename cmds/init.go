package cmds

import (
	"github.com/golang/glog"
	"github.com/pharmer/csi-linode/cmds/options"
	"github.com/pharmer/csi-linode/driver"
	"github.com/spf13/cobra"
)

func NewCmdInit() *cobra.Command {
	cfg := options.NewConfig()
	cmd := &cobra.Command{
		Use:               "init",
		Short:             "Initializes the driver.",
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			drv, err := driver.NewDriver(cfg.Endpoint, cfg.Token, cfg.Region, cfg.NodeName)
			if err != nil {
				glog.Fatalln(err)
			}

			if err := drv.Run(); err != nil {
				glog.Fatalln(err)
			}
		},
	}
	cfg.AddFlags(cmd.Flags())
	return cmd
}
