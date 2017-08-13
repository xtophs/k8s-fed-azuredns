package azuredns

import (
	"encoding/json"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	credentialsPath = "/.azure/credentials.json"
)

// ToJSON returns the passed item as a pretty-printed JSON string. If any JSON error occurs,
// it returns the empty string.
func ToJSON(v interface{}) (string, error) {
	j, err := json.MarshalIndent(v, "", "  ")
	return string(j), err
}

// NewServicePrincipalTokenFromCredentials creates a new ServicePrincipalToken using values of the
// passed credentials map.
// This implementation is "borrowed" from a later version of the azuresdk-for-go/arm/examples
func NewServicePrincipalTokenFromCredentials(config Config, scope string) (*adal.ServicePrincipalToken, error) {

	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.Global.TenantID)
	if err != nil {
		panic(err)
	}

	return adal.NewServicePrincipalToken(*oauthConfig, config.Global.ClientID, config.Global.Secret, scope)
}
