// A simple Ngrok module

package main

import (
	"context"
	"dagger/ngrok/internal/dagger"
	"strconv"
	"time"
)

type Ngrok struct{}

const (
	ngrokConfig = `
log_level: error
log: stderr
region: us
web_addr: 0.0.0.0:4040
version: "2"
`
	ngrokConfigPath = "/my/ngrok.yml"
)

// Share a Service via Ngrok
func (m *Ngrok) Share(service *dagger.Service, port int, token *dagger.Secret) *dagger.Container {
	return dag.Container().
		From("ngrok/ngrok:latest").
		WithSecretVariable("NGROK_AUTHTOKEN", token).
		WithNewFile(ngrokConfigPath, ngrokConfig).		
		WithServiceBinding("localhost", service).
		WithExposedPort(4040).
		WithExec([]string{"ngrok", "http", "--config", ngrokConfigPath, strconv.Itoa(port)})
}

// Retrieve first Ngrok public url
func (m *Ngrok) Url(ctx context.Context, apiToken *dagger.Secret) (string, error) {
	return dag.Container().
		From("alpine").
		WithSecretVariable("NGROK_APITOKEN", apiToken).
		WithExec([]string{"apk", "add", "curl", "bash", "jq"}).
		WithEnvVariable("CACHEBUST", strconv.FormatInt(time.Now().Unix(), 10)).
		WithExec([]	string{
			"bash",
			"-c",		
			`curl -X GET -H "Authorization: Bearer $NGROK_APITOKEN" ` +
			`-H "Ngrok-Version: 2" ` +
			`https://api.ngrok.com/tunnels | jq -r '.tunnels[0].public_url'`,
		}).
		Stdout(ctx)
}

// Return first Ngrok public url QR code
func (m *Ngrok) Qr(ctx context.Context, apiToken *dagger.Secret) (string, error) {
	url, err := m.Url(ctx, apiToken)
	if err != nil {
		return "", err
	}
	return dag.Qr().GenerateASCIIQr(ctx, url)
}

// Test sharing via Ngrok with Nginx
func (m *Ngrok) Test(ctx context.Context, port int, token *dagger.Secret) *dagger.Container {
	service := dag.Container().From("nginx:latest").WithExposedPort(80).AsService()
	return m.Share(service, port, token)
}
