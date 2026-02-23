package jita

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/narcos/internal/gcp"
	"github.com/nais/narcos/internal/jita/command/flag"
)

func Revoke(ctx context.Context, flags *flag.Revoke, entitlementName, tenantName string) error {
	username, err := gcp.GCloudActiveUser(ctx)
	if err != nil {
		return fmt.Errorf("getting active user: %w", err)
	}

	tenantMetadata, err := gcp.FetchTenantMetadata(ctx, tenantName)
	if err != nil {
		return fmt.Errorf("fetching tenant metadata: %w", err)
	}

	entitlements, err := gcp.ListEntitlements(ctx, tenantMetadata.NaisFolderID)
	if err != nil {
		return fmt.Errorf("listing entitlements: %w", err)
	}

	entitlement := gcp.GetEntitlementByName(entitlements, entitlementName)
	if entitlement == nil {
		return fmt.Errorf("entitlement with name %q does not exist for this tenant", entitlementName)
	}

	grants, err := gcp.ListActiveGrants(ctx, entitlement.GetName(), username)
	if err != nil {
		return fmt.Errorf("listing active grants: %w", err)
	}

	if len(grants) == 0 {
		fmt.Printf("No active grants found for entitlement %q on tenant %q.\n", entitlementName, tenantName)
		return nil
	}

	stdin := bufio.NewReader(os.Stdin)
	promptedFlags := 0

	if len(flags.Reason) == 0 {
		promptedFlags++
		fmt.Print("Why are you revoking this grant? (optional, press enter to skip)\n")
		fmt.Print("Reason: ")
		text, err := stdin.ReadString('\n')
		if err != nil {
			return err
		}
		flags.Reason = strings.TrimSpace(text)
		fmt.Println()
	}

	fmt.Printf("*** REVOKE PRIVILEGES ***\n")
	fmt.Println()
	fmt.Printf("Entitlement...: %s\n", entitlementName)
	fmt.Printf("Tenant........: %s\n", tenantName)
	fmt.Printf("Grants........: %d active grant(s)\n", len(grants))
	for _, g := range grants {
		fmt.Printf("  - %s remaining\n", g.TimeRemaining())
	}
	if len(flags.Reason) > 0 {
		fmt.Printf("Reason........: %s\n", flags.Reason)
	}
	fmt.Println()

	if promptedFlags > 0 {
		fmt.Printf("Are you sure you want to revoke? [Y/n]: ")
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
	fmt.Println("Revoking privileges...")

	for _, g := range grants {
		err := gcp.RevokeGrant(ctx, g.GetName(), flags.Reason)
		if err != nil {
			return fmt.Errorf("revoking grant: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("Privileges have been revoked.")

	return nil
}
