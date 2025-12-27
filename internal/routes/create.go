package routes

import (
	"log"
	"net/http"
	"tiny-bitly/internal/config"
	"tiny-bitly/internal/dao"
	"tiny-bitly/internal/routes/utils"
	"tiny-bitly/internal/service/create_service"
)

type CreateUrlRequest struct {
	URL string `json:"url"`

	// A specific user-provided alias to use in the short URL.
	// If not provided, a random short code will be created.
	Alias *string `json:"alias"`
}

type CreateUrlResponse struct {
	ShortURL string `json:"shortUrl"`
}

// Handles a POST request to create a short URL for a provided URL.
// - 200 OK with a CreateUrlResponse on success
// - 400 Bad Request if the original URL is an invalid URL
// - 400 Bad Request if the original URL is longer than 1000 chars
// - 500 System Error if anything else fails
func HandlePostURL(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Attempt to read the JSON request body.
	request, err := utils.ReadRequestJson[CreateUrlRequest](r)
	if err != nil {
		http.Error(w, "Malformatted request JSON", http.StatusBadRequest)
		return
	}

	maxURLLength := config.GetIntEnvOrDefault("MAX_URL_LENGTH", 6)
	if len(request.URL) > maxURLLength {
		log.Print("Bad request: original URL is longer than 1000 chars")
		http.Error(w, "URL must be no longer than 1000 chars", http.StatusBadRequest)
		return
	}

	// Log the inbound request.
	log.Printf("Request: URL=%s Alias=%s\n", request.URL, request.Alias)

	// Create a DAO.
	dao := dao.GetDAOOfType(dao.DAOTypeMemory)
	if dao == nil {
		log.Println("Internal server error: failed to get DAO")
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	// Create the short URL.
	shortURL, err := create_service.CreateShortURL(*dao, request.URL, request.Alias)
	if err != nil {
		log.Println("Internal server error:", err.Error())
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}

	// Send the JSON response.
	err = utils.WriteResponseJson(w, CreateUrlResponse{
		ShortURL: *shortURL,
	})
	if err != nil {
		http.Error(w, "Failed to create URL", http.StatusInternalServerError)
		return
	}
}
