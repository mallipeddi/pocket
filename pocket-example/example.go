package main

import (
	"fmt"
	"github.com/mallipeddi/pocket"
	"log"
)

const (
	appConsumerKey      string = "<your-consumer-key>"
	postAuthRedirectUri string = "<your-redirect-url>"
	accessToken         string = "<your-access-token-optional>"
	username            string = "<your-username-optional>"
)

func authenticate(client *pocket.Client) {
	requestToken, err := client.NewRequestToken(postAuthRedirectUri)
	if err != nil {
		log.Fatalf("error fetching request token: %s", err)
	}
	log.Println("fetched request token: ", requestToken)

	log.Println("Visit uri to authorize this app: ",
		client.GetAuthorizationUrl(requestToken, postAuthRedirectUri))

	fmt.Print("Press any key after authorizing")
	var dummy string
	fmt.Scanf("%s", &dummy)

	if err := client.FetchAccessToken(requestToken); err != nil {
		log.Fatalf("error fetching access token: %s", err)
	}
}

func main() {
	log.Println("libpocket example app")

	var client *pocket.Client

	if len(accessToken) <= 0 {
		client = pocket.NewClient(appConsumerKey)
		authenticate(client)
	} else {
		client = pocket.NewClientWithAccessToken(appConsumerKey, accessToken, username)
	}

	log.Printf("access token: %s (for user %s)\n", client.AccessToken, client.Username)

	req := pocket.NewRetrieveRequest().Count(5)
	m, err := client.Retrieve(req)
	if err != nil {
		log.Fatalf("error in retrieve: %s", err)
	}
	log.Printf("retrieve response: %s\n", m)

	req2 := new(pocket.AddRequest)
	req2.SetUrl("http://blog.kodekabuki.com")
	req2.SetTitle("kodekabuki")
	m2, err := client.Add(req2)
	if err != nil {
		log.Fatal("error in add: %s", err)
	}
	log.Printf("add response: %s\n", m2)

	req3 := new(pocket.ModifyRequest)
	action := pocket.Action{Kind: pocket.ActionFavorite, Params: map[string]string{"item_id":"<some-item-id>"}}
	req3.AddAction(action)
	m3, err := client.Modify(req3)
	if err != nil {
		log.Fatal("error in modify: %s", err)
	}
	log.Printf("modify response: %s\n", m3)
}
