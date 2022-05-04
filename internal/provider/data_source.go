package provider

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func dataSourceDependencyNexusRaw() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDependencyNexusRawRead,

		Schema: map[string]*schema.Schema{
			"nexus_server": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"basic_auth": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				StateFunc: func(val interface{}) string {
					return "*** sensitive ***"
				},
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				StateFunc: func(val interface{}) string {
					return "*** sensitive ***"
				},
			},

			"asset_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_md5": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_content_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_size": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type Asset struct {
	DownloadUrl  string
	Path         string
	Id           string
	Repository   string
	Format       string
	Checksum     map[string]string
	ContentType  string
	LastModified string
}

type Tag struct {
	Name  string
	Value string
}

type SearchAssetsResponse struct {
	Items             []Asset
	ContinuationToken string
}

func searchRawRepo(ctx context.Context, client *http.Client, server string, repository string, name string, authentication string) (res *Asset, err error) {

	url := fmt.Sprintf("https://%s/service/rest/v1/search/assets?name=%s&repository=%s", server, url.QueryEscape(name), repository)

	tflog.Debug(ctx, fmt.Sprintf("Sending request %s", url))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "Got response back")

	if authentication != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", authentication))

	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("HTTP request error. Response code %d", resp.StatusCode))
	}

	contentType := resp.Header.Get("Content-Type")

	if contentType != "application/json" {
		return nil, errors.New(fmt.Sprintf("Invalid Content-Type in response. Expected application/json but got %s", contentType))
	}

	var searchResponse SearchAssetsResponse

	err = json.NewDecoder(resp.Body).Decode(&searchResponse)

	if err != nil {
		return nil, err
	}

	tflog.Debug(ctx, "Got proper response back", map[string]interface{}{
		"response": searchResponse,
	})

	if len(searchResponse.Items) != 1 {
		return nil, errors.New(fmt.Sprintf("Failed to find exactly one asset (Got %d)", len(searchResponse.Items)))
	}

	return &(searchResponse.Items[0]), nil
}

func resolvePassword(ctx context.Context, password string) (string, error) {
	if strings.HasPrefix(password, "gcp_secret!") {
		secret := strings.Replace(password, "gcp_secret!", "", 1)
		tflog.Debug(ctx, "Will read password from GCP Secret")

		client, err := secretmanager.NewClient(ctx)
		if err != nil {
			return "", err
		}
		defer client.Close()

		accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
			Name: secret,
		}

		result, err := client.AccessSecretVersion(ctx, accessRequest)
		if err != nil {
			return "", err
		}
		password = string(result.GetPayload().GetData())
	}
	return password, nil
}

func dataSourceDependencyNexusRawRead(ctx context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {

	server := d.Get("nexus_server").(string)
	name := d.Get("name").(string)
	directory := d.Get("destination").(string)
	repository := "raw-trusted"

	username := d.Get("username").(string)
	password := d.Get("password").(string)
	authentication := d.Get("basic_auth").(string)

	client := &http.Client{}

	if username != "" && password != "" {
		if authentication == "" {
			tflog.Debug(ctx, "Will use username/password authentication ("+username+"/***")
			resolvedPassword, err := resolvePassword(ctx, password)
			if err != nil {
				return append(diags, diag.Errorf("Failed to resolve password: %s", err)...)
			}
			password = resolvedPassword
			authentication = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
		} else {
			return append(diags, diag.Errorf("Cannot provide authentication as well as username/password")...)
		}
	} else if username != "" || password != "" {
		return append(diags, diag.Errorf("Please provide both username/password or none")...)
	} else if authentication != "" {
		tflog.Debug(ctx, "Will use authentication with base64 token")
		_, err := base64.StdEncoding.DecodeString(authentication)
		if err != nil {
			return append(diags, diag.Errorf("Provided basic_auth is not a valid base64 string")...)
		}
	}

	asset, err := searchRawRepo(ctx, client, server, repository, name, authentication)

	if err != nil {
		return append(diags, diag.Errorf("Error retrieving asset information: %s", err)...)
	}

	expectedMd5 := asset.Checksum["md5"]
	downloadUrl := asset.DownloadUrl

	fileName := fmt.Sprintf("%s/%s", directory, name)

	d.Set("asset_id", asset.Id)
	d.Set("asset_url", downloadUrl)
	d.Set("asset_md5", expectedMd5)
	d.Set("asset_content_type", asset.ContentType)
	d.Set("asset_path", fileName)
	d.SetId(expectedMd5)

	hasher := md5.New()

	if _, err := os.Stat(fileName); err == nil {
		tflog.Debug(ctx, "File exists - checking md5 to match "+expectedMd5)
		// File exists - checking md5
		file, err := os.Open(fileName)
		if err == nil {
			tflog.Trace(ctx, "Opened file "+fileName)
			if size, err := io.Copy(hasher, file); err == nil {
				tflog.Trace(ctx, "Read contents into hasher")
				md5Value := fmt.Sprintf("%x", hasher.Sum(nil))
				tflog.Trace(ctx, "Calculated hash value: "+md5Value)
				if md5Value == expectedMd5 {
					d.Set("asset_size", fmt.Sprintf("%d", size))
					return diags
				}
			}
		}
	}

	tflog.Debug(ctx, "File does not exist - will download")
	dir := filepath.Dir(fileName)
	tflog.Trace(ctx, "Creating directory for download")
	err = os.MkdirAll(dir, 0750)

	if err != nil && !os.IsExist(err) {
		return append(diags, diag.Errorf("Failed to create directory to download asset: %s", err)...)
	}

	tflog.Trace(ctx, "Creating file")
	file, err := os.Create(fileName)
	if err != nil {
		return append(diags, diag.Errorf("Failed to create file for downloading: %s", err)...)
	}

	defer file.Close()

	tflog.Trace(ctx, "Creating request for download with url "+downloadUrl)
	req, err := http.NewRequestWithContext(ctx, "GET", downloadUrl, nil)
	if err != nil {
		return append(diags, diag.Errorf("Error creating request: %s", err)...)
	}
	if authentication != "" {
		tflog.Trace(ctx, "Adding authentication header")
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", authentication))
	}

	tflog.Trace(ctx, "Sending request to server...")
	resp, err := client.Do(req)

	if err != nil {
		return append(diags, diag.Errorf("Error making request: %s", err)...)
	}

	defer resp.Body.Close()

	tflog.Trace(ctx, "Reading response status code")
	if resp.StatusCode != 200 {
		return append(diags, diag.Errorf("HTTP request error. Response code: %d", resp.StatusCode)...)
	}

	hasher = md5.New()
	tflog.Trace(ctx, "Downloading file")
	// size, err := io.Copy(io.MultiWriter(file, hasher), resp.Body)
	size, err := io.Copy(file, resp.Body)

	if err != nil {
		return append(diags, diag.Errorf("Error downloading to file: %s", err)...)
	}

	d.Set("asset_size", fmt.Sprintf("%d", size))

	tflog.Trace(ctx, "Calculating hash")
	file, err = os.Open(fileName)
	if err == nil {
		tflog.Trace(ctx, "Opened file "+fileName)
		if size, err := io.Copy(hasher, file); err == nil {
			tflog.Trace(ctx, "Read contents into hasher")
			md5Value := fmt.Sprintf("%x", hasher.Sum(nil))
			tflog.Trace(ctx, "Calculated hash value: "+md5Value)
			if md5Value == expectedMd5 {
				d.Set("asset_size", fmt.Sprintf("%d", size))
				tflog.Trace(ctx, "Done")
				return diags
			}
		}
	}
	return append(diags, diag.Errorf("Downloaded md5 does not match expected value")...)
}
