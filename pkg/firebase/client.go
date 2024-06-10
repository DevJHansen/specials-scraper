package firebase

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"github.com/DevJHansen/specials/internal"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func sanitizeFileName(fileName string) string {
	// Replace invalid characters with underscores
	re := regexp.MustCompile(`[^\w\-]`)
	return re.ReplaceAllString(fileName, "_")
}

func ensurePDFExtension(fileName string) string {
	if filepath.Ext(fileName) == "" {
		return fileName + ".pdf"
	}
	return fileName
}

func NewFirebaseApp(ctx context.Context) (*firebase.App, error) {
	opt := option.WithCredentialsFile("../firebase-sa.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
		return nil, err
	}
	return app, nil
}

func GetSpecials(app *firebase.App, ctx context.Context) ([]internal.Special, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	specials := make([]internal.Special, 0)
	iter := client.Collection(internal.SpecialsFirestoreCollection).Documents(ctx)

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		var special internal.Special
		if err := doc.DataTo(&special); err != nil {
			log.Printf("Error decoding document: %v", err)
			continue
		}

		specials = append(specials, special)
	}

	return specials, nil
}

func GetSpecialByField(app *firebase.App, ctx context.Context, field string, value string) (internal.Special, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		return internal.Special{}, err
	}
	defer client.Close()

	specials := make([]internal.Special, 0)
	iter := client.Collection(internal.SpecialsFirestoreCollection).Where(field, "==", value).Limit(1).Documents(ctx)

	for {
		doc, err := iter.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			return internal.Special{}, err
		}

		var special internal.Special
		if err := doc.DataTo(&special); err != nil {
			log.Printf("Error decoding document: %v", err)
			continue
		}

		specials = append(specials, special)
	}

	if len(specials) == 0 {
		return internal.Special{}, nil
	}

	return specials[0], nil
}

func AddSpecial(app *firebase.App, ctx context.Context, special internal.Special) error {
	client, err := app.Firestore(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, _, err = client.Collection(internal.SpecialsFirestoreCollection).Add(ctx, special)
	if err != nil {
		return err
	}

	return nil
}

func UploadFileToStorage(app *firebase.App, ctx context.Context, filename string, filepath string) (string, error) {
	client, err := app.Storage(ctx)
	if err != nil {
		return "", err
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return "", err
	}

	obj := bucket.Object(filename)
	w := obj.NewWriter(ctx)
	defer w.Close()

	if err := w.Close(); err != nil {
		return "", err
	}

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", err
	}

	return attrs.MediaLink, nil
}

func UploadFileToFirebase(app *firebase.App, ctx context.Context, fileURL, scrapingID string) (string, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	sanitizedFileName := sanitizeFileName(scrapingID)
	localFilePath := ensurePDFExtension(fmt.Sprintf("/tmp/%s", sanitizedFileName))
	out, err := os.Create(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy file to local storage: %w", err)
	}

	bucketName := "specials-c4acf.appspot.com"
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("../firebase-sa.json"))
	if err != nil {
		return "", fmt.Errorf("failed to create storage client: %w", err)
	}

	bucket := client.Bucket(bucketName)
	object := bucket.Object(filepath.Base(localFilePath))

	wc := object.NewWriter(ctx)
	localFile, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	if _, err := io.Copy(wc, localFile); err != nil {
		return "", fmt.Errorf("failed to upload file to Firebase Storage: %w", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Make the object publicly accessible
	if err := object.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		return "", fmt.Errorf("failed to set object ACL: %w", err)
	}

	downloadURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, filepath.Base(localFilePath))
	return downloadURL, nil
}
