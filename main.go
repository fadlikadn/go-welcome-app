package main
/**
We import 4 important libraries
1. "net/http" to access the core go http functionality
2. "fmt" for formatting our text
3. "html/templates" a library that allows us to interact with our html file
4. "time" - a library for working with date and time.
 */


import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

// Create a struct that holds information to be displayed in out HTML file
type Welcome struct {
	Name string
	Time string
}

/**
	Controller to handle Welcome page
 */
func welcomeController(w http.ResponseWriter, r *http.Request) {
	welcome := Welcome{"Anonymous", time.Now().Format(time.Stamp)}
	templates := template.Must(template.ParseFiles("templates/welcome-template.html"))

	if name := r.FormValue("name"); name !="" {
		welcome.Name = name;
	}
	// If errors show an internal server error message
	// I also pass the welcome struct to the welcome-templates.html file.

	if err := templates.ExecuteTemplate(w, "welcome-template.html", welcome); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

	}
}

func balancesController(w http.ResponseWriter, r *http.Request) {
	// @section1: reading the response body
	type Balance struct{
		Address string
		Contract string
	}

	var b Balance

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&b)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	fmt.Println("Post request received. It's payload ->")
	fmt.Println("Address:" + b.Address)
	fmt.Println("Contract:" + b.Contract)

	// @section2: creating a post request to heroku api
	x := new(bytes.Buffer)
	json.NewEncoder(x).Encode(b)
	res, _ := http.Post("https://web3-challenge-heroku.herokuapp.com/balances", "application/json; charset=utf-8", x)

	type BalanceAPIResponse struct{
		Balance string
	}

	var bar BalanceAPIResponse

	json.NewDecoder(res.Body).Decode(&bar)

	fmt.Println("Post request to heroku working. It's a response payload ->")
	fmt.Println("Balance:" + bar.Balance)

	// @section3 : sending bac to the user a JSON payload
	json.NewEncoder(w).Encode(bar)
}

// Go application entrypoint
func main() {

	http.Handle("/static/", // final url can be anything
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	// Routing
	http.HandleFunc("/", welcomeController)
	http.HandleFunc("/balances", balancesController)

	// Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	// Print any errors from starting the webserver using fmt
	fmt.Println("Listening")
	fmt.Println(http.ListenAndServe(":8080", nil));
}


