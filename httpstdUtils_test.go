package netutils

import (
	"fmt"
	"testing"
)

func TestHttpGet(t *testing.T) {
	client := HttpClientNewClient(HttpClientNewTransPort())

	httpReps, body, err := HttpGet(client, "https://ifconfig.me", nil)
	if err != nil {
		t.Errorf("Error when do HttpGET %v", err)
	}
	fmt.Printf("Response string body: %s\n", string(body))
	t.Logf("Response string body: %s", string(body))
	t.Logf("Response Status: %s", httpReps.Status)
}

/* func BenchmarkHttpGet(b *testing.B) {
	client := HttpClientNewClient(HttpClientNewTransPort())
	for range b.N {
		httpReps, body, err := HttpGet(client, "https://ifconfig.me", nil)
		if err != nil {
			fmt.Printf("Response string body: %s\n", string(body))
			fmt.Errorf("Error when do HttpGET %v", err)
		}
		fmt.Printf("Response string body: %s\n", string(body))
		fmt.Printf("Response Status: %s", httpReps.Status)
	}
}
*/
