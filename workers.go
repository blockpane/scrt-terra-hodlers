package lunatics

import (
	"context"
	"log"
	"sync"
	"time"
)

// "google.golang.org/grpc/metadata"

const (
	// IBC coin identifiers
	lunaIbc = "" // TODO: get IBC coin for $LLUNA
	ustIbc  = "" // TODO: get IBC coin for $UST

	// how many decimal places each token has: tokens = identifier/10^precision
	lunaPrecision = 6
	UstPrecision  = 6
)

// getBalances watches for accounts discovered and dispatches a lookup worker to get balances.
// if a balance for UST or LUNA is found, it is sent over the foundLuna or found Ust channel
// and will be saved in the appropriate .csv file.
func getBalances(ctx context.Context, wg *sync.WaitGroup, workers int, block int64, account chan string, foundLuna chan string, foundUst chan string) {

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
			for {
				select {
				case <-account:
					searched += 1
					//TODO: lookup all bank balances and match on UST/LUNA then send via channel
				case <-ctx.Done():
					return
				default:
					continue
				}
			}
		}(i)
	}
}
