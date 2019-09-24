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
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Create a struct that holds information to be displayed in out HTML file
type Welcome struct {
	Name string
	Time string
}

type Customer struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

var customers []Customer
var db *sql.DB
var err error

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
	//db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/service_support_db")
	db ,err = sql.Open("mysql", "root:@/service_support_db")

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()
	/**
		With Gorilla Mux Router
	 */
	router := mux.NewRouter()

	// Mock Data
	/*customers = append(customers, Customer{ID: "1", Name: "Fadlika Dita Nurjanto", Phone: "0851672719", Email: "fadlikadn@gmail.com" })
	customers = append(customers, Customer{ID: "2", Name: "Fauzan Ibnu Prihadiyono", Phone: "0851672789", Email: "fauzanibnup@gmail.com" })
	customers = append(customers, Customer{ID: "3", Name: "Fauzi Triagung Wiguna", Phone: "0812671872", Email: "fauzitri@gmail.com" })*/

	router.HandleFunc("/", welcomeController).Methods("GET")
	router.HandleFunc("/balances", balancesController).Methods("POST")

	router.HandleFunc("/customers", getCustomers).Methods("GET")
	router.HandleFunc("/customers", createCustomers).Methods("POST")
	router.HandleFunc("/customers/{id}", getCustomer).Methods("GET")
	router.HandleFunc("/customers/{id}", updateCustomer).Methods("PUT")
	router.HandleFunc("/customers/{id}", deleteCustomer).Methods("DELETE")

	http.ListenAndServe(":8080", router)


	/**
		Without Gorilla Mux Router
	 */
	/*http.Handle("/static/", // final url can be anything
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	// Routing
	http.HandleFunc("/", welcomeController)
	http.HandleFunc("/balances", balancesController)

	fmt.Println("Listening")
	fmt.Println(http.ListenAndServe(":8080", nil));*/
}

func deleteCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	// Using memory
	/*for index, item := range customers {
		if item.ID == params["id"] {
			customers = append(customers[:index], customers[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(customers)*/

	// Using database
	stmt, err := db.Prepare("DELETE FROM customers WHERE id  = ?")
	if err != nil {
		panic(err.Error())
	}

	_, err = stmt.Exec(params["id"])
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(w, "Post with ID = %s was deleted", params["id"])
}

func updateCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	// Using memory
	/*for index, item := range customers {
		if item.ID == params["id"] {
			customers = append(customers[:index], customers[index+1:]...)

			var customer Customer
			_ = json.NewDecoder(r.Body).Decode(&customer)
			customer.ID = params["id"]
			customers = append(customers, customer)
			json.NewEncoder(w).Encode(&customer)

			return
		}
	}
	json.NewEncoder(w).Encode(customers)*/

	// Using database
	stmt, err := db.Prepare("UPDATE customers SET name = ?, phone = ?, email = ? WHERE id = ?")
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	newName := keyVal["name"]
	newPhone := keyVal["phone"]
	newEmail := keyVal["email"]

	_, err = stmt.Exec(newName, newPhone, newEmail, params["id"])
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(w, "Customer with ID = %s was updated", params["id"])
}

func getCustomer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	// Using memory
	/*for _, item := range customers {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Customer{})*/

	// Using database
	result, err := db.Query("SELECT id, name, phone, email FROM customers where id = ?", params["id"])
	if err != nil {
		panic(err.Error())
	}

	defer result.Close()

	var customer Customer

	for result.Next() {
		err := result.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Email)
		if err != nil {
			panic(err.Error())
		}
	}

	json.NewEncoder(w).Encode(customer)
}

func createCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Using Memory
	/*var customer Customer
	_ = json.NewDecoder(r.Body).Decode(&customer)
	customer.ID = strconv.Itoa(rand.Intn(1000000))
	customers = append(customers, customer)
	json.NewEncoder(w).Encode(&customer)*/

	// Using MySQL
	stmt, err := db.Prepare("INSERT INTO customers(id, name, phone, email) VALUES(?, ?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}

	body, err:= ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err.Error())
	}

	keyVal := make(map[string]string)
	json.Unmarshal(body, &keyVal)
	name := keyVal["name"]
	phone := keyVal["phone"]
	email := keyVal["email"]
	id := strconv.Itoa(rand.Intn(1000000))

	_, err = stmt.Exec(id, name, phone, email)
	if err != nil {
		panic(err.Error())
	}

	fmt.Fprintf(w, "New customer was created")
}

func getCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// GET FROM MySQL
	var customers []Customer

	result, err := db.Query("SELECT id, name, phone, email from customers")
	if err != nil {
		panic(err.Error())
	}

	defer result.Close()

	for result.Next() {
		var customer Customer
		err := result.Scan(&customer.ID, &customer.Name, &customer.Phone, &customer.Email)
		if err != nil {
			//panic(err.Error())
			panic(err.Error())
		}
		customers = append(customers, customer)
	}

	json.NewEncoder(w).Encode(customers)
}


