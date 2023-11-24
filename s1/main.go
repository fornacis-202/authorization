package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"requestHandler/model"
	"requestHandler/pkg/conf"
	"requestHandler/pkg/mqtt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/labstack/echo/v4"
	"github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

var (
	db        *gorm.DB
	mq        *mqtt.MQTT
	s3Session *session.Session
	cfg       conf.Config
)

func statusHandler(c echo.Context) error {
	ip := c.RealIP()
	nationalID := c.Param("Nid")
	var person = model.Person{}
	db.Where("National_ID = ?", nationalID).First(&person)
	if person.NationalID != nationalID {
		return c.JSON(http.StatusNotFound, "No request was submitted with this National ID")
	}
	if person.IP != ip {
		return c.JSON(http.StatusUnauthorized, "Your registeration ip differs from your currnent ip.")
	}
	return c.JSON(http.StatusOK, "Your request status is: "+person.State)

}

func authenticateHandler(c echo.Context) error {

	person := model.Person{}
	person.NationalID = c.FormValue("nationalID")
	person.Email = c.FormValue("email")
	person.Lastname = c.FormValue("lastname")
	person.IP = c.RealIP()
	person.State = "Pending"
	fmt.Println(person.ID)
	result := db.Create(&person)
	if result.Error != nil {
		log.Println(result.Error)
	}

	f1, err := c.FormFile("image1")
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error loading image")

	}
	f2, err := c.FormFile("image2")
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error loading image")

	}

	im1, err := f1.Open()
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error openning image")

	}
	im2, err := f2.Open()
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error openning image")

	}

	uid1 := fmt.Sprint(person.ID) + "im1.jpg"
	uid2 := fmt.Sprint(person.ID) + "im2.jpg"
	// creating a new uploader
	uploader := s3manager.NewUploader(s3Session)
	// upload image into s3 database
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(uid1),
		Body:   im1,
	})
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error connecting to s3 database")

	}

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(cfg.S3.Bucket),
		Key:    aws.String(uid2),
		Body:   im2,
	})
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error connecting to s3 database")

	}
	person.Image1 = uid1
	person.Image2 = uid2
	db.Save(&person)
	// rabbit mq code

	// publish id over mqtt

	err = mq.Channel.PublishWithContext(
		context.Background(),
		"",
		mq.Queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte(fmt.Sprint(person.ID)),
		},
	)
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error sending to rabbitmq database")

	}

	return c.JSON(http.StatusOK, "Your authentication request has been submited.")

}

func main() {
	var err error
	cfg = conf.Load()
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
	e := echo.New()
	e.GET("/status/:Nid", statusHandler)
	e.POST("/authenticate", authenticateHandler)
	e.Start("0.0.0.0:8080")
}
