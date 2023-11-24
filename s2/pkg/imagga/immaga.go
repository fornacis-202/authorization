package imagga

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// Imagga
// handles the image tagging.
type Imagga struct {
	Cfg Config
}

// Response
// is for image tagging response.
type DetectionResponse struct {
	Result struct {
		Faces []struct {
			Confidence  float64 `json:"confidence"`
			Coordinates struct {
				Height int64 `json:"height"`
				Width  int64 `json:"width"`
				Xmax   int64 `json:"xmax"`
				Xmin   int64 `json:"xmin"`
				Ymax   int64 `json:"ymax"`
				Ymin   int64 `json:"ymin"`
			} `json:"coordinates"`
			FaceID string `json:"face_id"`
		} `json:"faces"`
	} `json:"result"`
}

type SimilarityResponse struct {
	Result struct {
		Score float64
	} `json:"result"`
}

type Config struct {
	ApiKey    string `koanf:"api_key"`
	ApiSecret string `koanf:"api_secret"`
}

// Process
// sends one http request to Imagga website.
func (i *Imagga) FaceDetection(imURL string) (*DetectionResponse, error) {
	// creating a new client
	client := &http.Client{}

	// creating a new get request
	address := url.QueryEscape(imURL)
	req, _ := http.NewRequest("GET", "https://api.imagga.com/v2/faces/detections?return_face_id=1&image_url="+address, nil)
	// set the auth
	req.SetBasicAuth(i.Cfg.ApiKey, i.Cfg.ApiSecret)

	log.Printf("sending request to imagga:\n\t%s\n", req.URL)

	// do the http request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// return response
	respBody, _ := ioutil.ReadAll(resp.Body)
	var response DetectionResponse

	if resp.StatusCode != http.StatusOK {
		log.Printf("imagga response: %s\n", resp.Status)
		log.Printf("\t%s\n", string(respBody))
	} else {
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, err
		}
	}

	return &response, nil
}

func (i *Imagga) FaceSimilarity(face1ID string, face2ID string) (*SimilarityResponse, error) {
	// creating a new client
	client := &http.Client{}

	// creating a new get request
	req, _ := http.NewRequest("GET", "https://api.imagga.com/v2/faces/similarity?face_id="+face1ID+"&second_face_id="+face2ID, nil)
	// set the auth
	req.SetBasicAuth(i.Cfg.ApiKey, i.Cfg.ApiSecret)

	log.Printf("sending request to imagga:\n\t%s\n", req.URL)

	// do the http request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// return response
	respBody, _ := ioutil.ReadAll(resp.Body)
	var response SimilarityResponse

	if resp.StatusCode != http.StatusOK {
		log.Printf("imagga response: %s\n", resp.Status)
		log.Printf("\t%s\n", string(respBody))
	} else {
		if err := json.Unmarshal(respBody, &response); err != nil {
			return nil, err
		}
	}

	return &response, nil
}
