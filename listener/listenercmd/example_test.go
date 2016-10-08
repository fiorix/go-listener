package listenercmd_test

import (
	"github.com/spf13/cobra"

	"github.com/fiorix/go-listener/listener"
	"github.com/fiorix/go-listener/listener/listenercmd"
)

func Example() {
	conf := listenercmd.NewConfig()
	cmd := &cobra.Command{
		Use: "hello",
		Run: func(cmd *cobra.Command, args []string) {
			ln, err := listener.New(conf.ListenAddr, conf.Options()...)
			if err != nil {
				panic(err)
			}
			defer ln.Close()
			// ...
		},
	}
	conf.AddFlags(cmd.Flags())
	cmd.Execute()
}
