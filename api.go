package main

import (
        "encoding/json"
        "log"
        "net/http"
        "database/sql"
	"github.com/satori/go.uuid"
        _ "github.com/lib/pq"

        "fmt"
)

var db *sql.DB

type Product struct {
        Id int
        Name string
        Description string
        Balance int
        Discount int
        Category int
}

type User struct {
        Id int
        Email string
        Password string
        Role int
        Status bool
}

type Products struct {
        Products []Product
}

type Users struct {
        Users []User
}

func main () {
        var err error

        db, err = sql.Open("postgres", "host=127.0.0.1 user=api password=12345678 dbname=api sslmode=disable")
        if err != nil {
                panic(err)
        }

        defer db.Close()

        fmt.Println("starting server...")

        http.HandleFunc("/v1/products/add", addProduct)
        http.HandleFunc("/v1/products/", getProducts)
        http.HandleFunc("/v1/user/add", addUser)
        http.HandleFunc("/v1/user/delete", delUser)
        http.HandleFunc("/v1/user/get_token", getToken)
        http.HandleFunc("/v1/user/get_users", getUsers)
	http.HandleFunc("/v1/test/smoke", smokeTest)
        log.Fatal(http.ListenAndServe(":8080", nil))
}

func smokeTest (w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
                fmt.Fprintln(w, "{\"Api for Rebrainme.store v1.0 by @yourmedv")
                }
}


func getUsers (w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
                x_token := r.Header.Get("X-API-TOKEN")
                query := fmt.Sprintf("select u.role from sessions s left join users u on u.id = s.user_id where s.token = '%s' and ((s.added + interval '1h') > now()) and u.role = 1", x_token)
                row := db.QueryRow(query)
                var id int
                err := row.Scan(&id)
                if err != nil {
                        http.Error(w, "Token not found", 403)
                } else {
			w_array := Users{}

			rows, err := db.Query("SELECT id, email, role, status from users")

			if err != nil {
                        	panic(err)
                	}

                	for rows.Next() {
				w_user := User{}
                        	err = rows.Scan(&w_user.Id,&w_user.Email,&w_user.Role,&w_user.Status)
                        	if err != nil {
                                	panic(err)
                        	}
                        	w_array.Users = append(w_array.Users, w_user)

                	}

                	json.NewEncoder(w).Encode(w_array)
        	}
	}
}


func delUser (w http.ResponseWriter, r *http.Request) {
        if r.Method != "DELETE" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
		x_token := r.Header.Get("X-API-TOKEN")
		query := fmt.Sprintf("select u.role from sessions s left join users u on u.id = s.user_id where s.token = '%s' and ((s.added + interval '1h') > now()) and u.role = 1", x_token)
		row := db.QueryRow(query)
		var id int
		err := row.Scan(&id)
		if err != nil {
			http.Error(w, "Token not found", 403)
		} else {
			decoder := json.NewDecoder(r.Body)
			var g_user User

			err:= decoder.Decode(&g_user)
			if err != nil {
				panic(err)
			}

			query := fmt.Sprintf("DELETE FROM users where id = %d", g_user.Id)
			_, err = db.Exec(query)
			if err != nil {
				http.Error(w, "Internal Error", 500)
			} else {
				fmt.Fprintf(w, "User was successfully removed")
			}
		}
	}
}

func getToken (w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
                decoder := json.NewDecoder(r.Body)
                var g_user User

                err := decoder.Decode(&g_user)
                if err != nil {
                        panic(err)
                }

                query := fmt.Sprintf("select id from users where email = '%s' and password = '%s' and status", g_user.Email, g_user.Password)

                fmt.Println("# INSERT QUERY: %s", query)

                row := db.QueryRow(query)

		var id int
                err = row.Scan(&id)
                if err != nil {
			http.Error(w, "User Not Found", 403)
                } else {
			query := fmt.Sprintf("select token from sessions where user_id = %d and ((added + interval '1h') > now())", id)
			fmt.Println("# select from sessions where user_id = %d and ((added + interval '1h') > now())", id)
			row := db.QueryRow(query)

			var token string
			err = row.Scan(&token)
			if err != nil {
				token := uuid.Must(uuid.NewV4())
				fmt.Println("# GOT TOKEN: %s", token)
				query := fmt.Sprintf("INSERT INTO sessions (user_id, token) VALUES (%d, '%s')", id, token)
				_, err := db.Exec(query)
				if err != nil {
					http.Error(w, "Internal Error", 500)
				} else {
					fmt.Fprintf(w, "{\"token\": %s}", token)
				}
			} else {
				fmt.Fprintf(w, "{\"token\":\"%s\"}", token)
			}

		}
        }
}

func addUser (w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
                decoder := json.NewDecoder(r.Body)
                var g_user User

                err := decoder.Decode(&g_user)
                if err != nil {
                        panic(err)
                }

                query := fmt.Sprintf("INSERT INTO users (email, password, role, status) VALUES ('%s', '%s', %d, %t) RETURNING id", g_user.Email, g_user.Password, g_user.Role, g_user.Status)

                fmt.Println("# INSERT QUERY: %s", query)

                rows, err := db.Query(query)
                if err != nil {
                        panic(err)
                }

                for rows.Next() {
                        var id int
                        err = rows.Scan(&id)
                        if err != nil {
                                panic(err)
                        }
                        fmt.Fprintf(w, "{\"id\":%d}", id)
                }
        }
}


func addProduct (w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
                decoder := json.NewDecoder(r.Body)
                var g_product Product

                err := decoder.Decode(&g_product)
                if err != nil {
                        panic(err)
                }

                query := fmt.Sprintf("INSERT INTO products(name, description, balance, discount, category) VALUES('%s', '%s', %d, %d, %d) RETURNING id", g_product.Name, g_product.Description, g_product.Balance, g_product.Discount, g_product.Category)

                fmt.Println("# INSERT QUERY: %s", query)

                rows, err := db.Query(query)
                if err != nil {
                        panic(err)
                }

                for rows.Next() {
                        var id int
                        err = rows.Scan(&id)
                        if err != nil {
                                panic(err)
                        }
                        fmt.Fprintf(w, "{\"id\":%d}", id)
                }

        }
}

func getProducts(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
                http.Error(w, "Method Not Allowed", 405)
        } else {
                w_array := Products{}

                fmt.Println("# Querying")
                rows, err := db.Query("SELECT id,name,description,discount,category from products")
                if err != nil {
                        panic(err)
                }

                for rows.Next() {
                        w_product := Product{}

                        err = rows.Scan(&w_product.Id,&w_product.Name,&w_product.Description,&w_product.Discount,&w_product.Category)
                        if err != nil {
                                panic(err)
                        }
                        w_array.Products = append(w_array.Products, w_product)

                }

                json.NewEncoder(w).Encode(w_array)
        }
}
