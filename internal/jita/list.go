package jita

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/jita/command/flag"
)

func List(ctx context.Context, flags *flag.ListFlags, out cli.Output, args []string) error {
	userName, err := gcp.GCloudActiveUser(ctx)
	if err != nil {
		return err
	}

	var tenants []string
	if len(args) == 0 {
		tenantBuckets, err := gcp.FetchAllTenantNames()
		if err != nil {
			panic(fmt.Errorf("failed parsing xml:, %v", err))
		}
		for _, tenant := range tenantBuckets {
			if !strings.HasSuffix(tenant.Name, ".json") {
				continue
			}
			tenants = append(tenants, strings.TrimSuffix(tenant.Name, ".json"))
		}
	} else {
		tenants = args
	}

	if flags.IsVerbose() {
		fmt.Printf("Tenant                    Entitlement           Granted  Remaining  Max. duration\n")
		fmt.Printf("---------------------------------------------------------------------------------\n")
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(tenants))
	defer close(errCh)

	type OutputItem struct {
		TenantName    string
		EntShortName  string
		HasGrants     YesNoIcon
		TimeRemaining string
		MaxDuration   time.Duration
		Roles         []string
		Verbose       bool
	}

	var outputMutex sync.Mutex
	var outputs []OutputItem

	for _, tenantName := range tenants {
		wg.Add(1)

		go func(tenant string) {
			defer wg.Done()
			tenantMetadata, err := gcp.FetchTenantMetadata(tenantName)
			if err != nil {
				errCh <- fmt.Errorf("GCP error fetching tenant metadata: %w", err)
				return
			}

			entitlements, err := gcp.ListEntitlements(ctx, tenantMetadata.NaisFolderID)
			if err != nil {
				errCh <- fmt.Errorf("GCP error listing entitlements: %w", err)
				return
			}

			var entitlementWaitGroup sync.WaitGroup
			var tenantOutputs []OutputItem

			for _, ent := range entitlements.Entitlements {
				entitlementWaitGroup.Add(1)
				go func(userName string) {
					defer entitlementWaitGroup.Done()

					output := OutputItem{
						TenantName:   tenant,
						EntShortName: ent.ShortName(),
						MaxDuration:  ent.MaxDuration(),
						Verbose:      flags.IsVerbose(),
					}

					if output.Verbose {
						output.Roles = ent.Roles()
					}
					grants, err := ent.ListActiveGrants(ctx, userName)
					if err != nil {
						errCh <- fmt.Errorf("fetchin active grants: %w", err)
						return
					} else if len(grants) > 0 {
						output.HasGrants = true
						output.TimeRemaining = grants[0].TimeRemaining().String()
					}

					outputMutex.Lock()
					tenantOutputs = append(tenantOutputs, output)
					outputMutex.Unlock()
				}(userName)
			}
			entitlementWaitGroup.Wait()

			outputMutex.Lock()
			outputs = append(outputs, tenantOutputs...)
			outputMutex.Unlock()
		}(tenantName)
	}
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		// sort on time and then sort on tenantname, active grants go on top.
		sort.Slice(outputs, func(i, j int) bool {
			if outputs[i].TimeRemaining != outputs[j].TimeRemaining {
				return outputs[i].TimeRemaining > outputs[j].TimeRemaining
			}
			return outputs[i].TenantName < outputs[j].TenantName
		})

		for _, out := range outputs {
			fmt.Printf("%-24s  %-20s  %-6s  %-9s  %-9s\n",
				out.TenantName,
				out.EntShortName,
				out.HasGrants,
				out.TimeRemaining,
				out.MaxDuration,
			)
			if out.Verbose {
				for _, role := range out.Roles {
					fmt.Printf("           `- %s\n", role)
				}
			}
		}
		return nil
	}
}
