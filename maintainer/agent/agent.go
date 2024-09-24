package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client authentication with the authorization that returns an authorization token.
// Expects environment variables: TOKEN_ENDPOINT, CLIENT_ID, and CLIENT_SECRET.
func authenticateClient() string {
	endpoint, ok := os.LookupEnv("TOKEN_ENDPOINT")
	if !ok {
		log.Fatalln("ERROR: Missing required TOKEN_ENDPOINT environment variable.")
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("ERROR: While parsing TOKEN_ENDPOINT environment variable.\n%v\n", err)
	}

	// Configure the application's registered client ID
	id, ok := os.LookupEnv("CLIENT_ID")
	if !ok {
		log.Fatalln("WARNING: No resource owner access token request endpoint at /authorized. Missing required CLIENT_ID environment variable.")
	}

        // Configure the application's registered client secret
        secret, ok := os.LookupEnv("CLIENT_SECRET")
        if !ok {
		log.Fatalln("WARNING: No resource owner access token request endpoint at /authorized. Missing required CLIENT_SECRET environment variable.")
        }

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(form.Encode()))
	req.SetBasicAuth(id, secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("ERROR: While requesting access token.\n%v\n", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ERROR: While reading response body.\n%v\n", err)
	}

	status := make(map[string]interface{})
	err = json.Unmarshal(body, &status)
	if err != nil {
		log.Fatalf("ERROR: While parsing client authentication response body.\n%v\n", err)
	}

	accessToken, ok := status["access_token"]
	if !ok {
		log.Fatalln("ERROR: Missing \"access_token\" property in client authentication response.")
	}

	token := reflect.ValueOf(accessToken)
	if token.Kind() != reflect.String {
		log.Fatalf("ERROR: The \"access_token\" property in the client authentication response is not a string. token=%+v\n", token)
	}

	return fmt.Sprintf("Bearer %s", token.String())
}

// TODO: Retry on conflicts using
//  "k8s.io/client-go/util/retry"
//  retry.RetryOnConflict(retry.DefaultRetry)
func main() {
	// The access token that will be stored in a secret.
	// TODO: Check whether the access token value has changed.
	accessToken := authenticateClient()

	// Create a Kubernetes API configuration that uses the service account provided to the pod.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("ERROR: While configuring.\n%v\n", err)
	}

	// Create a Kubernetes API client for each resource.
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("ERROR: While configuring.\n%v\n", err)
	}

	// Create secret
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "client-access-token-",
			Namespace: "default",
		},
		StringData: map[string]string{ "ACCESS_TOKEN": accessToken },
	}
	secretResource, err := clientSet.CoreV1().Secrets("default").Create(context.TODO(), &secret, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("ERROR: While creating a secret.\n%v\n", err)
	}
	log.Printf("Created secret %v\n", secretResource.ObjectMeta.Name)

	// Configure the deployment name to be updated from the environment.
	deploymentName, ok := os.LookupEnv("DEPLOYMENT_NAME")
	if !ok {
		log.Fatalln("ERROR: Missing required DEPLOYMENT_NAME environment variable.")
	}
	// Get the deployment resource.
	deployment, err := clientSet.AppsV1().Deployments("default").Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("ERROR: While getting deployment.\n%v\n", err)
	}

	// Iterate over the secrets in the deployment.
	for i, container := range deployment.Spec.Template.Spec.Containers {
		for j, source := range container.EnvFrom {
			if source.SecretRef != nil {
				// Update secrets matching the prefix of the new secret.
				if !strings.HasPrefix(source.SecretRef.LocalObjectReference.Name, "client-access-token-") {
					continue
				}
				deployment.Spec.Template.Spec.Containers[i].EnvFrom[j].SecretRef = &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretResource.ObjectMeta.Name,
					},
				}
				if _, err = clientSet.AppsV1().Deployments("default").Update(context.TODO(), deployment, metav1.UpdateOptions{}); err != nil {
					log.Fatalf("ERROR: While updating Deployment.\n%v\n", err)
				}
				log.Printf("Updated Container %v in Deployment %v from Secret %v to Secret %v\n",
					container.Name,
					deployment.ObjectMeta.Name,
					source.SecretRef.LocalObjectReference.Name,
					secretResource.ObjectMeta.Name,
				)
			}
		}
	}
}
