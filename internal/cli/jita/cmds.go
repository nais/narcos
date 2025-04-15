package jita

import (
	"bufio"
	"context"
	"fmt"
	"github.com/nais/narcos/internal/gcp"
	"github.com/urfave/cli/v3"
	"os"
	"strings"
	"sync"
	"time"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:            "jita",
		Usage:           "Just-in-time privilege elevation for tenants.",
		HideHelpCommand: true,
		Commands:        subCommands(),
	}
}

type YesNoIcon bool

func (yn YesNoIcon) String() string {
	if yn {
		return "✅"
	} else {
		return "⛔"
	}
}

func subCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "list",
			Usage:     "List active and potential privilege elevations",
			UsageText: "narc jita list <TENANT>",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Aliases:     []string{"v"},
					Name:        "verbose",
					HideDefault: true,
					Usage:       "display roles contained in each entitlement",
				},
			},
			Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
				err := gcp.ValidateUserLogin(ctx)
				if err != nil {
					return ctx, fmt.Errorf("checking valid user login: %w", err)
				}
				return ctx, err
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {

				userName, err := gcp.GCloudActiveUser(ctx)
				if err != nil {
					return err
				}

				var tenants []string
				if cmd.NArg() == 0 {
					tenantBuckets, err := gcp.FetchAllTenantNames()
					if err != nil {
						panic(fmt.Errorf("Failed parsing xml:, %v", err))
					}
					for _, tenant := range tenantBuckets {
						if !strings.HasSuffix(tenant.Name, ".json") {
							continue
						}
						tenants = append(tenants, strings.TrimSuffix(tenant.Name, ".json"))
					}
				} else {
					for index := range cmd.NArg() {
						tenants = append(tenants, cmd.Args().Get(index))
					}
				}
				if cmd.Bool("verbose") {
					fmt.Printf("Tenant                    Entitlement           Granted  Remaining  Max. duration\n")
					fmt.Printf("---------------------------------------------------------------------------------\n")
				}

				var wg sync.WaitGroup
				var errCh = make(chan error, len(tenants))
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

						for _, ent := range entitlements.Entitlements {
							var hasGrants YesNoIcon
							var timeRemaining string

							grants, err := ent.ListActiveGrants(ctx, userName)
							if err != nil {
								errCh <- err
								return
							} else if len(grants) > 0 {
								hasGrants = true
								timeRemaining = grants[0].TimeRemaining().String()
							}

							fmt.Printf("\r%-24s  %-20s  %-6s  %-9s  %-9s\n",
								tenantName,
								ent.ShortName(),
								hasGrants,
								timeRemaining,
								ent.MaxDuration(),
							)
							if cmd.Bool("verbose") {
								for _, role := range ent.Roles() {
									fmt.Printf("           `- %s\n", role)
								}
							}
						}

						return
					}(tenantName)
				}
				wg.Wait()
				select {
				case err := <-errCh:
					return err
				default:
					return nil
				}
			},
		},
		{
			Name:        "grant",
			Usage:       "Elevate privileges for this tenant",
			UsageText:   "narc jita grant <ENTITLEMENT> <TENANT> [--duration DURATION] [--reason REASON]",
			Description: "TENANT is one of the tenants given by `narc tenant list`\nENTITLEMENT is one the entitlements given by `narc jita list <TENANT>`\nDURATION is the amount of time you need privileges for, given as 0h0m\nREASON is a human-readable description of why you need to elevate privileges.",
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
			Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
				return ctx, gcp.ValidateUserLogin(ctx)
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if cmd.NArg() < 1 {
					return fmt.Errorf("syntax: %s", cmd.UsageText)
				}

				if cmd.NArg() < 2 {
					return fmt.Errorf("syntax: %s", cmd.UsageText)
				}

				entitlementName := cmd.Args().Get(0)
				tenantName := cmd.Args().Get(1)
				duration := cmd.Duration("duration")
				reason := cmd.String("reason")

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

				if duration == 0 {
					promptedFlags++
					fmt.Printf("How long do you need the `%s` privilege? (30m - %s) [30m]: ", entitlementName, entitlement.MaxDuration())
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
						if duration < time.Minute*30 || duration > entitlement.MaxDuration() {
							return fmt.Errorf("duration must be between 30m and %s", entitlement.MaxDuration())
						}
					}
				}

				if len(reason) == 0 {
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
					reason = text
					fmt.Println()
				}

				fmt.Printf("*** ELEVATE PRIVILEGES ***\n")
				fmt.Println()
				fmt.Printf("Entitlement...: %s\n", entitlementName)
				fmt.Printf("Tenant........: %s\n", tenantName)
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
				fmt.Println("Elevating privileges...")

				grant := gcp.NewGrant(duration, reason)

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
			},
		},
	}
}
