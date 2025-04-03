package temporal

import (
	"encore.dev/rlog"
	"go.temporal.io/sdk/client"
)

var Client client.Client

func init() {
	var err error
	Client, err = client.NewClient(client.Options{
		HostPort: "localhost:7233",
	})

	if err != nil {
		rlog.Error("Failed to create Temporal Client", "error", err)
		panic(err)
	}
}
