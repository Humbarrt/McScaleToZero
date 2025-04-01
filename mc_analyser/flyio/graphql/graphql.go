package graphql

import (
	"context"
	"mc_analyser/logging"
	"mc_analyser/model"

	"github.com/machinebox/graphql"
)

func TestSSH(flyApi *model.FlyAPI) {
	client := graphql.NewClient("https://api.fly.io/graphql")

	logging.Info("Testing SSH connection to Fly.io")
	req := graphql.NewRequest(`
		mutation IssueCertificate {
            issueCertificate(
                input: {
                organizationId: "2j9m3yRpYo94bH7mGqNvyDwlZvC5bBP6a"
                appNames: ["partypeoplemc"]
                validHours: 1
                principals: ["root"]
                publicKey: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIUHRGRPh5jrpVp8lpgDqiKglbvU6lYIvoLihbrTk46u humpert@Mars"
                }
            ) {
                certificate
                clientMutationId
            }
        }
	`)

	// Set headers if needed (e.g., for authentication)
	req.Header.Set("Authorization", "Bearer "+flyApi.FLY_API_TOKEN)

	var resp struct {
		IssueCertificate struct {
			Certificate string `json:"certificate"`
		} `json:"issueCertificate"`
	}

	logging.Info("Sending request to issue certificate")
	err := client.Run(context.Background(), req, &resp)
	if err != nil {
		logging.Info("Error sending request", "error", err)
		logging.Fatal("Error fetching data", "error", err)
	}

	logging.Info("Certificate issued successfully")

	// Log the response
	logging.Info("Response", "certificate", resp.IssueCertificate.Certificate)
}
