package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/afero"
	"log"
	"os/exec"
	"syscall"
)

func start() {
	checkEnabledMode()
	provisioning()
}

func checkEnabledMode() {
	fmt.Println("[CERTM] MODE OBTAINING ENABLED:", obtaining)
	fmt.Println("[CERTM] MODE RENEW ENABLED:", renew)
	fmt.Println("[CERTM] MODE COMBINE ENABLED:", combine)
	fmt.Println("[CERTM] MODE TRANSFER ENABLED:", transfer)
	fmt.Println("[CERTM] MODE PULLER ENABLED:", puller)
	fmt.Println("----------------------------")
}

func provisioning() {
	fmt.Println("[CERTM] PROVISIONING")
	if obtaining {
		fmt.Println("[CERTM] PROVISIONING OBTAINING MODE")
		runObtainingMode()
	}
	if renew {
		fmt.Println("[CERTM] PROVISIONING RENEW MODE")
		if ruNow {
			runRenewMode()
		} else {
			gocron.Every(1).Day().At("2:00").Do(runRenewMode)
		}
		fmt.Println("[CERTM] RENEW MODE ENABLED")
	}
	if combine {
		fmt.Println("[CERTM] PROVISIONING COMBINE MODE")
		if ruNow {
			runCombineMode()
		} else {
			gocron.Every(1).Day().At("2:10").Do(runCombineMode)
		}
		fmt.Println("[CERTM] COMBINE MODE ENABLED")
	}
	if transfer {
		fmt.Println("[CERTM] PROVISIONING TRANSFER MODE")
		if ruNow {
			runTransferMode()
		} else {
			gocron.Every(1).Day().At("2:20").Do(runTransferMode)
		}
		fmt.Println("[CERTM] TRANSFER MODE ENABLED")
	}
	if puller {
		fmt.Println("[CERTM] PROVISIONING PULLER MODE")
		if ruNow {
			runPullerMode()
		} else {
			gocron.Every(1).Day().At("2:30").Do(runPullerMode)
		}
		fmt.Println("[CERTM] PULLER MODE ENABLED")
	}

	// remove, clear and next_run
	_, time := gocron.NextRun()
	fmt.Println(time)

	// function Start start all the pending jobs
	<-gocron.Start()

	fmt.Println("[CERTM] STARTED")
}

func generateObtainingModeOptions() string {
	var options string

	options += "--standalone --preferred-challenges http --http-01-address 127.0.0.1 --http-01-port " + certBotPort + " -d " + domainName + " --email " + email + " --agree-tos --non-interactive"

	return options
}

func runObtainingMode() {
	fmt.Println("[CERTM] OBTAINING [STARTED]")
	options := generateObtainingModeOptions()
	cmd := exec.Command("/bin/sh",
		"-c",
		"certbot certonly "+options)

	_, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("[CERTM][ERROR][OBTAINING] CAN'T RUN certbot program [1]")
		//log.Fatal(err)
	}

	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			fmt.Println("[CERTM][ERROR][OBTAINING] CAN'T RUN certbot program [2]")
			//log.Fatal(err)
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			fmt.Println("[CERTM][ERROR][OBTAINING] CAN'T RUN certbot program [3]")
			fmt.Printf("[CERTM][ERROR][OBTAINING] Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("[CERTM][ERROR][OBTAINING] Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		fmt.Println("[CERTM][OBTAINING] [SUCCESS]")
	}
}

func generateRenewModeOptions() string {
	var options string

	options += "--standalone --preferred-challenges http --http-01-address 127.0.0.1 --http-01-port " + certBotPort + " --quiet"

	return options
}

func runRenewMode() {
	fmt.Println("[CERTM][RENEW] [STARTED]")
	options := generateRenewModeOptions()
	cmd := exec.Command("/bin/sh",
		"-c",
		"certbot renew "+options)

	_, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("[CERTM][ERROR][RENEW] CAN'T RUN certbot program [1]")
		//log.Fatal(err)
	}

	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		if err != nil {
			fmt.Println("[CERTM][ERROR][RENEW] CAN'T RUN certbot program [2]")
			//log.Fatal(err)
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			fmt.Println("[CERTM][ERROR][RENEW] CAN'T RUN certbot program [3]")
			fmt.Printf("[CERTM][ERROR][RENEW] Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		}
	} else {
		// Success
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		fmt.Printf("[CERTM][ERROR][RENEW] Output: %s\n", []byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		fmt.Println("[CERTM][RENEW] [SUCCESS]")
	}
}

func runCombineMode() {
	fmt.Println("[CERTM] COMBINE [STARTED]")
	var AppFs = afero.NewOsFs()

	afs := &afero.Afero{Fs: AppFs}

	haproxyCertsDirCreated, err := afs.DirExists(haproxyCertsLocation)
	if err != nil || !haproxyCertsDirCreated {
		fmt.Println("[CERTM][COMBINE] Make sure you have HAProxy dirs for CERT!: " + haproxyCertsLocation)
		log.Fatal(err)
	}

	letsEncryptCertsDirCreated, err := afs.DirExists(letsEncryptCertsLocation)
	if err != nil || !letsEncryptCertsDirCreated {
		fmt.Println("[CERTM][COMBINE] Make sure you have Let's Encrypt directory: " + letsEncryptCertsLocation)
		log.Fatal(err)
	}

	if letsEncryptCertsDirCreated && haproxyCertsDirCreated {
		letsEncryptDirItems, _ := afs.ReadDir(letsEncryptCertsLocation)
		for _, item := range letsEncryptDirItems {
			fmt.Println("[CERTM][COMBINE][DOMAINS] " + item.Name())
			if item.IsDir() {
				subItems, _ := afs.ReadDir(letsEncryptCertsLocation + "/" + item.Name())
				for _, subItem := range subItems {
					fmt.Println("[CERTM][COMBINE] FIND FILE: " + subItem.Name())
				}

				fullChainCreated, _ := afs.IsEmpty(letsEncryptCertsLocation + "/fullchain.pem")
				privkeyCreated, _ := afs.IsEmpty(letsEncryptCertsLocation + "/privkey.pem")

				if !fullChainCreated && !privkeyCreated {
					fmt.Println("[CERTM][COMBINE] SUCCESS FIND FILES: fullchain.pem,privkey.pem")
					fmt.Println("[CERTM][COMBINE] COMBINING")
					// handle file there
					subItemPath := letsEncryptCertsLocation + "/" + item.Name()

					fmt.Println("[CERTM][COMBINE] INIT FILE LOCATION: " + subItemPath + "/fullchain.pem")
					fullChainPemFile, _ := afs.ReadFile(subItemPath + "/fullchain.pem")
					fmt.Println("[CERTM][COMBINE] INIT FILE LOCATION: " + subItemPath + "/privkey.pem")
					privKeyPemFile, _ := afs.ReadFile(subItemPath + "/privkey.pem")

					var buf bytes.Buffer
					buf.Write(fullChainPemFile)
					buf.Write(privKeyPemFile)

					var pemFile []byte
					_, err = buf.Write(pemFile)
					var haproxyPemFileLocation string = haproxyCertsLocation + "/" + item.Name() + ".pem"

					afs.WriteFile(haproxyPemFileLocation, pemFile, 644)
					fmt.Println("[CERTM][COMBINE] HAProxy PEM created: " + haproxyPemFileLocation)
				} else {
					fmt.Println("[CERTM][COMBINE] fullchain.pem or privkey not created for domain: " + item.Name())
				}

			} else {
				// handle file there
				fmt.Println(item.Name())
			}
		}
	}
	fmt.Println("[CERTM] COMBINE [SUCCESS]")
}

func runTransferMode() {
	fmt.Println("[CERTM][TRANSFER] [STARTED]" + s3EndPoint)

	ctx := context.Background()
	endpoint := s3EndPoint
	accessKeyID := s3AccessKey
	secretAccessKey := s3PrivateKey
	useSSL := s3SSLEnabled

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Make a new bucket called certm.
	bucketName := "certm"
	location := "us-east-1"

	err = minioClient.MakeBucket(ctx, s3BucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("[CERTM][TRANSFER] We already own bucket: %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("[CERTM][TRANSFER] Successfully created [SUCCESS]%s\n", bucketName)
	}

	var AppFs = afero.NewOsFs()

	afs := &afero.Afero{Fs: AppFs}

	haproxyCertsDirCreated, err := afs.DirExists(haproxyCertsLocation)
	if err != nil || !haproxyCertsDirCreated {
		fmt.Println("[CERTM][TRANSFER] Make sure you have HAProxy dirs for CERT!: " + haproxyCertsLocation)
		log.Fatal(err)
	}

	if haproxyCertsDirCreated {
		haproxyDirItems, _ := afs.ReadDir(haproxyCertsLocation)
		for _, item := range haproxyDirItems {
			fmt.Println("[CERTM][TRANSFER][DOMAINS] " + item.Name())
			if !item.IsDir() {
				fmt.Println("[CERTM][TRANSFER] FIND FILE: " + item.Name())

				fmt.Println("[CERTM][COMBINE] SUCCESS FIND FILES: fullchain.pem,privkey.pem")
				fmt.Println("[CERTM][COMBINE] COMBINING")
				// handle file there
				itemPath := haproxyCertsLocation + "/" + item.Name()

				fmt.Println("[CERTM][TRANSFER] HAProxy PEM found: " + itemPath)

				// Upload the zip file
				objectName := item.Name()
				filePath := itemPath
				//contentType := "application/zip"

				// Upload the zip file with FPutObject
				info, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
				if err != nil {
					log.Fatalln(err)
				}

				log.Printf("[CERTM][TRANSFER] Successfully uploaded %s of size %d\n", objectName, info.Size)
			} else {
				fmt.Println("[CERTM][WARNING][TRANSFER] Look's like NO FILE IN HAPROXY SSL Dir! Try to hold only pem there!: " + item.Name())
			}
		}
	}
}

func runPullerMode() {
	fmt.Println("[CERTM][PULLER] PULLER [STARTED]")

	var AppFs = afero.NewOsFs()

	afs := &afero.Afero{Fs: AppFs}

	haproxyCertsDirCreated, err := afs.DirExists(haproxyCertsLocation)
	if err != nil || !haproxyCertsDirCreated {
		fmt.Println("[CERTM][TRANSFER] Make sure you have HAProxy dirs for CERT!: " + haproxyCertsLocation)
		log.Fatal(err)
	}

	if haproxyCertsDirCreated {
		s3Client, err := minio.New(s3EndPoint, &minio.Options{
			Creds:  credentials.NewStaticV4(s3AccessKey, s3PrivateKey, ""),
			Secure: s3SSLEnabled,
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		opts := minio.ListObjectsOptions{
			UseV1:     true,
			Recursive: true,
		}

		// List all objects from a bucket.
		for object := range s3Client.ListObjects(context.Background(), s3BucketName, opts) {

			if object.Err != nil {
				fmt.Println(object.Err)
				return
			}
			fmt.Println("[CERTM][PULLER] FOUND FILE: " + object.Key)

			err = s3Client.FGetObject(context.Background(), s3BucketName, object.Key, haproxyCertsLocation+"/"+object.Key, minio.GetObjectOptions{})
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("[CERTM][PULLER] DOWNLOAD FILE: " + haproxyCertsLocation + "/" + object.Key)
		}
	}
	fmt.Println("[CERTM][PULLER] [SUCCESS]")
}
