package table

import (
	"context"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/kolide/osquery-go"
	"github.com/kolide/osquery-go/plugin/table"
	"github.com/pkg/errors"
)

func EmailAddresses(client *osquery.ExtensionManagerClient, logger log.Logger) *table.Plugin {
	columns := []table.ColumnDefinition{
		table.TextColumn("email"),
		table.TextColumn("domain"),
	}
	t := &emailAddressesTable{
		onePasswordAccountConfig: &onePasswordAccountConfig{client: client, logger: logger},
		chromeUserProfilesTable:  &chromeUserProfilesTable{client: client, logger: logger},
	}
	return table.NewPlugin("kolide_email_addresses", columns, t.generateEmailAddresses)
}

type emailAddressesTable struct {
	onePasswordAccountConfig *onePasswordAccountConfig
	chromeUserProfilesTable  *chromeUserProfilesTable
}

func (t *emailAddressesTable) generateEmailAddresses(ctx context.Context, queryContext table.QueryContext) ([]map[string]string, error) {
	var results []map[string]string

	// add results from chrome profiles
	chromeResults, err := t.chromeUserProfilesTable.generate(ctx, queryContext)
	if err != nil {
		return nil, errors.Wrap(err, "get email addresses from chrome state")
	}
	for _, result := range chromeResults {
		email := result["email"]
		// chrome profiles don't require an email skip ones without emails
		if email == "" {
			continue
		}
		results = addEmailToResults(email, results)
	}

	// add results from 1password
	onePassResults, err := t.onePasswordAccountConfig.generate(ctx, queryContext)
	if err != nil {
		return nil, errors.Wrap(err, "adding email results from 1password config")
	}
	results = append(results, onePassResults...)

	return results, nil
}

func emailDomain(email string) string {
	parts := strings.Split(email, "@")
	switch len(parts) {
	case 0:
		return email
	default:
		return parts[len(parts)-1]
	}
}

func addEmailToResults(email string, results []map[string]string) []map[string]string {
	return append(results, map[string]string{
		"email":  email,
		"domain": emailDomain(email),
	})
}
