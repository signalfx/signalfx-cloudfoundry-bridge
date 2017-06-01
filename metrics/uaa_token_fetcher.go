package metrics

import (
    "log"
    "strings"

    "golang.org/x/oauth2"

    "github.com/signalfx/uaago"
)

type UAATokenFetcher struct {
    UaaUrl        string
    Username      string
    Password      string
    SSLSkipVerify bool
}

func (uaa *UAATokenFetcher) FetchAuthToken() string {
    uaaClient, err := uaago.NewClient(uaa.UaaUrl)
    if err != nil {
        log.Fatalf("Error creating uaa client: %s", err.Error())
    }

    var authToken string
    log.Printf("Getting UAA token...")
    authToken, err = uaaClient.GetAuthToken(uaa.Username, uaa.Password, uaa.SSLSkipVerify)
    if err != nil {
        log.Fatalf("Error getting oauth token at %s for %s: %s",
                   uaa.UaaUrl, uaa.Username, err.Error())
    }
    return authToken
}

// Satisfy the oauth2.TokenSource interface
func (uaa *UAATokenFetcher) Token() (*oauth2.Token, error) {
    token := strings.Replace(uaa.FetchAuthToken(), "bearer ", "", 1)
    return &oauth2.Token{
        AccessToken: token,
        TokenType: "Bearer",
    }, nil
}
