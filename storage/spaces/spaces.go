package spaces

import (
	"bytes"
	"fmt"

	"github.com/DMarby/picsum-photos/database"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Provider implements a digitalocean spaces based image storage
type Provider struct {
	spaces *s3.S3
	space  string
}

// New returns a new Provider instance
func New(space, region, accessKey, secretKey string) (*Provider, error) {
	spacesSession := session.New(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(fmt.Sprintf("https://%s.digitaloceanspaces.com", region)),
		Region:      aws.String("us-east-1"), // Needs to be us-east-1 for Spaces, or it'll fail
	})

	spaces := s3.New(spacesSession)

	object := s3.GetObjectInput{
		Bucket: &space,
		Key:    aws.String("/"),
	}

	_, err := spaces.GetObject(&object)
	if err != nil {
		return nil, err
	}

	return &Provider{
		spaces: spaces,
		space:  space,
	}, nil
}

// Get returns the image data for an image id
func (p *Provider) Get(id string) ([]byte, error) {
	object := s3.GetObjectInput{
		Bucket: &p.space,
		Key:    aws.String(fmt.Sprintf("%s.jpg", id)),
	}

	output, err := p.spaces.GetObject(&object)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == s3.ErrCodeNoSuchKey {
			return nil, database.ErrNotFound
		}

		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
