package graphql

import (
	"context"
	"fmt"
	"mc_analyser/logging"

	"github.com/machinebox/graphql"
)

func GetBilling(accessToken string) {
	client := graphql.NewClient("https://api.fly.io/graphql")

	// Create a request with the GraphQL query
	req := graphql.NewRequest(`
        query {
            personalOrganization {
                id
                name
                creditBalance
                billingStatus
                billable
            }
            apps{
                nodes{
                    name
                    status
                    usage{
                        interval
                        requestsCount
                        totalAppExecS
                        totalDataOutGB
                        ts
                    }
                }
            }
        }
    `)

	// Set the query variables
	req.Var("slug", "personal") // Replace with your actual organization slug

	// Add the authorization header
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the request
	var resp struct {
		PersonalOrganization struct {
			ID                     string `json:"id"`
			Name                   string `json:"name"`
			CreditBalance          int64  `json:"creditBalance"`
			CreditBalanceFormatted string `json:"creditBalanceFormatted"`
			BillingStatus          string `json:"billingStatus"`
			Billable               bool   `json:"billable"`
		} `json:"personalOrganization"`
		Apps struct {
			Nodes []struct {
				Name     string `json:"name"`
				Status   string `json:"status"`
				AppUsage []struct {
					Interval       string  `json:"interval"`
					RequestsCount  int     `json:"requestsCount"`
					TotalAppExecS  int     `json:"totalAppExecS"`
					TotalDataOutGB float64 `json:"totalDataOutGB"`
					Ts             string  `json:"ts"`
				} `json:"usage"`
			} `json:"nodes"`
		} `json:"apps"`
	}

	// Execute the request
	err := client.Run(context.Background(), req, &resp)
	if err != nil {
		logging.Fatal("Error fetching data", "error", err)
	}

	// Process and display the response
	fmt.Printf("Personal Organization ID: %s\n", resp.PersonalOrganization.ID)
	fmt.Printf("Personal Organization Name: %s\n", resp.PersonalOrganization.Name)
	fmt.Printf("Personal CreditBalance: %d\n", resp.PersonalOrganization.CreditBalance)
	fmt.Printf("Personal CreditBalanceFormatted: %s\n", resp.PersonalOrganization.CreditBalanceFormatted)
	fmt.Printf("Personal Billing Status: %s\n", resp.PersonalOrganization.BillingStatus)
	fmt.Printf("Personal Is Billable: %t\n", resp.PersonalOrganization.Billable)

	fmt.Println("Apps:")
	for _, app := range resp.Apps.Nodes {
		fmt.Printf("Name: %s, Status: %s\n", app.Name, app.Status)
		fmt.Printf("Usage: %v\n", app.AppUsage)
		fmt.Println()
	}
}

// Initialize a new GraphQL client
