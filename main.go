package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var domainName, email, haproxyCertsLocation, letsEncryptCertsLocation, certBotPort, s3EndPoint, s3BucketName, s3AccessKey, s3PrivateKey string
var obtaining, renew, combine, transfer, puller, s3SSLEnabled bool

func main() {
	app := &cli.App{
		Name:    "certm",
		Version: "v1.0.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "domain_name",
				Usage:       "CERTM_DOMAIN_NAME",
				EnvVars:     []string{"CERTM_DOMAIN_NAME"},
				Destination: &domainName,
			},
			&cli.StringFlag{
				Name:        "email",
				Usage:       "CERTM_EMAIL",
				EnvVars:     []string{"CERTM_EMAIL"},
				Destination: &email,
			},
			&cli.StringFlag{
				Name:        "haproxy_certs_location",
				Usage:       "CERTM_HAPROXY_CERTS_LOCATION - Directory like /etc/haproxy/certs",
				EnvVars:     []string{"CERTM_HAPROXY_CERTS_LOCATION"},
				Destination: &haproxyCertsLocation,
			},
			&cli.StringFlag{
				Name:        "letsencrypt_certs_location",
				Usage:       "CERTM_LETSENCRYPT_CERTS_LOCATION - Directory like /etc/letsencrypt/live",
				EnvVars:     []string{"CERTM_LETSENCRYPT_CERTS_LOCATION"},
				Destination: &letsEncryptCertsLocation,
			},
			&cli.StringFlag{
				Name:        "certbot_port",
				Usage:       "CERTM_CERTBOT_PORT - 9080",
				EnvVars:     []string{"CERTM_CERTBOT_PORT"},
				Destination: &certBotPort,
			},
			&cli.StringFlag{
				Name:        "s3_endpoint",
				Usage:       "CERTM_S3_ENDPOINT - S3 API ENDPOINT",
				EnvVars:     []string{"CERTM_S3_ENDPOINT"},
				Destination: &s3EndPoint,
			},
			&cli.StringFlag{
				Name:        "s3_bucket_name",
				Usage:       "CERTM_S3_BUCKET_NAME - S3 BUCKET NAME",
				EnvVars:     []string{"CERTM_S3_BUCKET_NAME"},
				Destination: &s3BucketName,
			},
			&cli.StringFlag{
				Name:        "s3_access_keys",
				Usage:       "CERTM_S3_ACCESS_KEY - S3 Access Key",
				EnvVars:     []string{"CERTM_S3_ACCESS_KEY"},
				Destination: &s3AccessKey,
			},
			&cli.StringFlag{
				Name:        "s3_private_key",
				Usage:       "CERTM_S3_PRIVATE_KEY - S3 Private Key",
				EnvVars:     []string{"CERTM_S3_PRIVATE_KEY"},
				Destination: &s3PrivateKey,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "start",
				Aliases: []string{"s"},
				Usage:   "START WORKER - EXAMPLE: start --obtaining --renew --migrate",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "obtaining",
						Usage:       "--obtaining - Mode Type Obtaining",
						Destination: &obtaining,
					},
					&cli.BoolFlag{
						Name:        "combine",
						Usage:       "--combine - Mode Type Combine",
						Destination: &combine,
					},
					&cli.BoolFlag{
						Name:        "renew",
						Usage:       "--renew - Mode Type Renew",
						Destination: &renew,
					},
					&cli.BoolFlag{
						Name:        "transfer",
						Usage:       "--transfer - Mode Type Transfer",
						Destination: &transfer,
					},
					&cli.BoolFlag{
						Name:        "puller",
						Usage:       "--puller - Mode Type Puller",
						Destination: &puller,
					},
					&cli.BoolFlag{
						Name:        "s3-ssl-disable",
						Usage:       "--s3-ssl-disable - Disabled SSL for S3 endpoint useful for puller and transfer modes",
						Destination: &s3SSLEnabled,
					},
				},
				Action: func(c *cli.Context) error {
					fmt.Println("[CERTM] LOADING WORKER AND CHECK MODE")

					start()

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
