package testhelpers

import "strconv"

type FakeTokenFetcher struct {
    NumCalls int
}

func (tokenFetcher *FakeTokenFetcher) FetchAuthToken() string {
    tokenFetcher.NumCalls++
    return "auth token " + strconv.Itoa(tokenFetcher.NumCalls)
}
