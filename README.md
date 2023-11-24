# authorization
In this project we are going to create a Authorization By Image Service. We are building two Backend services for getting user information and processing the next steps. <br>

# s1
In the first API (s1) we are going to get all of our clients information including ID, name, two images, and eamil. <br>
Then we need to store User email, name, ID, and IP in a MySQL cluster and the images in a S3 database. After that we need to publish this information on a RabbitMQ Cluster for the second API (s2).

# s2
In this API (s2) we need to process the images. We need to subscribe over RabbitMQ Cluster to receive events from first API. After that we are going to get images from S3 cluster and then it will send that image to Image Similarity service. If the images are similar, the user will be authorized. then, it will send an email to client via Mailgun to inform the results.

## Requirements 

- Cloud based service
- DBaaS (Database as a service)
- Amazon S3 database for Object storage
- RabbitMQ service
- Image Similarity service
- Mail service
