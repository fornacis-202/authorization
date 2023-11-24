package main

import (
	"ImageProccessor/model"
	"ImageProccessor/pkg/conf"
	"ImageProccessor/pkg/imagga"
	"ImageProccessor/pkg/mailgun"
	"ImageProccessor/pkg/mqtt"

	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3sdk "github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
)

var (
	mq             *mqtt.MQTT
	db             *gorm.DB
	s3Session      *session.Session
	imaggaSession  *imagga.Imagga
	mailgunSession *mailgun.Mailgun
	cfg            conf.Config
)

func handle() {

	// creating a consumer for rabbitMQ
	events, err := mq.Channel.Consume(
		mq.Queue,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("processor started ...")

	// listen over rabbitMQ events
	for event := range events {
		// get id from rabbitMQ
		id := string(event.Body)
		log.Printf("receive id:\n\t%s\n", id)

		var person model.Person
		db.Where("ID = ?", id).First(&person)
		if fmt.Sprint(person.ID) != id {
			log.Println("No Person in database with this National ID")
			continue
		}

		svc := s3sdk.New(s3Session, &aws.Config{
			Region:   aws.String(cfg.S3.Region),
			Endpoint: aws.String(cfg.S3.Endpoint),
		})

		req, _ := svc.GetObjectRequest(&s3sdk.GetObjectInput{
			Bucket: aws.String(cfg.S3.Bucket),
			Key:    aws.String(person.Image1),
		})

		im1UrlStr, err := req.Presign(15 * time.Minute)
		if err != nil {
			log.Println(err)
			continue
		}

		req, _ = svc.GetObjectRequest(&s3sdk.GetObjectInput{
			Bucket: aws.String(cfg.S3.Bucket),
			Key:    aws.String(person.Image2),
		})
		im2UrlStr, err := req.Presign(15 * time.Minute)
		if err != nil {
			log.Println(err)
			continue
		}
		res1, err := imaggaSession.FaceDetection(im1UrlStr)
		if err != nil {
			log.Println(err)
			continue
		}

		res2, err := imaggaSession.FaceDetection(im2UrlStr)
		if err != nil {
			log.Println(err)
			continue
		}
		var isFaceDetected bool
		if len(res1.Result.Faces) > 0 && len(res2.Result.Faces) > 0 {
			log.Println("face detected.")
			isFaceDetected = true
		} else {
			log.Println("no face detected.")
		}
		var isSimilar bool
		if isFaceDetected {
			fmt.Println("f1   ", res1.Result.Faces[0].FaceID, res2.Result.Faces[0].FaceID)
			res, err := imaggaSession.FaceSimilarity(res1.Result.Faces[0].FaceID, res2.Result.Faces[0].FaceID)
			if err != nil {
				log.Println(err)
				continue
			}
			if res.Result.Score >= 80 {
				log.Println("similarity detected.")
				isSimilar = true
			} else {
				log.Println("similarity not detected.")
				isSimilar = false
			}
		}
		if isSimilar {
			person.State = "Authorized"
		} else {
			person.State = "Unauthorized"
		}
		db.Save(&person)
		msg := fmt.Sprintf(
			"Dear '%s', your authentication request status has changed to: '%s'",
			strings.TrimSpace(person.Lastname),
			person.State,
		)
		if err := mailgunSession.Send(msg, "Authentication status", person.Email); err != nil {
			log.Println(err)
			log.Println("email did not send.")
		} else {
			log.Printf("email sent {id: %s}\n", id)
		}

	}
}

func main() {
	cfg = conf.Load()
	var err error
	db, err = model.OpenDB(cfg)
	if err != nil {
		panic(err)
	}
	mq, err = mqtt.NewConnection(cfg.MQTT)
	if err != nil {
		panic(err)
	}
	s3Session, err = session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(cfg.S3.AccessKey, cfg.S3.SecretKey, ""),
		Region:      aws.String("default"),
		Endpoint:    aws.String(cfg.S3.Endpoint),
	})
	if err != nil {
		panic(err)
	}
	imaggaSession = &imagga.Imagga{Cfg: cfg.Imagga}
	mailgunSession = mailgun.NewConnection(cfg.Mailgun)

	handle()

}
