package lunatics

import (
	"context"
	q "github.com/cosmos/cosmos-sdk/types/query"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
	"log"
)

// findAccounts requests every account known to the chain, and pushes each it finds to the found channel for processing.
func findAccounts(ctx context.Context, client *grpc.ClientConn, found chan string) {
	authClient := auth.NewQueryClient(client)
	// first query we use page 0, after this, switch to using the pagination key
	resp, err := authClient.Accounts(ctx, &auth.QueryAccountsRequest{
		Pagination: &q.PageRequest{Offset: 0, Limit: 20},
	})
	if err != nil {
		log.Fatal(err)
	}

	// not sure if pagination will be nil or 0 length, play it safe....
	for resp.Pagination.NextKey != nil && len(resp.Pagination.NextKey) != 0 {
		for i := range resp.Accounts {
			account := &auth.BaseAccount{}
			e := account.Unmarshal(resp.Accounts[i].Value)
			// some account types such as multi-sigs might not deserialize. Not sure how to deal with this.
			// it's only a handful of accounts with an error, for now, accept some failures.
			if e != nil {
				log.Println("couldn't deserialize protobuf, possibly not a auth.BaseAccount?", e)
				continue
			}
			found <- account.Address
		}
		resp, err = authClient.Accounts(ctx, &auth.QueryAccountsRequest{
			Pagination: &q.PageRequest{Key: resp.Pagination.NextKey, Limit: 100},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
