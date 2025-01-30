package jita

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nais/narcos/internal/gcp"
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "jita",
		Usage:           "Just-in-time privilege escalation for tenants.",
		HideHelpCommand: true,
		Commands:        subCommands(),
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "grant",
			Usage:       "Elevate privileges for this tenant",
			UsageText:   "narc jita grant <ENTITLEMENT> <TENANT> [--duration DURATION] [--reason REASON]",
			Description: "TENANT is one of the tenants given by `narc tenant list`\nENTITLEMENT is one of `nais-view` or `nais-admin`\nDURATION is the amount of time you need privileges for, given as 0h0m\nREASON is a human-readable description of why you need to elevate privileges.",
			Flags: []cli.Flag{
				&cli.DurationFlag{
					Name:        "duration",
					HideDefault: true,
					Usage:       "How long you need privileges for.",
				},
				&cli.StringFlag{
					Name:  "reason",
					Usage: "Human-readable description of why you need to elevate privileges. This value is read by the tenant.",
				},
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if cmd.NArg() < 1 {
					return fmt.Errorf("missing required argument: ENTITLEMENT")
				}

				if cmd.NArg() < 2 {
					return fmt.Errorf("missing required argument: TENANT")
				}

				entitlementName := cmd.Args().Get(0)
				tenant := cmd.Args().Get(1)
				duration := cmd.Duration("duration")
				reason := cmd.String("reason")

				stdin := bufio.NewReader(os.Stdin)
				promptedFlags := 0

				if duration == 0 {
					promptedFlags++
					fmt.Printf("How long do you need the `%s` privilege? [30m]: ", entitlementName)
					text, err := stdin.ReadString('\n')
					if err != nil {
						return err
					}
					text = strings.TrimSpace(text)
					if len(text) == 0 {
						duration = time.Minute * 30
					} else {
						duration, err = time.ParseDuration(text)
						if err != nil {
							return err
						}
					}
					fmt.Println()
				}

				if len(reason) == 0 {
					promptedFlags++
					fmt.Print("Why do you need to elevate privileges? Please provide a human-readable description.\n")
					fmt.Print("This value is sent to the tenant, and will be read by someone.\n")
					fmt.Print("Reason: ")
					text, err := stdin.ReadString('\n')
					if err != nil {
						return err
					}
					text = strings.TrimSpace(text)
					if len(text) == 0 {
						return fmt.Errorf("you MUST specify a reason for privilege escalation")
					}
					reason = text
					fmt.Println()
				}

				fmt.Printf("*** ESCALATE PRIVILEGES ***\n")
				fmt.Println()
				fmt.Printf("Entitlement...: %s\n", entitlementName)
				fmt.Printf("Tenant........: %s\n", tenant)
				fmt.Printf("Duration......: %s\n", duration)
				fmt.Printf("Reason........: %s\n", reason)
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

				tenantFolderIDs := gcp.TenantNaisFolderIDMapping()
				if _, ok := tenantFolderIDs[tenant]; !ok {
					return fmt.Errorf("invalid tenant %q", tenant)
				}

				entitlements, err := gcp.ListEntitlements(ctx, tenantFolderIDs[tenant])
				if err != nil {
					return fmt.Errorf("GCP returned error while listing entitlements: %w", err)
				}

				entitlement := entitlements.GetByName(entitlementName)
				if entitlement == nil {
					return fmt.Errorf("entitlement with name %q does not exist for this tenant", entitlementName)
				}

				fmt.Printf("Roles granted.: %v\n", entitlement.Roles())
				fmt.Printf("Max duration..: %s\n", entitlement.MaxDuration())

				fmt.Println()
				fmt.Println("FIXME: this isn't really implemented yet")

				// FIXME: make a request with entitlement.Name

				return nil
			},
		},
	}
}
