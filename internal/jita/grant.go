package jita

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nais/cli/pkg/cli"
	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/jita/command/flag"
)

func Grant(ctx context.Context, flags *flag.GrantFlags, out cli.Output, args []string) error {
	entitlementName := args[0]
	tenantName := args[1]

	/////
	// Fetch metadata from Google
	tenantMetadata, err := gcp.FetchTenantMetadata(tenantName)
	if err != nil {
		return fmt.Errorf("fetching tenant metadata: %w", err)
	}

	entitlements, err := gcp.ListEntitlements(ctx, tenantMetadata.NaisFolderID)
	if err != nil {
		return fmt.Errorf("listing entitlements: %w", err)
	}

	entitlement := entitlements.GetByName(entitlementName)
	if entitlement == nil {
		return fmt.Errorf("entitlement with name %q does not exist for this tenant", entitlementName)
	}

	/////
	// Read remaining parameters

	stdin := bufio.NewReader(os.Stdin)
	promptedFlags := 0

	if flags.Duration == 0 {
		promptedFlags++
		fmt.Printf("How long do you need the `%s` privilege? (30m - %s) [30m]: ", entitlementName, entitlement.MaxDuration())
		text, err := stdin.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			flags.Duration = time.Minute * 30
		} else {
			flags.Duration, err = time.ParseDuration(text)
			if err != nil {
				return err
			}
			if flags.Duration < time.Minute*30 || flags.Duration > entitlement.MaxDuration() {
				return fmt.Errorf("duration must be between 30m and %s", entitlement.MaxDuration())
			}
		}
	}

	if len(flags.Reason) == 0 {
		promptedFlags++
		fmt.Print("Why do you need to elevate privileges? Please provide a human-readable description .\n")
		fmt.Print("This value is sent to the tenant, and will be read by someone.\n")
		fmt.Print("Reason: ")
		text, err := stdin.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if len(text) == 0 {
			return fmt.Errorf("you MUST specify a reason for privilege elevation")
		}
		flags.Reason = text
		fmt.Println()
	}

	fmt.Printf("*** ELEVATE PRIVILEGES ***\n")
	fmt.Println()
	fmt.Printf("Entitlement...: %s\n", entitlementName)
	fmt.Printf("Tenant........: %s\n", tenantName)
	fmt.Printf("Duration......: %s\n", flags.Duration)
	fmt.Printf("Reason........: %s\n", flags.Reason)
	fmt.Println()

	if promptedFlags > 0 {
		fmt.Printf("Are these values correct? [Y/n]: ")
		text, err := stdin.ReadString('\n')
		if err != nil {
			return err
		}
		text = strings.TrimSpace(text)
		if text != "Y" && text != "y" && text != "yes" && text != "" {
			return fmt.Errorf("exiting")
		}
	}

	fmt.Println()
	fmt.Println("Elevating privileges...")

	grant := gcp.NewGrant(flags.Duration, flags.Reason)

	err = gcp.ElevatePrivileges(ctx, *entitlement, grant)
	if err != nil {
		return fmt.Errorf("GCP error requesting grant: %w", err)
	}

	fmt.Println()
	fmt.Printf("***       YOUR PRIVILEGES HAVE BEEN ELEVATED.        ***\n")
	fmt.Printf("***   WITH GREAT POWER COMES GREAT RESPONSIBILITY.   ***\n")
	fmt.Printf("***             THINK BEFORE YOU TYPE!               ***\n")
	fmt.Println()

	return nil
}
