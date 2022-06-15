package lunatics

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"math/big"
	"sync"
	"time"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// "google.golang.org/grpc/metadata"

const (
	// IBC coin identifiers
	lunaIbc = "ibc/D70B0FBF97AEB04491E9ABF4467A7F66CD6250F4382CE5192D856114B83738D2"
	ustIbc  = "ibc/4294C3DB67564CF4A0B2BFACC8415A59B38243F6FF9E288FBA34F9B4823BA16E"
)

var (
	// how many decimal places each token has: tokens = identifier/10^precision
	lunaPrecision = new(big.Float).SetInt64(1000000)
	UstPrecision  = new(big.Float).SetInt64(1000000)
)

// getBalances watches for accounts discovered and dispatches a lookup worker to get balances.
// if a balance for UST or LUNA is found, it is sent over the foundLuna or found Ust channel
// and will be saved in the appropriate .csv file.
func getBalances(ctx context.Context, wg *sync.WaitGroup, client *grpc.ClientConn, workers int, block int64, account chan string, foundLuna chan string, foundUst chan string) {

	var searched, ustCount, lunaCount int64

	go func() {
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-tick.C:
				log.Printf("checked %d accounts, %d have UST, %d have LUNA\n", searched, ustCount, lunaCount)
			case <-ctx.Done():
				log.Printf("DONE! checked %d accounts, %d have UST, %d have LUNA\n", searched, ustCount, lunaCount)
				return
			}
		}
	}()

	for i := 0; i < workers; i++ {
		go func(worker int) {
			defer wg.Done()
			log.Println("starting worker", worker)
			defer log.Printf("worker %d exiting", worker)

			bankClient := bank.NewQueryClient(client)
			var header metadata.MD
			metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", block))

			for {
				select {
				case a := <-account:
					searched += 1
					balResp, err := bankClient.AllBalances(
						metadata.AppendToOutgoingContext(context.Background(), grpctypes.GRPCBlockHeightHeader, fmt.Sprintf("%d", block)),
						&bank.QueryAllBalancesRequest{Address: a},
						grpc.Header(&header),
					)
					if err != nil {
						log.Println("lookup", a, err)
						// retry ...
						account <- a
						continue
					}
					for _, coin := range balResp.Balances {
						switch coin.Denom {

						case lunaIbc:
							flt, _, _ := new(big.Float).Parse(coin.Amount.String(), 10)
							foundLuna <- fmt.Sprintf("%s,LUNA,%s\n", a, new(big.Float).Quo(flt, lunaPrecision).String())
							lunaCount += 1

						case ustIbc:
							flt, _, _ := new(big.Float).Parse(coin.Amount.String(), 10)
							foundUst <- fmt.Sprintf("%s,UST,%s\n", a, new(big.Float).Quo(flt, UstPrecision).String())
							ustCount += 1

						}
					}
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
		}(i)
	}
}
